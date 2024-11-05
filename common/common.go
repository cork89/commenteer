package common

import (
	"fmt"
	"time"
)

type RedditRequest struct {
	Subreddit string
	Article   string
	Comment   string
}

type Base string

const (
	Image    Base = "Image"
	External Base = "External"
	Self     Base = "Self"
)

type Link struct {
	ImageUrl       string
	ProxyUrl       string
	RedditComments []Comment
	LinkType       Base
	CdnUrl         string
	UserId         int
	QueryId        string
	LinkId         int
}

type UserLinkData struct {
	Link
	*UserAction
}

type Comment struct {
	Comment string
	Author  string
}

type UserCookie struct {
	Username          string    `json:"username"`
	RefreshExpireDtTm time.Time `json:"refreshExpireDtTm"`
	AccessToken       string    `json:"accessToken"`
	IconUrl           string    `json:"icon_url"`
}

type User struct {
	UserCookie
	Subscribed        bool
	SubscriptionDtTm  string
	RefreshToken      string
	UserId            int
	RemainingUploads  int
	UploadRefreshDtTm time.Time
}

type ActionType string

const (
	Like   ActionType = "like"
	Follow ActionType = "follow"
)

type TargetType string

const (
	LinkTarget TargetType = "link"
	UserTarget TargetType = "user"
)

type UserAction struct {
	ActionType ActionType
	TargetType TargetType
	UserId     int
	TargetId   int
	Active     bool
}

type UserActionStatus string

const (
	Created UserActionStatus = "Created"
	Updated UserActionStatus = "Updated"
)

type HomeData struct {
	*UserCookie
	UserLinkData  []UserLinkData
	CommenteerUrl string
}

type SingleLinkData struct {
	*UserCookie
	UserLinkData
	RedditRequest string
	CommenteerUrl string
}

type HttpContext string

const (
	UserCtx HttpContext = "UserCtx"
)

type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func (RealClock) Now() time.Time {
	return time.Now()
}

func (req RedditRequest) AsString() string {
	return fmt.Sprintf("%s-%s-%s", req.Subreddit, req.Article, req.Comment)
}

func (link Link) AsString() string {
	return fmt.Sprintf("Image Url: %s\nProxy Url: %s\nLinkType: %s\nCdnUrl: %s\nUserId: %d\nRedditComments:%s\n", link.ImageUrl, link.ProxyUrl, link.LinkType, link.CdnUrl, link.UserId, prettyPrintComments(link.RedditComments))
}

func (c Comment) AsString() string {
	return fmt.Sprintf("\tComment: %s\n\tUser: %s\n", c.Comment, c.Author)
}

func prettyPrintComments(cmts []Comment) string {
	final := "[\n"
	for _, cmt := range cmts {
		final += cmt.AsString()
	}
	final += "]"
	return final
}
