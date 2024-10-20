package dataaccess

import (
	c "main/common"
	"time"
)

type Local struct{}

var links map[string]c.Link
var users map[string]c.User

func (l Local) GetRecentLinks(page int) map[string]c.Link {
	return links
}

func (l Local) GetLinks() map[string]c.Link {
	return links
}

func (l Local) GetLink(req c.RedditRequest) (*c.Link, bool) {
	link, ok := links[req.AsString()]
	return &link, ok
}

func (l Local) AddLink(req c.RedditRequest, link *c.Link, userId int) {
	links[req.AsString()] = *link
}

func (l Local) UpdateCdnUrl(req c.RedditRequest, cdnUrl string) {

}

func (l Local) GetUser(username string) (*c.User, bool) {
	user, ok := users[username]
	return &user, ok
}

func (l Local) AddUser(user c.User) bool {
	users[user.Username] = user
	return true
}

func (l Local) UpdateUser(username string, accessToken string, refreshExpireDtTm time.Time) bool {
	return true
}

func (l Local) DecrementUserUploadCount(userId int) bool {
	return true
}

func (l Local) RefreshUserUploadCount(userId int, newCount int) bool {
	return true
}

func init() {
	links = make(map[string]c.Link)
	users = make(map[string]c.User)
}
