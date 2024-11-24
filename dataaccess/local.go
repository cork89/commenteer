package dataaccess

import (
	c "main/common"
	"time"
)

type Local struct{}

var links map[string]c.Link
var users map[string]c.User

func (l Local) GetRecentLinks(page int) []c.Link {
	return make([]c.Link, 0)
}

func (l Local) GetRecentLinksByUsername(page int, username string) ([]c.Link, bool) {
	return make([]c.Link, 0), true
}

func (l Local) GetRecentLoggedInLinks(page int, userId int) (links []c.UserLinkData) {
	return make([]c.UserLinkData, 0)
}

func (l Local) GetRecentLoggedInSavedLinks(page int, userId int) (links []c.UserLinkData) {
	return make([]c.UserLinkData, 0)
}

func (l Local) GetRecentLoggedInLinksByUsername(page int, userId int, username string) (links []c.UserLinkData) {
	return make([]c.UserLinkData, 0)
}

func (l Local) GetLinks() map[string]c.Link {
	return links
}

func (l Local) GetLink(req c.RedditRequest) (*c.Link, bool) {
	link, ok := links[req.AsString()]
	return &link, ok
}

func (l Local) GetLoggedInLink(req c.RedditRequest, userId int) (*c.UserLinkData, bool) {
	return nil, false
}

func (l Local) AddLink(req c.RedditRequest, link *c.Link, userId int) {
	links[req.AsString()] = *link
}

func (l Local) UpdateCdnUrl(req c.RedditRequest, cdnUrl string, height int, width int) {

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

func (l Local) AddUserAction(userAction c.UserAction) bool {
	return true
}

func (l Local) UpdateUserActionActive(userAction c.UserAction) bool {
	return true
}

func (l Local) AddLinkStyles(linkStyles []c.LinkStyle) bool {
	return true
}

func (l Local) GetLinkStyles(linkId int) (linkStyles []c.LinkStyle, err error) {
	return make([]c.LinkStyle, 0), nil
}

func init() {
	links = make(map[string]c.Link)
	users = make(map[string]c.User)
}
