package snoo

import (
	c "main/common"
	"main/dataaccess"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

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
		Name:     CookieName,
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
		Name:     CookieName,
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
		Name:     CookieName,
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
		Name:     CookieName,
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
		Name:     CookieName,
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
