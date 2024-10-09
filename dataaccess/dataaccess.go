package dataaccess

import (
	"log"
	c "main/common"
	"os"
	"time"

	"github.com/joho/godotenv"
)

var dataAccessType string

var dataAccess DataAccess

type DataAccess interface {
	GetLinks() map[string]c.Link
	GetRecentLinks(page int) map[string]c.Link
	GetLink(req c.RedditRequest) (*c.Link, bool)
	AddLink(req c.RedditRequest, link *c.Link, userId int)
	UpdateCdnUrl(req c.RedditRequest, cdnUrl string)
	GetUser(username string) (*c.User, bool)
	AddUser(user c.User) bool
	UpdateUser(username string, accessToken string, refreshExpireDtTm time.Time) bool
	DecrementUserUploadCount(userId int) bool
	RefreshUserUploadCount(userId int, newCount int) bool
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

func AddLink(req c.RedditRequest, link *c.Link, userId int) {
	go dataAccess.AddLink(req, link, userId)
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

func UpdateUser(username string, accessToken string, refreshExpireDtTm time.Time) bool {
	return dataAccess.UpdateUser(username, accessToken, refreshExpireDtTm)
}

func DecrementUserUploadCount(userId int) bool {
	return dataAccess.DecrementUserUploadCount(userId)
}

func RefreshUserUploadCount(userId int, newCount int) bool {
	return dataAccess.RefreshUserUploadCount(userId, newCount)
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
