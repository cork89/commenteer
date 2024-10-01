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
var redditJWTSecret string

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

func (atb AccessTokenBody) GetExpireDtTm() time.Time {
	return time.Now().UTC().Add(time.Second * time.Duration(atb.ExpiresIn))
}

func CreateUserJwt(user common.UserCookie) string {
	key := []byte(redditJWTSecret)
	signer, err := jwt.NewSignerHS(jwt.HS256, key)
	if err != nil {
		log.Printf("failed to create jwt signer: %v\n", err)
	}

	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Hour * time.Duration(24))),
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
	// cookieVal, err := json.Marshal(userCookie)
	// if err != nil {
	// 	log.Printf("error marshalling cookie, %v\n", err)
	// }

	jwt := CreateUserJwt(userCookie)

	cookie := http.Cookie{
		Name:     cookieName,
		Value:    jwt,
		Path:     "/",
		MaxAge:   0,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	return cookie
}

func GetUserCookie(r *http.Request) (userCookie *common.UserCookie, ok bool) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		log.Printf("No cookie found, %v\n", err)
		return nil, false
	}

	key := []byte(redditJWTSecret)
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

	// if err = json.Unmarshal(cookieVal, userCookie); err != nil {
	// 	log.Printf("error unmarshalling cookie, %v\n", err)
	// }
	return &user.UserCookie, true
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

func callAccessTokenApi(postBody PostBody) (*http.Response, error) {
	// reader := bytes.NewReader(bodyBytes)

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
	if state != redditOAuthState {
		log.Println("wrong state :(")
	}

	if code == "" {
		log.Println("no code bro")
	}

	log.Printf("state: %s, code: %s\n", state, code)

	body := PostBody{
		GrantType:   "authorization_code",
		Code:        code,
		RedirectUri: redditRedirectUri,
	}

	res, err := callAccessTokenApi(body)
	if err != nil {
		log.Println("failed to get access token ", err)
		return nil, false
	}

	// if (res.StatusCode == http.Status)

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
	log.Printf("resBody: %s\n", resBody)

	err = json.Unmarshal(resBody, &accessToken)
	if err != nil {
		log.Println("error unmarshalling response")
		return nil, false
	}

	log.Printf("accessToken: %v\n", accessToken)
	return accessToken, true
}

func init() {
	redditAccessToken = os.Getenv("REDDIT_OAUTH_STATE")
	redditRedirectUri = os.Getenv("REDDIT_REDIRECT_URI")
	redditClientId = os.Getenv("REDDIT_CLIENT_ID")
	redditSecret = os.Getenv("REDDIT_SECRET")
	redditJWTSecret = os.Getenv("REDDIT_JWT_SECRET")

	redditBasicAuth = func() string {
		auth := fmt.Sprintf("%s:%s", redditClientId, redditSecret)
		return base64.StdEncoding.EncodeToString([]byte(auth))
	}()
}
