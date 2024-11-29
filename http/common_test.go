package http

import (
	"main/common"
	c "main/common"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var testUser1 = common.User{UserId: 1, UserCookie: common.UserCookie{Username: "test1"}}
var testUser2 = common.User{UserId: 2, UserCookie: common.UserCookie{Username: "test2"}}

func TestExtractRedditRequest_MalformedRedditRequest(t *testing.T) {
	testURL := "http://localhost:8090/r/emj1f0l/"
	body := strings.NewReader("")

	r, _ := http.NewRequest("GET", testURL, body)

	_, err := ExtractRedditRequest(r)

	want := "malformed url"

	if err == nil || !cmp.Equal(err.Error(), want) {
		t.Fatalf(`ExtractRedditRequest("%v")=nil, %s, want match for %s`, r, err.Error(), want)
	}
}

func TestExtractRedditRequest_ValidRedditRequest(t *testing.T) {
	testURL := "http://localhost:8090/r/subreddit-emj1f0l-emj1f0l/"
	body := strings.NewReader("")

	r, _ := http.NewRequest("GET", testURL, body)
	r.SetPathValue("id", "subreddit-emj1f0l-emj1f0l")

	redditRequest, err := ExtractRedditRequest(r)

	want := &c.RedditRequest{Subreddit: "subreddit", Article: "emj1f0l", Comment: "emj1f0l"}

	if err != nil || !cmp.Equal(redditRequest, want) {
		t.Fatalf(`ExtractRedditRequest("%v")=%v,%v, want match for %v, nil`, r, redditRequest, err, want)
	}
}

func ExtractUsername_InvalidUsername_TooShort(t *testing.T) {
	testURL := "http://localhost:8090/u/a/"
	body := strings.NewReader("")

	r, _ := http.NewRequest("GET", testURL, body)
	r.SetPathValue("username", "a")

	username, err := ExtractUsername(r)

	want := ""
	wantError := "invalid username"

	if err == nil || !cmp.Equal(err.Error(), wantError) || !cmp.Equal(username, want) {
		t.Fatalf(`ExtractUsername("%v")=%s,%v, want match for %s, %v`, r, username, err, want, wantError)
	}
}

func ExtractUsername_InvalidUsername_TooLong(t *testing.T) {
	testURL := "http://localhost:8090/u/abcdefghiklmnopqrstuvwxyz/"
	body := strings.NewReader("")

	r, _ := http.NewRequest("GET", testURL, body)
	r.SetPathValue("username", "abcdefghiklmnopqrstuvwxyz")

	username, err := ExtractUsername(r)

	want := ""
	wantError := "invalid username"

	if err == nil || !cmp.Equal(err.Error(), wantError) || !cmp.Equal(username, want) {
		t.Fatalf(`ExtractUsername("%v")=%s,%v, want match for %s, %v`, r, username, err, want, wantError)
	}
}

func ExtractUsername_InvalidUsername_InvalidCharacters(t *testing.T) {
	testURL := "http://localhost:8090/u/abc%/"
	body := strings.NewReader("")

	r, _ := http.NewRequest("GET", testURL, body)
	r.SetPathValue("username", "abc%")

	username, err := ExtractUsername(r)

	want := ""
	wantError := "invalid username"

	if err == nil || !cmp.Equal(err.Error(), wantError) || !cmp.Equal(username, want) {
		t.Fatalf(`ExtractUsername("%v")=%s,%v, want match for %s, %v`, r, username, err, want, wantError)
	}
}

func ExtractUsername_ValidUsername(t *testing.T) {
	testURL := "http://localhost:8090/u/abc/"
	body := strings.NewReader("")

	r, _ := http.NewRequest("GET", testURL, body)
	r.SetPathValue("username", "abc")

	username, err := ExtractUsername(r)

	want := "abc"

	if err != nil || !cmp.Equal(username, want) {
		t.Fatalf(`ExtractUsername("%v")=%s,%v, want match for %s, %v`, r, username, err, want, nil)
	}
}
