package snoo

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"main/common"
	"main/dataaccess"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/cristalhq/jwt/v5"
)

var redditOAuthState string
var redditRedirectUri string
var redditClientId string
var redditSecret string
var redditBasicAuth string
var RedditAuthUrl string

var client http.Client

type PostBody struct {
	GrantType    string `json:"grant_type"`
	Code         string `json:"code,omitempty"`
	RedirectUri  string `json:"redirect_uri,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

type AccessTokenBody struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	RefreshToken string `json:"refresh_token"`
}

type Subreddit struct {
	Username string `json:"display_name_prefixed"`
	IconImg  string `json:"icon_img"`
}

type UserResponse struct {
	Data Subreddit `json:"subreddit"`
}

const (
	cookieName string = "commenteerCookie"
)

var clock common.Clock = common.RealClock{}

type RedditAuthCaller interface {
	callAccessTokenApi(PostBody) (*http.Response, error)
	callRefreshAccessTokenApi(PostBody) (*http.Response, error)
}

type RealRedditAuthCaller struct{}

var authCaller RedditAuthCaller = RealRedditAuthCaller{}

func (atb AccessTokenBody) GetExpireDtTm() time.Time {
	return clock.Now().UTC().Add(time.Second * time.Duration(atb.ExpiresIn))
}

func CreateUserJwt(user common.UserCookie) string {
	key := []byte(os.Getenv("REDDIT_JWT_SECRET"))

	signer, err := jwt.NewSignerHS(jwt.HS256, key)
	if err != nil {
		log.Printf("failed to create jwt signer: %v\n", err)
	}
	fmt.Println(clock.Now())
	fmt.Println(clock.Now().UTC())
	tm := clock.Now().UTC().Add(time.Hour * time.Duration(24))
	fmt.Println(tm.Unix())
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(tm),
		Subject:   user.Username,
	}

	builder := jwt.NewBuilder(signer)
	token, err := builder.Build(claims)
	if err != nil {
		log.Printf("failed to build jwt token: %v\n", err)
	}

	return token.String()
}

func CreateUserCookie(userCookie common.UserCookie) http.Cookie {
	jwt := CreateUserJwt(userCookie)

	cookie := http.Cookie{
		Name:     cookieName,
		Value:    jwt,
		Path:     "/",
		MaxAge:   int(time.Duration(2160 * time.Hour).Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	return cookie
}

func GetUserCookie(r *http.Request) (userCookie *common.User, ok bool) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		log.Printf("No cookie found, %v\n", err)
		return nil, false
	}

	key := []byte(os.Getenv("REDDIT_JWT_SECRET"))
	verifier, err := jwt.NewVerifierHS(jwt.HS256, key)
	if err != nil {
		log.Printf("failed to create jwt verifier: %v\n", err)
		return nil, false
	}

	cookieVal := []byte(cookie.Value)
	newToken, err := jwt.Parse(cookieVal, verifier)
	if err != nil {
		log.Printf("failed to parse cookie: %v\n", err)
		return nil, false
	}

	var claims jwt.RegisteredClaims
	if err = json.Unmarshal(newToken.Claims(), &claims); err != nil {
		log.Printf("failed to unmarshal jwt claims: %v\n", err)
		return nil, false
	}

	user, ok := dataaccess.GetUser(claims.Subject)
	if !ok {
		return nil, false
	}

	return user, true
}

func GetUserData(accessToken AccessTokenBody) (user common.User) {
	req, err := http.NewRequest("GET", "https://oauth.reddit.com/api/v1/me", nil)
	if err != nil {
		log.Printf("error creating user data request, %v\n", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken.AccessToken))

	res, err := client.Do(req)
	if err != nil {
		log.Printf("error retrieving user data request, %v\n", err)
	}
	defer res.Body.Close()
	var userResponse UserResponse

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("error reading response body, %v\n", err)
	}
	err = json.Unmarshal(body, &userResponse)

	if err != nil {
		log.Printf("error unmarshalling body: %v\n", err)
	}

	user.Username = userResponse.Data.Username
	iconUrl := strings.Replace(userResponse.Data.IconImg, "&amp;", "&", -1)
	user.IconUrl = iconUrl
	user.AccessToken = accessToken.AccessToken
	user.RefreshExpireDtTm = accessToken.GetExpireDtTm()
	user.RefreshToken = accessToken.RefreshToken
	return user
}

func (RealRedditAuthCaller) callRefreshAccessTokenApi(postBody PostBody) (*http.Response, error) {
	data := url.Values{}
	data.Set("grant_type", postBody.GrantType)
	data.Set("refresh_token", postBody.RefreshToken)

	req, err := http.NewRequest("POST", "https://www.reddit.com/api/v1/access_token", strings.NewReader(data.Encode()))

	if err != nil {
		log.Println("failed to create post request ", err)
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", redditBasicAuth))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	return client.Do(req)
}

func (RealRedditAuthCaller) callAccessTokenApi(postBody PostBody) (*http.Response, error) {
	data := url.Values{}
	data.Set("grant_type", postBody.GrantType)
	data.Set("code", postBody.Code)
	data.Set("redirect_uri", postBody.RedirectUri)

	req, err := http.NewRequest("POST", "https://www.reddit.com/api/v1/access_token", strings.NewReader(data.Encode()))

	if err != nil {
		log.Println("failed to create post request ", err)
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", redditBasicAuth))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	return client.Do(req)
}

func GetRedditAccessToken(state string, code string) (accessToken *AccessTokenBody, ok bool) {
	if state != os.Getenv("REDDIT_OAUTH_STATE") {
		log.Println("wrong state :(")
	}

	if code == "" {
		log.Println("no code bro")
	}

	body := PostBody{
		GrantType:   "authorization_code",
		Code:        code,
		RedirectUri: redditRedirectUri,
	}

	res, err := authCaller.callAccessTokenApi(body)
	if err != nil {
		log.Println("failed to get access token ", err)
		return nil, false
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Println("error closing body: ", err)
		}
	}()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println("failed to read response body")
		return nil, false
	}

	err = json.Unmarshal(resBody, &accessToken)
	if err != nil {
		log.Println("error unmarshalling response")
		return nil, false
	}
	return accessToken, true
}

func RefreshRedditAccessToken(user *common.User) (*common.User, bool) {
	var accessToken *AccessTokenBody
	body := PostBody{
		GrantType:    "refresh_token",
		RefreshToken: user.RefreshToken,
	}

	res, err := authCaller.callRefreshAccessTokenApi(body)
	if err != nil {
		log.Println("failed to get access token ", err)
		return user, false
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Println("error closing body: ", err)
		}
	}()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println("failed to read response body")
		return user, false
	}

	err = json.Unmarshal(resBody, &accessToken)
	if err != nil {
		log.Println("error unmarshalling response")
		return user, false
	}
	go dataaccess.UpdateUser(user.Username, accessToken.AccessToken, accessToken.GetExpireDtTm())
	user.AccessToken = accessToken.AccessToken
	user.RefreshExpireDtTm = accessToken.GetExpireDtTm()
	return user, true
}

func DecrementUserUploadCount(user *common.User) {
	ok := dataaccess.DecrementUserUploadCount(user.UserId)

	if !ok {
		log.Printf("Problem decrementing user upload count for userId=%d\n", user.UserId)
	}
}

func RefreshUserUploadCount(user *common.User) {
	var newCount int
	if user.Subscribed {
		newCount = 50
	} else {
		newCount = 10
	}
	ok := dataaccess.RefreshUserUploadCount(user.UserId, newCount)

	if !ok {
		log.Printf("Problem refreshing user upload count for userId=%d\n", user.UserId)
	}
}

func init() {
	redditOAuthState = os.Getenv("REDDIT_OAUTH_STATE")
	redditRedirectUri = os.Getenv("REDDIT_REDIRECT_URI")
	redditClientId = os.Getenv("REDDIT_CLIENT_ID")
	redditSecret = os.Getenv("REDDIT_SECRET")

	RedditAuthUrl = fmt.Sprintf("https://www.reddit.com/api/v1/authorize?client_id=%s&response_type=code&state=%s&redirect_uri=%s&duration=permanent&scope=read,identity", redditClientId, redditOAuthState, redditRedirectUri)

	redditBasicAuth = func() string {
		auth := fmt.Sprintf("%s:%s", redditClientId, redditSecret)
		return base64.StdEncoding.EncodeToString([]byte(auth))
	}()
}
