package http

import (
	"log"
	c "main/common"
	d "main/dataaccess"
	s "main/snoo"
	"net/http"
	"os"
)

func ViewHandler(w http.ResponseWriter, r *http.Request) {
	var singleLinkData SingleLinkData

	var data *c.Link
	var userLinkData *c.UserLinkData

	redditRequest, err := ExtractRedditRequest(r)
	if err != nil {
		log.Printf("error getting filename: %v\n", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data, ok := d.GetLink(*redditRequest)
	if !ok {
		log.Printf("failed to retrieve link for %s", redditRequest.AsString())
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	ctx := r.Context()
	user, ok := ctx.Value(c.UserCtx).(*c.User)
	if !ok {
		user, ok = s.GetUserCookie(r)
	}
	if ok {
		singleLinkData.User = user
	}

	if ok && data.UserId == user.UserId {
		userLinkData, _ = d.GetLoggedInLink(*redditRequest, user.UserId)
	} else {
		userLinkData = &c.UserLinkData{Link: *data}
	}

	singleLinkData.UserLinkData = *userLinkData
	singleLinkData.RedditRequest = redditRequest.AsString()
	singleLinkData.CommenteerUrl = os.Getenv("COMMENTEER_URL")

	if err = Templates.Get("view").ExecuteTemplate(w, "base", singleLinkData); err != nil {
		log.Printf("viewHandler err: %v", err)
	}
}
