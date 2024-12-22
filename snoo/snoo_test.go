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

	if !cmp.Equal(link, want) || err != nil {
		t.Fatalf(`callRedditApi("%s, %v") = %s, %v, want match for %s, nil`, redditReq.AsString(), user, link.AsString(), err, want.AsString())
	}
}

func TestNewLink_Returned(t *testing.T) {
	t.Setenv("IMGPROXY_URL", "example.com")
	t.Setenv("IMGPROXY_SALT", "salt")
	t.Setenv("IMGPROXY_KEY", "key")

	redditReq := c.RedditRequest{Subreddit: "pics", Article: "1fe0l1d", Comment: "lmlaavt"}
	userId := 1
	cmt1 := c.Comment{Comment: "Just here to watch people who voted for the host of “Celebrity Apprentice” (who had never held elected office before running for President) say that nobody should care what a celebrity says.", Author: "MyDesign630"}
	cmt2 := c.Comment{Comment: "Go read the replies to FuckJerry’s Instagram announcing Swift is voting for Harris. Every reply is “people who care about who celebs vote for are losers”", Author: "Chessh2036"}
	cmt3 := c.Comment{Comment: "Weird for a group that seems to care a lot about what Kevin Sorbo and Kid Rock thinks, even though I never hear about them anymore otherwise.", Author: "mtaw"}
	cmts := []c.Comment{cmt1, cmt2, cmt3}
	want := &c.Link{RedditComments: cmts,
		ImageUrl: "https://i.redd.it/h2y07ob2m3od1.png",
		LinkType: c.Image,
		ProxyUrl: "http://example.com/PumR_I8gs1jSB2qPlMwOIOxiD635pZC-YfKV85RbnZo/resize:fit:1024:0:1/padding:0:0/wm:1:soea:0:0:0.3/background:255:255:255/plain/https://i.redd.it/h2y07ob2m3od1.png"}
	user := &c.User{UserId: userId}

	dataaccess.Initialize("local")

	link, err := testRedditCaller.callRedditApi(redditReq, user)

	if !cmp.Equal(link, want) || err != nil {
		t.Fatalf(`callRedditApi("%s, %v") = %s, %v, want match for %s, nil`, redditReq.AsString(), user, link.AsString(), err, want.AsString())
	}
}

func TestNewLink_SavedToDataAccess(t *testing.T) {
	t.Setenv("IMGPROXY_URL", "example.com")
	t.Setenv("IMGPROXY_SALT", "salt")
	t.Setenv("IMGPROXY_KEY", "key")
	redditReq := c.RedditRequest{Subreddit: "pics", Article: "1fe0l1d", Comment: "lmlaavt"}
	userId := 1
	cmt1 := c.Comment{Comment: "Just here to watch people who voted for the host of “Celebrity Apprentice” (who had never held elected office before running for President) say that nobody should care what a celebrity says.", Author: "MyDesign630"}
	cmt2 := c.Comment{Comment: "Go read the replies to FuckJerry’s Instagram announcing Swift is voting for Harris. Every reply is “people who care about who celebs vote for are losers”", Author: "Chessh2036"}
	cmt3 := c.Comment{Comment: "Weird for a group that seems to care a lot about what Kevin Sorbo and Kid Rock thinks, even though I never hear about them anymore otherwise.", Author: "mtaw"}
	cmts := []c.Comment{cmt1, cmt2, cmt3}
	want := &c.Link{RedditComments: cmts, ImageUrl: "https://i.redd.it/h2y07ob2m3od1.png", LinkType: c.Image, ProxyUrl: "http://example.com/PumR_I8gs1jSB2qPlMwOIOxiD635pZC-YfKV85RbnZo/resize:fit:1024:0:1/padding:0:0/wm:1:soea:0:0:0.3/background:255:255:255/plain/https://i.redd.it/h2y07ob2m3od1.png"}
	user := &c.User{UserId: userId}

	dataaccess.Initialize("local")

	_, err := testRedditCaller.callRedditApi(redditReq, user)

	link, _ := dataaccess.GetLink(redditReq)

	if !cmp.Equal(link, want) || err != nil {
		t.Fatalf(`callRedditApi("%s, %v") = %s, %v, want match for %s, nil`, redditReq.AsString(), user, link.AsString(), err, want.AsString())
	}
}
