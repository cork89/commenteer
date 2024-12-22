package http

import (
	"errors"
	"fmt"
	c "main/common"
	"net/http"
	"regexp"
	"strings"
	"text/template"
)

type tmplMap struct {
	tmpl map[string]*template.Template
}

var Templates tmplMap

func (t tmplMap) Initialize() tmplMap {
	if t.tmpl != nil {
		return t
	}
	t.tmpl = make(map[string]*template.Template)
	return t
}

func (t tmplMap) Set(name string, template *template.Template) {
	t.tmpl[name] = template
}

func (t tmplMap) Get(name string) *template.Template {
	return t.tmpl[name]
}

type UserState string

const (
	Posts    UserState = "posts"
	Saved    UserState = "saved"
	Settings UserState = "settings"
)

type ErrorType string

const (
	Subreddit ErrorType = "subreddit"
	Other     ErrorType = "other"
)

type MultipleLinkData struct {
	*c.User
	UserLinkData  []c.UserLinkData
	CommenteerUrl string
	ErrorText     string
	ErrorType     ErrorType
	Path          string
	UserState     UserState
}

type SingleLinkData struct {
	*c.User
	c.UserLinkData
	RedditRequest string
	CommenteerUrl string
	Params        string
}

var validPathValue = regexp.MustCompile("^[a-zA-Z0-9_]+-[a-zA-Z0-9]{7}-[a-zA-Z0-9]{7}$")
var validUsername = regexp.MustCompile("^[-_a-zA-Z0-9]{3,20}$")

func ExtractRedditRequest(r *http.Request) (*c.RedditRequest, error) {
	m := validPathValue.FindStringSubmatch(r.PathValue("id"))
	if len(m) == 0 {
		return nil, errors.New("malformed url")
	}
	parts := strings.Split(m[0], "-")
	if len(parts) != 3 {
		return nil, errors.New("invalid url")
	}

	return &c.RedditRequest{Subreddit: parts[0], Article: parts[1], Comment: parts[2]}, nil
}

func ExtractUsername(r *http.Request) (string, error) {
	m := validUsername.FindStringSubmatch(r.PathValue("username"))
	if len(m) != 1 {
		return "", errors.New("invalid username")
	}
	return fmt.Sprintf("u/%s", m[0]), nil
}

func LinkWrap(commenteerUrl string, userLinkData c.UserLinkData, user *c.User) map[string]interface{} {
	return map[string]interface{}{
		"CommenteerUrl": commenteerUrl,
		"UserLinkData":  userLinkData,
		"User":          user,
	}
}

func Initialize(subreddits string) {
	allowedSubreddits = strings.Split(subreddits, "\n")

	Templates = Templates.Initialize()
}
