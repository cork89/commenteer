package http

import (
	"context"
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

func TestViewHandler_InvalidRedditRequest(t *testing.T) {
	Initialize("valid")
	Templates.Set("view", template.New("view"))

	testURL := "http://localhost:8090/r/invalid/"

	body := strings.NewReader("")

	r, _ := http.NewRequest("GET", testURL, body)
	r.SetPathValue("id", "invalid")

	dataaccess.Initialize("local")

	resp := httptest.NewRecorder()

	ViewHandler(resp, r)

	wantCode := http.StatusSeeOther
	wantLocation := "/"

	if !cmp.Equal(wantCode, resp.Code) {
		t.Fatalf(`ViewHandler("%v, %v"), want status code match for %d`, resp, r, wantCode)
	}

	if !cmp.Equal(wantLocation, resp.Result().Header.Get("Location")) {
		t.Fatalf(`ViewHandler("%v, %v"), want location match for %s, got %s`, resp, r, wantLocation, resp.Result().Header.Get("Location"))
	}
}

func TestViewHandler_NoLinkRetrieved(t *testing.T) {
	Initialize("subreddit")
	Templates.Set("view", template.New("view"))

	redditRequest := common.RedditRequest{Subreddit: "subreddit", Article: "emj1f0l", Comment: "emj1f0l"}

	testURL := fmt.Sprintf("http://localhost:8090/r/%s/", redditRequest.AsString())

	body := strings.NewReader("")

	r, _ := http.NewRequest("GET", testURL, body)
	r.SetPathValue("id", redditRequest.AsString())

	dataaccess.Initialize("local")

	resp := httptest.NewRecorder()

	ViewHandler(resp, r)

	wantCode := http.StatusSeeOther
	wantLocation := "/"

	if !cmp.Equal(wantCode, resp.Code) {
		t.Fatalf(`ViewHandler("%v, %v"), want status code match for %d`, resp, r, wantCode)
	}

	if !cmp.Equal(wantLocation, resp.Result().Header.Get("Location")) {
		t.Fatalf(`ViewHandler("%v, %v"), want location match for %s, got %s`, resp, r, wantLocation, resp.Result().Header.Get("Location"))
	}
}

func TestViewHandler_UserLoggedOut(t *testing.T) {
	Initialize("subreddit")
	Templates.Set("view", template.New("view"))

	redditRequest := common.RedditRequest{Subreddit: "subreddit", Article: "emj1f0l", Comment: "emj1f0l"}

	testURL := fmt.Sprintf("http://localhost:8090/r/%s/", redditRequest.AsString())

	body := strings.NewReader("")

	r, _ := http.NewRequest("GET", testURL, body)
	r.SetPathValue("id", redditRequest.AsString())

	dataaccess.Initialize("local")
	link := common.Link{LinkId: 1}
	dataaccess.AddLink(redditRequest, &link, 1)

	resp := httptest.NewRecorder()

	ViewHandler(resp, r)

	wantCode := http.StatusOK

	if !cmp.Equal(wantCode, resp.Code) {
		t.Fatalf(`HomeHandler("%v, %v"), want status code match for %d, got %d`, resp, r, wantCode, resp.Code)
	}
}

func TestViewHandler_UserNotInContext_NormalLink(t *testing.T) {
	Initialize("subreddit")
	Templates.Set("view", template.New("view"))

	redditRequest := common.RedditRequest{Subreddit: "subreddit", Article: "emj1f0l", Comment: "emj1f0l"}

	testURL := fmt.Sprintf("http://localhost:8090/r/%s/", redditRequest.AsString())

	body := strings.NewReader("")

	r, _ := http.NewRequest("GET", testURL, body)
	r.SetPathValue("id", redditRequest.AsString())

	dataaccess.Initialize("local")
	link := common.Link{LinkId: 1}
	dataaccess.AddLink(redditRequest, &link, testUser2.UserId)
	dataaccess.AddUser(testUser1)

	resp := httptest.NewRecorder()

	ViewHandler(resp, r)

	wantCode := http.StatusOK

	if !cmp.Equal(wantCode, resp.Code) {
		t.Fatalf(`HomeHandler("%v, %v"), want status code match for %d, got %d`, resp, r, wantCode, resp.Code)
	}
}

func TestViewHandler_UserNotInContext_LoggedInLink(t *testing.T) {
	Initialize("subreddit")
	Templates.Set("view", template.New("view"))

	redditRequest := common.RedditRequest{Subreddit: "subreddit", Article: "emj1f0l", Comment: "emj1f0l"}

	testURL := fmt.Sprintf("http://localhost:8090/r/%s/", redditRequest.AsString())

	body := strings.NewReader("")

	r, _ := http.NewRequest("GET", testURL, body)
	r.SetPathValue("id", redditRequest.AsString())

	dataaccess.Initialize("local")
	link := common.Link{LinkId: 1}
	dataaccess.AddLink(redditRequest, &link, testUser1.UserId)
	dataaccess.AddUser(testUser1)

	resp := httptest.NewRecorder()

	ViewHandler(resp, r)

	wantCode := http.StatusOK

	if !cmp.Equal(wantCode, resp.Code) {
		t.Fatalf(`HomeHandler("%v, %v"), want status code match for %d, got %d`, resp, r, wantCode, resp.Code)
	}
}

func TestViewHandler_UserInContext_NormalLink(t *testing.T) {
	Initialize("subreddit")
	Templates.Set("view", template.New("view"))

	redditRequest := common.RedditRequest{Subreddit: "subreddit", Article: "emj1f0l", Comment: "emj1f0l"}

	testURL := fmt.Sprintf("http://localhost:8090/r/%s/", redditRequest.AsString())

	body := strings.NewReader("")

	r, _ := http.NewRequest("GET", testURL, body)
	r.SetPathValue("id", redditRequest.AsString())
	ctx := context.WithValue(r.Context(), common.UserCtx, &testUser1)
	r = r.Clone(ctx)

	dataaccess.Initialize("local")
	link := common.Link{LinkId: 1}
	dataaccess.AddLink(redditRequest, &link, testUser2.UserId)
	dataaccess.AddUser(testUser1)

	resp := httptest.NewRecorder()

	ViewHandler(resp, r)

	wantCode := http.StatusOK

	if !cmp.Equal(wantCode, resp.Code) {
		t.Fatalf(`HomeHandler("%v, %v"), want status code match for %d, got %d`, resp, r, wantCode, resp.Code)
	}
}

func TestViewHandler_UserInContext_LoggedInLink(t *testing.T) {
	Initialize("subreddit")
	Templates.Set("view", template.New("view"))

	redditRequest := common.RedditRequest{Subreddit: "subreddit", Article: "emj1f0l", Comment: "emj1f0l"}

	testURL := fmt.Sprintf("http://localhost:8090/r/%s/", redditRequest.AsString())

	body := strings.NewReader("")

	r, _ := http.NewRequest("GET", testURL, body)
	r.SetPathValue("id", redditRequest.AsString())
	ctx := context.WithValue(r.Context(), common.UserCtx, &testUser1)
	r = r.Clone(ctx)

	dataaccess.Initialize("local")
	link := common.Link{LinkId: 1}
	dataaccess.AddLink(redditRequest, &link, testUser1.UserId)
	dataaccess.AddUser(testUser1)

	resp := httptest.NewRecorder()

	ViewHandler(resp, r)

	wantCode := http.StatusOK

	if !cmp.Equal(wantCode, resp.Code) {
		t.Fatalf(`HomeHandler("%v, %v"), want status code match for %d, got %d`, resp, r, wantCode, resp.Code)
	}
}
