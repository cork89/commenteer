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
	ImageUrl       string    `json:"imageUrl"`
	ProxyUrl       string    `json:"proxyUrl"`
	RedditComments []Comment `json:"redditComments,omitempty"`
	LinkType       Base      `json:"linkType"`
	CdnUrl         string    `json:"cdnUrl"`
	UserId         int       `json:"userId"`
	QueryId        string    `json:"queryId"`
	LinkId         int       `json:"id"`
	ImageWidth     int       `json:"imageWidth"`
	ImageHeight    int       `json:"imageHeight"`
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
	AccessToken       string    `json:"accessToken,omitempty"`
	IconUrl           string    `json:"iconUrl,omitempty"`
}

type User struct {
	UserCookie
	Subscribed        bool      `json:"subscribed"`
	SubscriptionDtTm  string    `json:"subscriptionDtTm"`
	RefreshToken      string    `json:"refreshToken,omitempty"`
	UserId            int       `json:"userId"`
	RemainingUploads  int       `json:"remainingUploads"`
	UploadRefreshDtTm time.Time `json:"uploadRefreshDtTm"`
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
	ActionType ActionType `json:"actionType"`
	TargetType TargetType `json:"targetType"`
	UserId     int        `json:"userId"`
	TargetId   int        `json:"targetId"`
	Active     bool       `json:"active"`
}

type UserActionStatus string

const (
	Created UserActionStatus = "Created"
	Updated UserActionStatus = "Updated"
)

type LinkStyle struct {
	LinkStyleId int
	LinkId      int
	QueryId     string
	Key         string
	Value       string
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
