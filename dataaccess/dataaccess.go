package dataaccess

import (
	"log"
	c "main/common"
	"os"

	"github.com/joho/godotenv"
)

var dataAccessType string

var dataAccess DataAccess

type DataAccess interface {
	GetLinks() map[string]c.Link
	GetRecentLinks(page int) map[string]c.Link
	GetLink(req c.RedditRequest) (*c.Link, bool)
	AddLink(req c.RedditRequest, link *c.Link)
	UpdateCdnUrl(req c.RedditRequest, cdnUrl string)
	GetUser(username string) (*c.User, bool)
	AddUser(user c.User) bool
}

func GetLinks() (linkJson map[string]c.Link) {
	return dataAccess.GetLinks()
}

func GetRecentLinks(page int) (linkJson map[string]c.Link) {
	return dataAccess.GetRecentLinks(page)
}

func GetLink(req c.RedditRequest) (*c.Link, bool) {
	return dataAccess.GetLink(req)
}

func AddLink(req c.RedditRequest, link *c.Link) {
	go dataAccess.AddLink(req, link)
}

func UpdateCdnUrl(req c.RedditRequest, cdnUrl string) {
	go dataAccess.UpdateCdnUrl(req, cdnUrl)
}

func GetUser(username string) (*c.User, bool) {
	return dataAccess.GetUser(username)
}

func AddUser(user c.User) bool {
	return dataAccess.AddUser(user)
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	dataAccessType = os.Getenv("DATA_ACCESS_TYPE")

	if dataAccessType == "local" {
		dataAccess = &Local{}
	} else {
		dataAccess = &Db{}
	}
}
