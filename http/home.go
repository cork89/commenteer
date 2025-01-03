package http

import (
	"encoding/json"
	"log"
	c "main/common"
	d "main/dataaccess"
	s "main/snoo"
	"net/http"
	"os"
	"strconv"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	var multipleLinkData MultipleLinkData
	var userLinkData []c.UserLinkData
	var posts []c.Link

	userLinkData = make([]c.UserLinkData, 0, len(posts))

	ctx := r.Context()
	user, ok := ctx.Value(c.UserCtx).(*c.User)
	if !ok {
		user, ok = s.GetUserCookie(r)
	}
	if ok {
		multipleLinkData.User = user
		userLinkData = d.GetRecentLoggedInLinks(1, user.UserId)
	} else {
		posts = d.GetRecentLinks(1)
		for _, post := range posts {
			userLinkDataItem := c.UserLinkData{Link: post}
			userLinkData = append(userLinkData, userLinkDataItem)
		}
	}

	multipleLinkData.UserLinkData = userLinkData
	multipleLinkData.CommenteerUrl = os.Getenv("COMMENTEER_URL")
	multipleLinkData.ErrorText = r.Header.Get("ErrorText")
	template := Templates.Get("home")
	if err := template.ExecuteTemplate(w, "base", multipleLinkData); err != nil {
		log.Printf("homeHandler err: %v", err)
	}
}

func HomeDataHandler(w http.ResponseWriter, r *http.Request) {
	var multipleLinkData MultipleLinkData = HomeDataRetriever(w, r)

	template := Templates.Get("homedata")
	if err := template.ExecuteTemplate(w, "links", multipleLinkData); err != nil {
		log.Printf("homeMoreDataHandler err: %v", err)
	}
}

func HomeDataApiHandler(w http.ResponseWriter, r *http.Request) {
	var multipleLinkData MultipleLinkData = HomeDataRetriever(w, r)
	if multipleLinkData.User != nil {
		multipleLinkData.User.AccessToken = ""
		multipleLinkData.User.IconUrl = ""
		multipleLinkData.User.RefreshToken = ""
	}

	data, err := json.Marshal(multipleLinkData)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

func HomeDataRetriever(w http.ResponseWriter, r *http.Request) MultipleLinkData {
	var multipleLinkData MultipleLinkData
	var userLinkData []c.UserLinkData
	var posts []c.Link

	userLinkData = make([]c.UserLinkData, 0)

	offset := r.URL.Query().Get("offset")
	pageNum, err := strconv.Atoi(offset)
	if err != nil {
		pageNum = 1
	}

	ctx := r.Context()
	user, ok := ctx.Value(c.UserCtx).(*c.User)
	if !ok {
		user, ok = s.GetUserCookie(r)
	}
	if ok {
		multipleLinkData.User = user
		userLinkData = d.GetRecentLoggedInLinks(pageNum, user.UserId)
	} else {
		posts = d.GetRecentLinks(pageNum)
		for _, post := range posts {
			userLinkDataItem := c.UserLinkData{Link: post}
			userLinkData = append(userLinkData, userLinkDataItem)
		}
	}

	multipleLinkData.UserLinkData = userLinkData
	multipleLinkData.CommenteerUrl = os.Getenv("COMMENTEER_URL")
	return multipleLinkData
}
