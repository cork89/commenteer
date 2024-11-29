package dataaccess

import (
	"errors"
	c "main/common"
	"time"
)

type Local struct{}

var links map[string]c.Link
var userLinkData map[string]c.UserLinkData
var users map[string]c.User
var usersById map[int]c.User
var linkStylesMap map[int][]c.LinkStyle

func (l Local) GetRecentLinks(page int) []c.Link {
	var recentLinks []c.Link
	recentLinks = make([]c.Link, 0)

	for _, v := range links {
		recentLinks = append(recentLinks, v)
	}

	return recentLinks
}

func (l Local) GetRecentLinksByUsername(page int, username string) ([]c.Link, bool) {
	var recentLinks []c.Link
	recentLinks = make([]c.Link, 0)

	user, ok := users[username]

	if !ok {
		return recentLinks, false
	}

	for _, v := range links {
		if user.UserId == v.UserId {
			recentLinks = append(recentLinks, v)
		}
	}

	return recentLinks, true
}

func (l Local) GetRecentLoggedInLinks(page int, userId int) (links []c.UserLinkData) {
	var recentUserLinkData []c.UserLinkData
	recentUserLinkData = make([]c.UserLinkData, 0)

	for _, v := range userLinkData {
		if userId == v.Link.UserId {
			recentUserLinkData = append(recentUserLinkData, v)
		}
	}

	return recentUserLinkData
}

func (l Local) GetRecentLoggedInSavedLinks(page int, userId int) (links []c.UserLinkData) {
	return l.GetRecentLoggedInLinks(page, userId)
}

func (l Local) GetRecentLoggedInLinksByUsername(page int, userId int, username string) (links []c.UserLinkData) {
	return l.GetRecentLoggedInLinks(page, userId)
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
	userLinkData[req.AsString()] = c.UserLinkData{Link: *link}
}

func (l Local) UpdateCdnUrl(req c.RedditRequest, cdnUrl string, height int, width int) {

}

func (l Local) GetUser(username string) (*c.User, bool) {
	user, ok := users[username]
	return &user, ok
}

func (l Local) AddUser(user c.User) bool {
	users[user.Username] = user
	usersById[user.UserId] = user
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
	if len(linkStyles) < 1 {
		return false
	}
	linkStylesMap[linkStyles[0].LinkId] = linkStyles
	return true
}

func (l Local) GetLinkStyles(linkId int) (linkStyles []c.LinkStyle, err error) {
	linkStyles, ok := linkStylesMap[linkId]

	if !ok {
		return nil, errors.New("no linkstyle")
	}
	return linkStyles, nil
}

func (l Local) InitializeDb() {
	links = make(map[string]c.Link)
	userLinkData = make(map[string]c.UserLinkData)
	users = make(map[string]c.User)
	usersById = make(map[int]c.User)
	linkStylesMap = make(map[int][]c.LinkStyle)
}
