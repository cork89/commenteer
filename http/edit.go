package http

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	c "main/common"
	d "main/dataaccess"
	s "main/snoo"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"
)

var allowedSubreddits []string

type ImageInfo struct {
	ImgData string `json:"imgData"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	Params  string `json:"params"`
}

var imageInfo ImageInfo

type ParamData struct {
	Cmt    string `json:"cmt"`
	Brd    string `json:"brd"`
	Bc     string `json:"bc"`
	Font   string `json:"font"`
	Bold   string `json:"bold"`
	Italic string `json:"italic"`
}

func (pd ParamData) ToJson() (string, error) {
	rslt, err := json.Marshal(pd)
	return string(rslt), err
}

var validRedditDomains = []string{"reddit.com", "www.reddit.com", "old.reddit.com", "new.reddit.com"}

func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	vals := strings.Split(r.URL.String(), "/edit/")
	if !(len(vals) == 2) {
		log.Printf("Malformed redirect url, %v\n", vals)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	redditUrl := vals[1]

	u, err := url.ParseRequestURI(redditUrl)

	if err != nil {
		log.Printf("Invalid redirect url, %v\n", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	tokens := strings.Split(u.String(), "/")
	if len(tokens) < 8 {
		log.Printf("Invalid redirect url, %v\n", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	domain := tokens[1]
	if !slices.Contains(validRedditDomains, domain) {
		log.Printf("Not a reddit url, %v\n", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	subreddit, article, comment := tokens[3], tokens[5], tokens[7]
	redditRequest := c.RedditRequest{Subreddit: subreddit, Article: article, Comment: comment}
	http.Redirect(w, r, fmt.Sprintf("/r/%s/submit/", redditRequest.AsString()), http.StatusFound)
}

func styleHandler(w http.ResponseWriter, r *http.Request, data *c.Link) ParamData {
	query := r.URL.Query()
	before := r.URL.String()
	var params ParamData

	linkStyles, err := d.GetLinkStyles(data.LinkId)
	if err == nil && len(linkStyles) > 0 {
		for _, linkStyle := range linkStyles {
			if !query.Has(linkStyle.Key) {
				query.Set(linkStyle.Key, linkStyle.Value)
			} else {
				continue
			}
			if linkStyle.Key == "cmt" {
				params.Cmt = linkStyle.Value
			} else if linkStyle.Key == "brd" {
				params.Brd = linkStyle.Value
			} else if linkStyle.Key == "bc" {
				params.Bc = linkStyle.Value
			} else if linkStyle.Key == "font" {
				params.Font = linkStyle.Value
			} else if linkStyle.Key == "bold" {
				params.Bold = linkStyle.Value
			} else if linkStyle.Key == "italic" {
				params.Italic = linkStyle.Value
			}
		}
		r.URL.RawQuery = query.Encode()
		after := r.URL.String()
		if before != after {
			http.Redirect(w, r, after, http.StatusFound)
		}
	}
	return params
}

func EditHandler(w http.ResponseWriter, r *http.Request) {
	var singleLinkData SingleLinkData

	redditRequest, err := ExtractRedditRequest(r)

	if err != nil {
		r.Header.Add("ErrorText", "Not a properly formatted reddit link")
		r.Header.Add("ErrorType", "other")
		HomeHandler(w, r)
		return
	}

	if !slices.Contains(allowedSubreddits, redditRequest.Subreddit) {
		r.Header.Add("ErrorText", fmt.Sprintf("r/%s is not a supported subreddit.", redditRequest.Subreddit))
		r.Header.Add("ErrorType", "subreddit")
		HomeHandler(w, r)
		return
	}

	var data *c.Link

	ctx := r.Context()
	user, ok := ctx.Value(c.UserCtx).(*c.User)
	if !ok {
		log.Println("user context missing")
		data = s.CreateErrorLink()
	} else {
		link := make(chan *c.Link)
		go s.GetRedditDetails(*redditRequest, link, user)
		data = <-link
	}
	if data.UserId != user.UserId {
		log.Printf("data.UserId: %d, user.UserId: %d", data.UserId, user.UserId)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	params := styleHandler(w, r, data)

	singleLinkData.UserLinkData = c.UserLinkData{Link: *data}
	singleLinkData.User = user
	singleLinkData.RedditRequest = redditRequest.AsString()
	singleLinkData.CommenteerUrl = os.Getenv("COMMENTEER_URL")
	paramsJson, err := params.ToJson()
	if err != nil {
		log.Printf("failed to marshal edit params, err: %v\n", err)
		singleLinkData.Params = "{}"
	} else {
		singleLinkData.Params = paramsJson
	}

	if err := Templates.Get("edit").ExecuteTemplate(w, "base", singleLinkData); err != nil {
		log.Printf("editHandler err: %v", err)
	}
}

func saveParams(params string, queryId string) bool {
	if len(params) == 0 {
		return false
	}
	minusQ := strings.TrimPrefix(params, "?")
	if minusQ == params {
		return false
	}
	paramList := strings.Split(minusQ, "&")

	linkStyles := make([]c.LinkStyle, 0, len(paramList))

	for _, val := range paramList {
		if !strings.Contains(val, "=") {
			continue
		}
		parameter := strings.Split(val, "=")
		linkStyle := c.LinkStyle{QueryId: queryId, Key: parameter[0], Value: parameter[1]}
		linkStyles = append(linkStyles, linkStyle)
	}
	if len(linkStyles) > 0 {
		ok := d.AddLinkStyles(linkStyles)
		if !ok {
			log.Printf("Failed to save style query params for %s\n", queryId)
			return false
		}
	}
	return true
}

func SaveHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	redditRequest, err := ExtractRedditRequest(r)
	if err != nil {
		log.Printf("error getting filename: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	_, ok := ctx.Value(c.UserCtx).(*c.User)
	if !ok {
		log.Println("user context missing")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	contentType := r.Header.Get("Content-Type")
	contentInfo := strings.Split(contentType, "/")
	if len(contentInfo) < 2 {
		log.Println("content-type malformed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	imageType := contentInfo[1]
	fileName := fmt.Sprintf("%s.%s", redditRequest.AsString(), imageType)

	if err = json.NewDecoder(r.Body).Decode(&imageInfo); err != nil {
		log.Printf("failed to unmarshal request body, err=%v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	go saveParams(imageInfo.Params, redditRequest.AsString())

	b64, err := base64.StdEncoding.DecodeString(imageInfo.ImgData)
	if err != nil {
		log.Printf("failed to decode base64 image, err=%v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	slog.Info("saveHandler", "ContentType", contentType, "imageType", imageType, "filename", fileName)
	imgDataReader := bytes.NewReader(b64)
	_, err = d.UploadImage(imgDataReader, fileName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	cdnUrl := fmt.Sprintf("%s/%s.%s", s.CdnBaseUrl, redditRequest.AsString(), imageType)
	go d.UpdateCdnUrl(*redditRequest, cdnUrl, imageInfo.Height, imageInfo.Width)

	w.Header().Set("Cache-Control", "max-age=0")
	http.Redirect(w, r, fmt.Sprintf("/r/%s", redditRequest.AsString()), http.StatusFound)
}
