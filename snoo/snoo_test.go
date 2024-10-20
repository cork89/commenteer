package snoo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	c "main/common"
	"main/dataaccess"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

type FakeRedditCaller struct{}

var testRedditCaller RedditCaller = &FakeRedditCaller{}

type FakeClock struct{}

func (FakeClock) Now() time.Time {
	dttm, _ := time.Parse("2006-01-02T15:04:05+0000", "2024-10-20T00:00:00+0000")
	return dttm
}

type FakeRedditAuthCaller struct{}

func (FakeRedditAuthCaller) callAccessTokenApi(postBody PostBody) (*http.Response, error) {
	tokenBody := AccessTokenBody{AccessToken: "test"}
	tokenBodyBytes, _ := json.Marshal(tokenBody)

	res := http.Response{}
	res.Body = io.NopCloser(bytes.NewReader(tokenBodyBytes))

	return &res, nil
}

func (FakeRedditAuthCaller) callRefreshAccessTokenApi(postBody PostBody) (*http.Response, error) {
	tokenBody := AccessTokenBody{AccessToken: "accesstoken", ExpiresIn: 86400}
	tokenBodyBytes, _ := json.Marshal(tokenBody)

	res := http.Response{}
	res.Body = io.NopCloser(bytes.NewReader(tokenBodyBytes))

	return &res, nil
}

type ErrorRedditAuthCaller struct{}

func (ErrorRedditAuthCaller) callAccessTokenApi(postBody PostBody) (*http.Response, error) {
	return nil, errors.New("test")
}

func (ErrorRedditAuthCaller) callRefreshAccessTokenApi(postBody PostBody) (*http.Response, error) {
	return nil, errors.New("test")
}

func (f FakeRedditCaller) callRedditApi(req c.RedditRequest, user *c.User) (link *c.Link, err error) {
	link, ok := dataaccess.GetLink(req)
	if ok {
		return link, nil
	}

	body, err := os.ReadFile("../test/test4.json")
	if err != nil {
		fmt.Println(err)
	}

	jsonData := make(JsonData, 2)
	for i := range jsonData {
		jsonData[i] = make(map[string]interface{})
	}

	err = json.Unmarshal(body, &jsonData)
	if err != nil {
		log.Printf("error unmarshalling: %s\n", err)
	}
	link = parseJsonData(jsonData, req.Comment)
	link.ProxyUrl, err = GetImgProxyUrl(link.ImageUrl)
	if err != nil {
		return CreateErrorLink(), nil
	}
	// go addToCache(req, link)
	dataaccess.AddLink(req, link, user.UserId)
	return link, nil
}

// snoo.go

func TestExistingLink_Returned(t *testing.T) {
	redditReq := c.RedditRequest{Subreddit: "test", Article: "test", Comment: "test"}
	userId := 1
	want := &c.Link{UserId: userId}
	user := &c.User{UserId: userId}

	dataaccess.Initialize("local")
	dataaccess.AddLink(redditReq, want, userId)

	link, err := testRedditCaller.callRedditApi(redditReq, user)

	fmt.Println(link)
	fmt.Println(want)

	if !cmp.Equal(link, want) || err != nil {
		t.Fatalf(`callRedditApi("%s, %v") = %s, %v, want match for %s, nil`, redditReq.AsString(), user, link.AsString(), err, want.AsString())
	}
}

func TestNewLink_Returned(t *testing.T) {
	redditReq := c.RedditRequest{Subreddit: "pics", Article: "1fe0l1d", Comment: "lmlaavt"}
	userId := 1
	cmt1 := c.Comment{Comment: "Just here to watch people who voted for the host of “Celebrity Apprentice” (who had never held elected office before running for President) say that nobody should care what a celebrity says.", Author: "MyDesign630"}
	cmt2 := c.Comment{Comment: "Go read the replies to FuckJerry’s Instagram announcing Swift is voting for Harris. Every reply is “people who care about who celebs vote for are losers”", Author: "Chessh2036"}
	cmt3 := c.Comment{Comment: "Weird for a group that seems to care a lot about what Kevin Sorbo and Kid Rock thinks, even though I never hear about them anymore otherwise.", Author: "mtaw"}
	cmts := []c.Comment{cmt1, cmt2, cmt3}
	want := &c.Link{RedditComments: cmts, ImageUrl: "https://i.redd.it/h2y07ob2m3od1.png", LinkType: c.Image, ProxyUrl: "http://localhost:8080/crYrtsQ3GJQ_H7yXUlf0PrPZmlK7NGBYqT9WkvAF09Q/resize:fit:1024:0:1/padding:0:0/wm:1:soea:0:0:0.5/background:255:255:255/plain/https://i.redd.it/h2y07ob2m3od1.png"}
	user := &c.User{UserId: userId}

	dataaccess.Initialize("local")

	link, err := testRedditCaller.callRedditApi(redditReq, user)

	if !cmp.Equal(link, want) || err != nil {
		t.Fatalf(`callRedditApi("%s, %v") = %s, %v, want match for %s, nil`, redditReq.AsString(), user, link.AsString(), err, want.AsString())
	}
}

func TestNewLink_SavedToDataAccess(t *testing.T) {
	redditReq := c.RedditRequest{Subreddit: "pics", Article: "1fe0l1d", Comment: "lmlaavt"}
	userId := 1
	cmt1 := c.Comment{Comment: "Just here to watch people who voted for the host of “Celebrity Apprentice” (who had never held elected office before running for President) say that nobody should care what a celebrity says.", Author: "MyDesign630"}
	cmt2 := c.Comment{Comment: "Go read the replies to FuckJerry’s Instagram announcing Swift is voting for Harris. Every reply is “people who care about who celebs vote for are losers”", Author: "Chessh2036"}
	cmt3 := c.Comment{Comment: "Weird for a group that seems to care a lot about what Kevin Sorbo and Kid Rock thinks, even though I never hear about them anymore otherwise.", Author: "mtaw"}
	cmts := []c.Comment{cmt1, cmt2, cmt3}
	want := &c.Link{RedditComments: cmts, ImageUrl: "https://i.redd.it/h2y07ob2m3od1.png", LinkType: c.Image, ProxyUrl: "http://localhost:8080/crYrtsQ3GJQ_H7yXUlf0PrPZmlK7NGBYqT9WkvAF09Q/resize:fit:1024:0:1/padding:0:0/wm:1:soea:0:0:0.5/background:255:255:255/plain/https://i.redd.it/h2y07ob2m3od1.png"}
	user := &c.User{UserId: userId}

	dataaccess.Initialize("local")

	_, err := testRedditCaller.callRedditApi(redditReq, user)

	link, _ := dataaccess.GetLink(redditReq)

	if !cmp.Equal(link, want) || err != nil {
		t.Fatalf(`callRedditApi("%s, %v") = %s, %v, want match for %s, nil`, redditReq.AsString(), user, link.AsString(), err, want.AsString())
	}
}

// image.go

func TestGetImgProxy(t *testing.T) {
	t.Setenv("IMGPROXY_URL", "example.com")
	t.Setenv("IMGPROXY_SALT", "salt")
	t.Setenv("IMGPROXY_KEY", "key")

	want := "http://example.com/4Kymlr9EvBzkC-KlnbBh9q39W7XEJ13c4UD8ZvxzB14/resize:fit:1024:0:1/padding:0:0/wm:1:soea:0:0:0.5/background:255:255:255/plain/example.jpg"

	url, err := GetImgProxyUrl("example.jpg")

	if !cmp.Equal(url, want) || err != nil {
		t.Fatalf(`GetImgProxyUrl("example.jpg") = %s, want match for %s`, url, want)
	}

}

// users.go

func TestCreateUserJwt(t *testing.T) {
	clock = FakeClock{}
	t.Setenv("REDDIT_JWT_SECRET", "secret")
	user := c.UserCookie{Username: "test"}
	want := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0IiwiZXhwIjoxNzI5NDY4ODAwfQ.o6g35x0DLhID8hLzn9DmUQa_2ZQOob2h9-QgG2yaEy8"

	jwt := CreateUserJwt(user)

	if !cmp.Equal(jwt, want) {
		t.Fatalf(`CreateUserJwt("%v") = %s, want match for %s`, user, jwt, want)
	}
}

func TestCreateUserCookie(t *testing.T) {
	clock = FakeClock{}
	t.Setenv("REDDIT_JWT_SECRET", "secret")
	user := c.UserCookie{Username: "test"}
	jwt := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0IiwiZXhwIjoxNzI5NDY4ODAwfQ.o6g35x0DLhID8hLzn9DmUQa_2ZQOob2h9-QgG2yaEy8"

	want := http.Cookie{
		Name:     cookieName,
		Value:    jwt,
		Path:     "/",
		MaxAge:   int(time.Duration(2160 * time.Hour).Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	cookie := CreateUserCookie(user)

	if !cmp.Equal(cookie, want) {
		t.Fatalf(`CreateUserCookie("%v") = %v, want match for %v`, user, cookie, want)
	}
}

func TestGetUserCookie_CookieMissing(t *testing.T) {
	var want *c.User
	wantOk := false
	req := &http.Request{}

	cookie, ok := GetUserCookie(req)

	if !cmp.Equal(cookie, want) || !cmp.Equal(ok, wantOk) {
		t.Fatalf(`GetUserCookie("%v") = %v, %t, want match for %v, %t`, req, cookie, ok, want, wantOk)
	}
}

func TestGetUserCookie_CookieParseFailure_BadFormat(t *testing.T) {
	var want *c.User
	wantOk := false
	t.Setenv("REDDIT_JWT_SECRET", "secret")
	jwt := "badformat"

	wantCookie := http.Cookie{
		Name:     cookieName,
		Value:    jwt,
		Path:     "/",
		MaxAge:   int(time.Duration(2160 * time.Hour).Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	req := http.Request{}
	req.Header = make(http.Header)
	req.AddCookie(&wantCookie)

	cookie, ok := GetUserCookie(&req)

	if !cmp.Equal(cookie, want) || ok != wantOk {
		t.Fatalf(`GetUserCookie("%v") = %v, %t, want match for %v, %t`, req, cookie, ok, want, wantOk)
	}
}

func TestGetUserCookie_CookieParseFailure_Badjwt(t *testing.T) {
	var want *c.User
	wantOk := false
	t.Setenv("REDDIT_JWT_SECRET", "secret")
	jwt := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0IiwiZXhwIjoxNzI5NDY4ODAwfQ.o6g35x0DLhID8hLzn9DmUQa_2ZQOob2h9-QgG2yaEyf"

	wantCookie := http.Cookie{
		Name:     cookieName,
		Value:    jwt,
		Path:     "/",
		MaxAge:   int(time.Duration(2160 * time.Hour).Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	req := http.Request{}
	req.Header = make(http.Header)
	req.AddCookie(&wantCookie)

	cookie, ok := GetUserCookie(&req)

	if !cmp.Equal(cookie, want) || ok != wantOk {
		t.Fatalf(`GetUserCookie("%v") = %v, %t, want match for %v, %t`, req, cookie, ok, want, wantOk)
	}
}

func TestGetUserCookie_CookieDataAccess_NoUserFound(t *testing.T) {
	var want *c.User
	wantOk := false
	t.Setenv("REDDIT_JWT_SECRET", "secret")
	dataaccess.Initialize("local")

	jwt := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0IiwiZXhwIjoxNzI5NDY4ODAwfQ.o6g35x0DLhID8hLzn9DmUQa_2ZQOob2h9-QgG2yaEy8"

	wantCookie := http.Cookie{
		Name:     cookieName,
		Value:    jwt,
		Path:     "/",
		MaxAge:   int(time.Duration(2160 * time.Hour).Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	req := http.Request{}
	req.Header = make(http.Header)
	req.AddCookie(&wantCookie)

	cookie, ok := GetUserCookie(&req)

	if !cmp.Equal(cookie, want) || ok != wantOk {
		t.Fatalf(`GetUserCookie("%v") = %v, %t, want match for %v, %t`, req, cookie, ok, want, wantOk)
	}
}

func TestGetUserCookie_CookieDataAccess_UserFound(t *testing.T) {
	want, wantOk := &c.User{UserCookie: c.UserCookie{Username: "test"}}, true
	t.Setenv("REDDIT_JWT_SECRET", "secret")
	dataaccess.Initialize("local")
	dataaccess.AddUser(*want)

	jwt := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0IiwiZXhwIjoxNzI5NDY4ODAwfQ.o6g35x0DLhID8hLzn9DmUQa_2ZQOob2h9-QgG2yaEy8"

	wantCookie := http.Cookie{
		Name:     cookieName,
		Value:    jwt,
		Path:     "/",
		MaxAge:   int(time.Duration(2160 * time.Hour).Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	req := http.Request{}
	req.Header = make(http.Header)
	req.AddCookie(&wantCookie)

	cookie, ok := GetUserCookie(&req)

	if !cmp.Equal(cookie, want) || ok != wantOk {
		t.Fatalf(`GetUserCookie("%v") = %v, %t, want match for %v, %t`, req, cookie, ok, want, wantOk)
	}
}

func TestGetRedditAccessToken_SuccessResponse(t *testing.T) {
	t.Setenv("REDDIT_OAUTH_STATE", "state")
	authCaller = FakeRedditAuthCaller{}
	state := "state"
	code := "code"

	want := &AccessTokenBody{AccessToken: "test"}
	wantOk := true

	tokenBody, ok := GetRedditAccessToken("state", "code")

	if !cmp.Equal(tokenBody, want) || ok != wantOk {
		t.Fatalf(`GetRedditAccessToken("%s, %s") = %v, %t, want match for %v, %t`, state, code, tokenBody, ok, want, wantOk)
	}
}

func TestGetRedditAccessToken_ErrorResponse(t *testing.T) {
	t.Setenv("REDDIT_OAUTH_STATE", "state")
	authCaller = ErrorRedditAuthCaller{}
	state := "state"
	code := "code"

	want := &AccessTokenBody{}
	wantOk := false

	tokenBody, ok := GetRedditAccessToken("state", "code")

	if !cmp.Equal(tokenBody, want) && ok != wantOk {
		t.Fatalf(`GetRedditAccessToken("%s, %s") = %v, %t, want match for %v, %t`, state, code, tokenBody, ok, want, wantOk)
	}
}

func TestRefreshRedditAccessToken_SuccessResponse(t *testing.T) {
	clock = FakeClock{}
	authCaller = &FakeRedditAuthCaller{}

	want := &c.User{UserCookie: c.UserCookie{Username: "test", AccessToken: "accesstoken", RefreshExpireDtTm: clock.Now().Add(24 * time.Hour)}, RefreshToken: "refresh"}
	wantOk := true

	refreshUser := c.User{RefreshToken: "refresh", UserCookie: c.UserCookie{Username: "test"}}

	user, ok := RefreshRedditAccessToken(&refreshUser)

	if !cmp.Equal(user, want) || ok != wantOk {
		t.Fatalf(`GetRedditAccessToken("%v") = %v, %t, want match for %v, %t`, refreshUser, user, ok, want, wantOk)
	}
}

func TestRefreshRedditAccessToken_ErrorResponse(t *testing.T) {
	clock = FakeClock{}
	authCaller = &ErrorRedditAuthCaller{}

	want := &c.User{RefreshToken: "refresh"}
	wantOk := false

	refreshUser := c.User{RefreshToken: "refresh"}

	user, ok := RefreshRedditAccessToken(&refreshUser)

	if !cmp.Equal(user, want) || ok != wantOk {
		t.Fatalf(`GetRedditAccessToken("%v") = %v, %t, want match for %v, %t`, refreshUser, user, ok, want, wantOk)
	}
}
