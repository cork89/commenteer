package snoo

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	c "main/common"
	"main/dataaccess"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

const (
	jpg  string = "jpg"
	jpeg string = "jpeg"
	png  string = "png"
	webp string = "webp"
)

type JsonData []map[string]interface{}

var linkCache map[c.RedditRequest]c.Link

var redditAccessToken string
var CdnBaseUrl string

var redditCaller RedditCaller

type RedditCaller interface {
	callRedditApi(req c.RedditRequest, user *c.User) (link *c.Link, err error)
}

type RealRedditCaller struct{}

type FakeRedditCaller struct{}

func parseCommentData(data map[string]interface{}, comments []c.Comment, commentId string, depth int) []c.Comment {
	log.Printf("Entering ParseCommentData, depth: %d", depth)
	comment := c.Comment{Comment: data["body"].(string), Author: data["author"].(string)}
	updatedComments := append(comments, comment)
	if data["id"].(string) == commentId {
		return updatedComments
	}
	replyComment := data["replies"].(map[string]interface{})["data"].(map[string]interface{})["children"].([]interface{})[0].(map[string]interface{})["data"].(map[string]interface{})
	if replyComment["body"] == nil || replyComment["body"].(string) == "[deleted]" {
		return updatedComments
	}
	log.Printf("Leaving ParseCommentData, depth: %d", depth)
	return parseCommentData(replyComment, updatedComments, commentId, depth+1)
}

func parsePostData(data map[string]interface{}) (imageUrl string, linkType c.Base) {
	postData := data["data"].(map[string]interface{})["children"].([]interface{})[0].(map[string]interface{})["data"].(map[string]interface{})

	if postData["is_reddit_media_domain"].(bool) {
		linkType = c.Image
		imageUrl = postData["url_overridden_by_dest"].(string)
	} else {
		if postData["is_self"].(bool) {
			linkType = c.Self
			imageUrl = "self.jpg"
		} else {
			linkType = c.External
			imageUrlParts := strings.Split(postData["url_overridden_by_dest"].(string), "-")
			ext := imageUrlParts[len(imageUrlParts)-1]
			if ext != jpeg && ext != jpg && ext != png {
				imageUrl = postData["thumbnail"].(string)
			} else {
				imageUrl = postData["url_overridden_by_dest"].(string)
			}
		}
	}
	return imageUrl, linkType
}

func parseJsonData(data []map[string]interface{}, commentId string) *c.Link {
	log.Println("Entering ParseJsonData")

	postNode := data[0]
	imageUrl, linkType := parsePostData(postNode)

	commentNode := data[1]
	firstComment := commentNode["data"].(map[string]interface{})["children"].([]interface{})[0].(map[string]interface{})["data"].(map[string]interface{})

	comments := make([]c.Comment, 0, 5)
	comments = parseCommentData(firstComment, comments, commentId, 0)
	log.Println("Leaving ParseJsonData")

	return &c.Link{ImageUrl: imageUrl, RedditComments: comments, LinkType: linkType}
}

func (f FakeRedditCaller) callRedditApi(req c.RedditRequest, user *c.User) (link *c.Link, err error) {
	link, ok := dataaccess.GetLink(req)
	if ok {
		return link, nil
	}

	body, err := os.ReadFile("./static/test4.json")
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
	link.ProxyUrl = GetImgProxyUrl(link.ImageUrl)
	// go addToCache(req, link)
	dataaccess.AddLink(req, link, user.UserId)
	return link, nil
}

func parseApiResponse(res *http.Response, req c.RedditRequest, user *c.User) (link *c.Link, err error) {
	log.Println("Entering ParseApiResponse")
	log.Println(res.StatusCode)
	if res.StatusCode == http.StatusOK {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			log.Printf("error reading response: %s\n", err)
		}
		log.Println("making jsondata")
		jsonData := make(JsonData, 2)
		for i := range jsonData {
			jsonData[i] = make(map[string]interface{})
		}
		log.Println("unmarshalling")
		err = json.Unmarshal(body, &jsonData)
		if err != nil {
			log.Printf("error unmarshalling: %s\n", err)
			return CreateErrorLink(), err
		}
		link = parseJsonData(jsonData, req.Comment)

		link.ProxyUrl = GetImgProxyUrl(link.ImageUrl)
		// go addToCache(req, link)
		dataaccess.AddLink(req, link, user.UserId)
		log.Println("Leaving ParseApiResponse")
		return link, nil
	}
	log.Println("Leaving ParseApiResponse")
	return CreateErrorLink(), nil
}

func (r RealRedditCaller) callRedditApi(req c.RedditRequest, user *c.User) (link *c.Link, err error) {
	link, ok := dataaccess.GetLink(req)
	if ok {
		if link.UserId == user.UserId {
			log.Printf("Pulled from db: %s", req.AsString())
			return link, nil
		} else {
			return nil, fmt.Errorf("trying to access a post that isn't yours, user: %d", user.UserId)
		}
	}

	log.Printf("request: %s\n", req)

	base := "https://oauth.reddit.com"
	subreddit := fmt.Sprintf("r/%s", req.Subreddit)
	article := fmt.Sprintf("comments/%s.json", req.Article)
	comment := fmt.Sprintf("?comment=%s", req.Comment)
	context := fmt.Sprintf("&context=%s", "3")
	limit := fmt.Sprintf("&limit=%s", "5")
	showmedia := fmt.Sprintf("&showmedia=%s", "true")
	// accessToken := fmt.Sprintf("&access_token=%s", redditAccessToken)

	requestUrl, err := url.JoinPath(base, subreddit, article)
	if err != nil {
		log.Println("Error calling reddit api")
	}

	requestUrl = fmt.Sprintf("%s%s%s%s%s", requestUrl, comment, context, limit, showmedia)

	log.Printf("Calling: %s\n", requestUrl)
	// res, err := http.Get(requestUrl)
	// req2, err := http.NewRequest("GET", requestUrl, nil)
	// if err != nil {
	// 	log.Println("Error creating reddit request")
	// }

	// req2.Header = http.Header{
	// 	"user-agent": {"test"},
	// }
	// client := http.Client{}
	dataRequest, err := http.NewRequest("GET", requestUrl, nil)
	dataRequest.Header.Add("Authorization", fmt.Sprintf("Bearer %s", user.AccessToken))
	log.Println(user.AccessToken)

	res, err := client.Do(dataRequest)

	// res, err := http.Get(requestUrl)
	//client.Do(req2)
	if err != nil {
		log.Println("Error calling reddit api")
	}
	defer res.Body.Close()

	return parseApiResponse(res, req, user)
}

func CreateErrorLink() (link *c.Link) {
	comment := c.Comment{Author: "oops", Comment: "brokie"}
	comments := []c.Comment{comment}
	return &c.Link{ImageUrl: "/static/error.webp", ProxyUrl: "/static/error.webp", RedditComments: comments}
}

func GetRedditDetails(req c.RedditRequest, link chan *c.Link, user *c.User) {
	res, err := redditCaller.callRedditApi(req, user)
	if err != nil {
		log.Printf("error making http request: %s\n", err)
		res = CreateErrorLink()
	}
	// log.Println(res)
	// log.Printf("%s", hello)

	// requestURL := fmt.Sprintf("http://localhost:%d/comment?subreddit=%s&article=%s&comment=%s", serverPort, req.Subreddit, req.Article, req.Comment)
	// res, err := http.Get(requestRL)

	link <- res
}

// func GetRedditLink(req c.RedditRequest) *c.Link {
// 	val, ok := linkCache[req]
// 	if ok {
// 		return val
// 	}
// 	return CreateErrorLink()
// }

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	redditAccessToken = os.Getenv("REDDIT_ACCESS_TOKEN")
	CdnBaseUrl = os.Getenv("R2_DEV_URL")

	redditFake := os.Getenv("REDDIT_FAKE")
	if redditFake == "1" {
		redditCaller = &FakeRedditCaller{}
	} else {
		redditCaller = &RealRedditCaller{}
	}

	// linkCache = make(map[c.RedditRequest]c.Link)
	// go populateCache()
}
