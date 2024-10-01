package dataaccess

import (
	"encoding/json"
	"log"
	c "main/common"
	"os"
)

type Local struct{}

func (l Local) GetRecentLinks(page int) (linkJson map[string]c.Link) {
	return nil
}

func (l Local) GetLinks() (linkJson map[string]c.Link) {
	f, err := os.ReadFile("./static/cache.json")
	if err != nil {
		log.Println(err)
	}

	err = json.Unmarshal(f, &linkJson)
	if err != nil {
		log.Println(err)
	}
	return linkJson
}

func (l Local) GetLink(req c.RedditRequest) (*c.Link, bool) {
	return nil, false
}

func (l Local) AddLink(req c.RedditRequest, link *c.Link) {
	f, err := os.ReadFile("./static/cache.json")

	if err != nil {
		log.Println(err)
	}
	var linkJson map[string]*c.Link
	err = json.Unmarshal(f, &linkJson)
	if err != nil {
		log.Println(err)
	}
	smallReq := req.AsString()

	_, ok := linkJson[smallReq]
	if !ok {
		linkJson[smallReq] = link
	}
	fileJson, err := json.Marshal(linkJson)
	if err != nil {
		log.Println(err)
	}

	err = os.WriteFile("./static/cache.json", fileJson, 0644)
	if err != nil {
		log.Println(err)
	}
}

func (l Local) UpdateCdnUrl(req c.RedditRequest, cdnUrl string) {

}

func (l Local) GetUser(username string) (*c.User, bool) {
	return &c.User{UserCookie: c.UserCookie{Username: username}}, true
}

func (l Local) AddUser(user c.User) bool {
	return true
}
