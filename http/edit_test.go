package http

import (
	"context"
	"encoding/json"
	"fmt"
	"main/common"
	"main/dataaccess"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"text/template"

	"github.com/google/go-cmp/cmp"
)

func TestRedirectHandler_ValidRedirect(t *testing.T) {
	testURL := "http://localhost:8090/edit/https:/www.reddit.com/r/shortcuts/comments/bkqsv7/what_website_do_you_edit_shortcuts_on/emj1f0l/"
	body := strings.NewReader("")

	r, _ := http.NewRequest("GET", testURL, body)

	resp := httptest.NewRecorder()

	RedirectHandler(resp, r)

	want := "/r/shortcuts-bkqsv7-emj1f0l/submit/"
	wantCode := http.StatusFound

	if !cmp.Equal(wantCode, resp.Code) {
		t.Fatalf(`RedirectHandler("%v, %v"), want match for %d, nil`, resp, r, wantCode)
	}

	if !cmp.Equal(want, resp.Result().Header.Get("Location")) {
		t.Fatalf(`RedirectHandler("%v, %v"), want match for %s, nil`, resp, r, want)
	}
}

func TestRedirectHandler_MalformedUrl(t *testing.T) {
	testURL := "http://localhost:8090/123"
	body := strings.NewReader("")

	r, _ := http.NewRequest("GET", testURL, body)

	resp := httptest.NewRecorder()

	RedirectHandler(resp, r)

	want := "/"
	wantCode := http.StatusSeeOther

	if !cmp.Equal(wantCode, resp.Code) {
		t.Fatalf(`RedirectHandler("%v, %v"), want match for %d, nil`, resp, r, wantCode)
	}

	if !cmp.Equal(want, resp.Result().Header.Get("Location")) {
		t.Fatalf(`RedirectHandler("%v, %v"), want match for %s, nil`, resp, r, want)
	}
}

func TestRedirectHandler_InvalidUrl(t *testing.T) {
	testURL := "http://localhost:8090/edit/"
	body := strings.NewReader("")

	r, _ := http.NewRequest("GET", testURL, body)

	resp := httptest.NewRecorder()

	RedirectHandler(resp, r)

	want := "/"
	wantCode := http.StatusSeeOther

	if !cmp.Equal(wantCode, resp.Code) {
		t.Fatalf(`RedirectHandler("%v, %v"), want match for %d, nil`, resp, r, wantCode)
	}

	if !cmp.Equal(want, resp.Result().Header.Get("Location")) {
		t.Fatalf(`RedirectHandler("%v, %v"), want match for %s, nil`, resp, r, want)
	}
}

func TestRedirectHandler_IncorrectLength(t *testing.T) {
	testURL := "http://localhost:8090/edit/https:/www.reddit.com/"
	body := strings.NewReader("")

	r, _ := http.NewRequest("GET", testURL, body)

	resp := httptest.NewRecorder()

	RedirectHandler(resp, r)

	want := "/"
	wantCode := http.StatusSeeOther

	if !cmp.Equal(wantCode, resp.Code) {
		t.Fatalf(`RedirectHandler("%v, %v"), want match for %d, nil`, resp, r, wantCode)
	}

	if !cmp.Equal(want, resp.Result().Header.Get("Location")) {
		t.Fatalf(`RedirectHandler("%v, %v"), want match for %s, nil`, resp, r, want)
	}
}

func TestRedirectHandler_RedditValidation(t *testing.T) {
	testURL := "http://localhost:8090/edit/https:/www.reddit2.com/r/shortcuts/comments/bkqsv7/what_website_do_you_edit_shortcuts_on/emj1f0l/"
	body := strings.NewReader("")
	r, _ := http.NewRequest("GET", testURL, body)
	resp := httptest.NewRecorder()

	RedirectHandler(resp, r)

	want := "/"
	wantCode := http.StatusSeeOther

	if !cmp.Equal(wantCode, resp.Code) {
		t.Fatalf(`RedirectHandler("%v, %v"), want match for %d, nil`, resp, r, wantCode)
	}

	if !cmp.Equal(want, resp.Result().Header.Get("Location")) {
		t.Fatalf(`RedirectHandler("%v, %v"), want match for %s, nil`, resp, r, want)
	}
}

func TestEditHandler_MalformedRedditRequest(t *testing.T) {
	Initialize("subreddit")
	Templates.Set("home", template.New("home"))

	testURL := "http://localhost:8090/r/subreddit-emj1f0l/submit/"
	body := strings.NewReader("")

	r, _ := http.NewRequest("GET", testURL, body)
	r.SetPathValue("id", "subreddit-emj1f0l")
	ctx := context.WithValue(r.Context(), common.UserCtx, &testUser1)
	r = r.Clone(ctx)

	dataaccess.Initialize("local")

	resp := httptest.NewRecorder()

	EditHandler(resp, r)

	wantCode := http.StatusOK
	wantErrorText := "Not a properly formatted reddit link"
	wantErrorType := "other"

	if !cmp.Equal(wantCode, resp.Code) {
		t.Fatalf(`EditHandler("%v, %v"), want status code match for %d`, resp, r, wantCode)
	}

	if !cmp.Equal(wantErrorText, r.Header.Get("ErrorText")) {
		t.Fatalf(`EditHandler("%v, %v"), want error text match for %s, got %s`, resp, r, wantErrorText, r.Header.Get("ErrorText"))
	}

	if !cmp.Equal(wantErrorType, r.Header.Get("ErrorType")) {
		t.Fatalf(`EditHandler("%v, %v"), want error type match for %s, got %s`, resp, r, wantErrorText, r.Header.Get("ErrorType"))
	}
}

func TestEditHandler_InvalidSubreddit(t *testing.T) {
	Initialize("subreddit")
	Templates.Set("home", template.New("home"))

	testURL := "http://localhost:8090/r/invalid-emj1f0l-emj1f0l/submit/"
	body := strings.NewReader("")

	r, _ := http.NewRequest("GET", testURL, body)
	r.SetPathValue("id", "invalid-emj1f0l-emj1f0l")
	ctx := context.WithValue(r.Context(), common.UserCtx, &testUser1)
	r = r.Clone(ctx)

	dataaccess.Initialize("local")

	resp := httptest.NewRecorder()

	EditHandler(resp, r)

	wantCode := http.StatusOK
	wantErrorText := "r/invalid is not a supported subreddit."
	wantErrorType := "subreddit"

	if !cmp.Equal(wantCode, resp.Code) {
		t.Fatalf(`EditHandler("%v, %v"), want status code match for %d`, resp, r, wantCode)
	}

	if !cmp.Equal(wantErrorText, r.Header.Get("ErrorText")) {
		t.Fatalf(`EditHandler("%v, %v"), want error text match for %s, got %s`, resp, r, wantErrorText, r.Header.Get("ErrorText"))
	}

	if !cmp.Equal(wantErrorType, r.Header.Get("ErrorType")) {
		t.Fatalf(`EditHandler("%v, %v"), want error type match for %s, got %s`, resp, r, wantErrorText, r.Header.Get("ErrorType"))
	}
}

func TestStyleHandler_NoStyles(t *testing.T) {
	testURL := "http://localhost:8090/r/subreddit-emj1f0l-emj1f0l/submit/"
	body := strings.NewReader("")

	r, _ := http.NewRequest("GET", testURL, body)

	dataaccess.Initialize("local")
	resp := httptest.NewRecorder()
	link := &common.Link{LinkId: 1}
	params := styleHandler(resp, r, link)

	want := ParamData{}

	if !cmp.Equal(want, params) {
		t.Fatalf(`StyleHandler("%v, %v, %v"), want param match for %v`, resp, r, link, want)
	}
}

func TestStyleHandler_SomeStyles(t *testing.T) {
	testURL := "http://localhost:8090/r/subreddit-emj1f0l-emj1f0l/submit/"
	body := strings.NewReader("")

	r, _ := http.NewRequest("GET", testURL, body)

	dataaccess.Initialize("local")
	var linkStyles []common.LinkStyle
	linkStyles = make([]common.LinkStyle, 0)
	linkStyles = append(linkStyles, common.LinkStyle{LinkId: 1, Key: "cmt", Value: "top"})
	dataaccess.AddLinkStyles(linkStyles)

	resp := httptest.NewRecorder()
	link := &common.Link{LinkId: 1}
	params := styleHandler(resp, r, link)

	want := ParamData{Cmt: "top"}
	wantLocation := fmt.Sprintf("%s?cmt=top", testURL)

	if !cmp.Equal(want, params) {
		t.Fatalf(`StyleHandler("%v, %v, %v"), want param match for %v`, resp, r, link, want)
	}

	if !cmp.Equal(wantLocation, resp.Result().Header.Get("Location")) {
		t.Fatalf(`StyleHandler("%v, %v, %v"), want location match for %s, got %s`, resp, r, link, wantLocation, resp.Result().Header.Get("Location"))
	}
}

func TestStyleHandler_AllStyles(t *testing.T) {
	testURL := "http://localhost:8090/r/subreddit-emj1f0l-emj1f0l/submit/"
	body := strings.NewReader("")

	r, _ := http.NewRequest("GET", testURL, body)

	dataaccess.Initialize("local")
	var linkStyles []common.LinkStyle
	linkStyles = make([]common.LinkStyle, 0)
	linkStyles = append(linkStyles, common.LinkStyle{LinkId: 1, Key: "cmt", Value: "top"})
	linkStyles = append(linkStyles, common.LinkStyle{LinkId: 1, Key: "brd", Value: "4"})
	linkStyles = append(linkStyles, common.LinkStyle{LinkId: 1, Key: "bc", Value: "1"})
	linkStyles = append(linkStyles, common.LinkStyle{LinkId: 1, Key: "font", Value: "2"})
	linkStyles = append(linkStyles, common.LinkStyle{LinkId: 1, Key: "bold", Value: "true"})
	linkStyles = append(linkStyles, common.LinkStyle{LinkId: 1, Key: "italic", Value: "false"})
	dataaccess.AddLinkStyles(linkStyles)

	resp := httptest.NewRecorder()
	link := &common.Link{LinkId: 1}
	params := styleHandler(resp, r, link)

	want := ParamData{Cmt: "top", Brd: "4", Bc: "1", Font: "2", Bold: "true", Italic: "false"}
	wantLocation := fmt.Sprintf("%s?bc=1&bold=true&brd=4&cmt=top&font=2&italic=false", testURL)

	if !cmp.Equal(want, params) {
		t.Fatalf(`StyleHandler("%v, %v, %v"), want param match for %v`, resp, r, link, want)
	}

	if !cmp.Equal(wantLocation, resp.Result().Header.Get("Location")) {
		t.Fatalf(`StyleHandler("%v, %v, %v"), want location match for %s, got %s`, resp, r, link, wantLocation, resp.Result().Header.Get("Location"))
	}
}

func TestEditHandler_InvalidUser(t *testing.T) {
	Initialize("subreddit")
	Templates.Set("edit", template.New("edit"))

	redditRequest := common.RedditRequest{Subreddit: "subreddit", Article: "emj1f0l", Comment: "emj1f0l"}

	testURL := fmt.Sprintf("http://localhost:8090/r/%s/submit/", redditRequest.AsString())
	body := strings.NewReader("")

	r, _ := http.NewRequest("GET", testURL, body)
	r.SetPathValue("id", redditRequest.AsString())
	ctx := context.WithValue(r.Context(), common.UserCtx, &testUser1)
	r = r.Clone(ctx)

	resp := httptest.NewRecorder()

	dataaccess.Initialize("local")
	link := &common.Link{LinkId: 1, UserId: testUser2.UserId, ImageUrl: "https://example.com/test.png"}
	dataaccess.AddLink(redditRequest, link, testUser2.UserId)

	EditHandler(resp, r)

	wantCode := http.StatusForbidden

	if !cmp.Equal(wantCode, resp.Code) {
		t.Fatalf(`EditHandler("%v, %v"), want status code match for %d`, resp, r, wantCode)
	}
}

func TestEditHandler_HappyPath(t *testing.T) {
	Initialize("subreddit")
	Templates.Set("edit", template.New("edit"))

	redditRequest := common.RedditRequest{Subreddit: "subreddit", Article: "emj1f0l", Comment: "emj1f0l"}

	testURL := fmt.Sprintf("http://localhost:8090/r/%s/submit/", redditRequest.AsString())
	body := strings.NewReader("")

	r, _ := http.NewRequest("GET", testURL, body)
	r.SetPathValue("id", redditRequest.AsString())
	ctx := context.WithValue(r.Context(), common.UserCtx, &testUser1)
	r = r.Clone(ctx)

	resp := httptest.NewRecorder()

	dataaccess.Initialize("local")
	link := &common.Link{LinkId: 1, UserId: testUser1.UserId, ImageUrl: "https://example.com/test.png"}
	dataaccess.AddLink(redditRequest, link, testUser1.UserId)

	EditHandler(resp, r)

	wantCode := http.StatusOK

	if !cmp.Equal(wantCode, resp.Code) {
		t.Fatalf(`EditHandler("%v, %v"), want status code match for %d`, resp, r, wantCode)
	}
}

func TestSaveParams_EmptyParams(t *testing.T) {
	redditRequest := common.RedditRequest{Subreddit: "subreddit", Article: "emj1f0l", Comment: "emj1f0l"}

	dataaccess.Initialize("local")
	ok := saveParams("", redditRequest.AsString())

	want := false

	if !cmp.Equal(want, ok) {
		t.Fatalf(`saveParams("%s, %s"), want match for %t, got %t`, "", redditRequest.AsString(), want, ok)
	}
}

func TestSaveParams_MalformedParams(t *testing.T) {
	redditRequest := common.RedditRequest{Subreddit: "subreddit", Article: "emj1f0l", Comment: "emj1f0l"}

	dataaccess.Initialize("local")
	ok := saveParams("cmt=test", redditRequest.AsString())

	want := false

	if !cmp.Equal(want, ok) {
		t.Fatalf(`saveParams("%s, %s"), want match for %t, got %t`, "", redditRequest.AsString(), want, ok)
	}
}

func TestSaveParams_HappyPath_SingleParam(t *testing.T) {
	redditRequest := common.RedditRequest{Subreddit: "subreddit", Article: "emj1f0l", Comment: "emj1f0l"}

	dataaccess.Initialize("local")
	ok := saveParams("?cmt=test", redditRequest.AsString())

	want := true

	if !cmp.Equal(want, ok) {
		t.Fatalf(`saveParams("%s, %s"), want match for %t, got %t`, "", redditRequest.AsString(), want, ok)
	}
}

func TestSaveParams_HappyPath_MultipleParam(t *testing.T) {
	redditRequest := common.RedditRequest{Subreddit: "subreddit", Article: "emj1f0l", Comment: "emj1f0l"}

	dataaccess.Initialize("local")
	ok := saveParams("?cmt=test&bc=1", redditRequest.AsString())

	want := true

	if !cmp.Equal(want, ok) {
		t.Fatalf(`saveParams("%s, %s"), want match for %t, got %t`, "", redditRequest.AsString(), want, ok)
	}
}

func TestSaveHandler_InvalidRedditRequest(t *testing.T) {
	testURL := fmt.Sprintf("http://localhost:8090/r/%s/submit/", "invalid")
	body := strings.NewReader("")

	r, _ := http.NewRequest("GET", testURL, body)
	r.SetPathValue("id", "invalid")

	resp := httptest.NewRecorder()

	SaveHandler(resp, r)

	wantCode := http.StatusInternalServerError

	if !cmp.Equal(wantCode, resp.Code) {
		t.Fatalf(`SaveHandler("%v, %v"), want status code match for %d`, resp, r, wantCode)
	}
}

func TestSaveHandler_InvalidUser(t *testing.T) {
	Initialize("subreddit")
	Templates.Set("edit", template.New("edit"))

	redditRequest := common.RedditRequest{Subreddit: "subreddit", Article: "emj1f0l", Comment: "emj1f0l"}

	testURL := fmt.Sprintf("http://localhost:8090/r/%s/submit/", redditRequest.AsString())
	body := strings.NewReader("")

	r, _ := http.NewRequest("GET", testURL, body)
	r.SetPathValue("id", redditRequest.AsString())

	resp := httptest.NewRecorder()

	SaveHandler(resp, r)

	wantCode := http.StatusInternalServerError

	if !cmp.Equal(wantCode, resp.Code) {
		t.Fatalf(`SaveHandler("%v, %v"), want status code match for %d`, resp, r, wantCode)
	}
}

func TestSaveHandler_MissingContentType(t *testing.T) {
	Initialize("subreddit")
	Templates.Set("edit", template.New("edit"))

	redditRequest := common.RedditRequest{Subreddit: "subreddit", Article: "emj1f0l", Comment: "emj1f0l"}

	testURL := fmt.Sprintf("http://localhost:8090/r/%s/submit/", redditRequest.AsString())

	// imgInfo, _ := json.Marshal(ImageInfo{ImgData: "", Width: 2, Height: 3, Params: ""})
	body := strings.NewReader("")

	r, _ := http.NewRequest("GET", testURL, body)
	r.SetPathValue("id", redditRequest.AsString())
	ctx := context.WithValue(r.Context(), common.UserCtx, &testUser1)
	r = r.Clone(ctx)

	resp := httptest.NewRecorder()

	SaveHandler(resp, r)

	wantCode := http.StatusInternalServerError

	if !cmp.Equal(wantCode, resp.Code) {
		t.Fatalf(`SaveHandler("%v, %v"), want status code match for %d`, resp, r, wantCode)
	}
}

func TestSaveHandler_BodyDecodeFailure(t *testing.T) {
	Initialize("subreddit")
	Templates.Set("edit", template.New("edit"))

	redditRequest := common.RedditRequest{Subreddit: "subreddit", Article: "emj1f0l", Comment: "emj1f0l"}

	testURL := fmt.Sprintf("http://localhost:8090/r/%s/submit/", redditRequest.AsString())

	// imgInfo, _ := json.Marshal(ImageInfo{ImgData: "", Width: 2, Height: 3, Params: ""})
	body := strings.NewReader("")

	r, _ := http.NewRequest("GET", testURL, body)
	r.Header.Add("Content-Type", "image/webp")
	r.SetPathValue("id", redditRequest.AsString())
	ctx := context.WithValue(r.Context(), common.UserCtx, &testUser1)
	r = r.Clone(ctx)

	resp := httptest.NewRecorder()

	SaveHandler(resp, r)

	wantCode := http.StatusInternalServerError

	if !cmp.Equal(wantCode, resp.Code) {
		t.Fatalf(`SaveHandler("%v, %v"), want status code match for %d`, resp, r, wantCode)
	}
}

func TestSaveHandler_ImageDecodeFailure(t *testing.T) {
	Initialize("subreddit")
	Templates.Set("edit", template.New("edit"))

	redditRequest := common.RedditRequest{Subreddit: "subreddit", Article: "emj1f0l", Comment: "emj1f0l"}

	testURL := fmt.Sprintf("http://localhost:8090/r/%s/submit/", redditRequest.AsString())

	imgInfo, _ := json.Marshal(ImageInfo{ImgData: "%", Width: 2, Height: 3, Params: ""})
	body := strings.NewReader(string(imgInfo))

	r, _ := http.NewRequest("GET", testURL, body)
	r.Header.Add("Content-Type", "image/webp")
	r.SetPathValue("id", redditRequest.AsString())
	ctx := context.WithValue(r.Context(), common.UserCtx, &testUser1)
	r = r.Clone(ctx)

	resp := httptest.NewRecorder()

	SaveHandler(resp, r)

	wantCode := http.StatusInternalServerError

	if !cmp.Equal(wantCode, resp.Code) {
		t.Fatalf(`SaveHandler("%v, %v"), want status code match for %d`, resp, r, wantCode)
	}
}

func TestSaveHandler_FailedUpload(t *testing.T) {
	Templates.Set("edit", template.New("edit"))

	redditRequest := common.RedditRequest{Subreddit: "invalid", Article: "invalid", Comment: "invalid"}

	testURL := fmt.Sprintf("http://localhost:8090/r/%s/submit/", redditRequest.AsString())

	imgInfo, _ := json.Marshal(ImageInfo{ImgData: "", Width: 2, Height: 3, Params: ""})
	body := strings.NewReader(string(imgInfo))

	r, _ := http.NewRequest("GET", testURL, body)
	r.Header.Add("Content-Type", "image/webp")
	r.SetPathValue("id", redditRequest.AsString())
	ctx := context.WithValue(r.Context(), common.UserCtx, &testUser1)
	r = r.Clone(ctx)

	resp := httptest.NewRecorder()

	SaveHandler(resp, r)

	wantCode := http.StatusInternalServerError

	if !cmp.Equal(wantCode, resp.Code) {
		t.Fatalf(`SaveHandler("%v, %v"), want status code match for %d`, resp, r, wantCode)
	}
}

func TestSaveHandler_SuccessfulUpload(t *testing.T) {
	redditRequest := common.RedditRequest{Subreddit: "valid", Article: "emj1f0l", Comment: "emj1f0l"}

	testURL := fmt.Sprintf("http://localhost:8090/r/%s/submit/", redditRequest.AsString())

	imgInfo, _ := json.Marshal(ImageInfo{ImgData: "", Width: 2, Height: 3, Params: ""})
	body := strings.NewReader(string(imgInfo))

	r, _ := http.NewRequest("GET", testURL, body)
	r.Header.Add("Content-Type", "image/webp")
	r.SetPathValue("id", redditRequest.AsString())
	ctx := context.WithValue(r.Context(), common.UserCtx, &testUser1)
	r = r.Clone(ctx)

	resp := httptest.NewRecorder()

	SaveHandler(resp, r)

	wantCode := http.StatusFound
	wantLocation := fmt.Sprintf("/r/%s", redditRequest.AsString())

	if !cmp.Equal(wantCode, resp.Code) {
		t.Fatalf(`SaveHandler("%v, %v"), want status code match for %d`, resp, r, wantCode)
	}

	if !cmp.Equal(wantLocation, resp.Result().Header.Get("Location")) {
		t.Fatalf(`SaveHandler("%v, %v"), want match for %s, got %s`, resp, r, wantLocation, resp.Result().Header.Get("Location"))
	}

}
