package http

import (
	"context"
	"main/common"
	"main/dataaccess"
	"main/snoo"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"text/template"

	"github.com/google/go-cmp/cmp"
)

func TestHomeHandler_UserLoggedOut(t *testing.T) {
	Templates.Set("home", template.New("home"))

	testURL := "http://localhost:8090/"

	body := strings.NewReader("")

	r, _ := http.NewRequest("GET", testURL, body)
	resp := httptest.NewRecorder()

	HomeHandler(resp, r)

	wantCode := http.StatusOK

	if !cmp.Equal(wantCode, resp.Code) {
		t.Fatalf(`HomeHandler("%v, %v"), want status code match for %d`, resp, r, wantCode)
	}
}

func TestHomeHandler_UserNotInContext(t *testing.T) {
	Initialize("valid")
	Templates.Set("home", template.New("home"))

	testURL := "http://localhost:8090/"

	body := strings.NewReader("")

	r, _ := http.NewRequest("GET", testURL, body)
	cookie := snoo.CreateUserCookie(testUser1.UserCookie)
	r.AddCookie(&cookie)
	// ctx := context.WithValue(r.Context(), common.UserCtx, &testUser1)
	// r = r.Clone(ctx)
	dataaccess.Initialize("local")
	dataaccess.AddUser(testUser1)

	resp := httptest.NewRecorder()

	HomeHandler(resp, r)

	wantCode := http.StatusOK

	if !cmp.Equal(wantCode, resp.Code) {
		t.Fatalf(`HomeHandler("%v, %v"), want status code match for %d`, resp, r, wantCode)
	}
}

func TestHomeHandler_UserInContext(t *testing.T) {
	Initialize("valid")
	Templates.Set("home", template.New("home"))

	testURL := "http://localhost:8090/"

	body := strings.NewReader("")

	r, _ := http.NewRequest("GET", testURL, body)
	ctx := context.WithValue(r.Context(), common.UserCtx, &testUser1)
	r = r.Clone(ctx)
	dataaccess.Initialize("local")

	resp := httptest.NewRecorder()

	HomeHandler(resp, r)

	wantCode := http.StatusOK

	if !cmp.Equal(wantCode, resp.Code) {
		t.Fatalf(`HomeHandler("%v, %v"), want status code match for %d`, resp, r, wantCode)
	}
}
