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
	Subscribed       bool
	SubscriptionDtTm string
	RefreshToken     string
}

type HomeData struct {
	UserCookie
	Posts map[string]Link
}

func (req RedditRequest) AsString() string {
	return fmt.Sprintf("%s-%s-%s", req.Subreddit, req.Article, req.Comment)
}
