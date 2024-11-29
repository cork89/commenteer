package http

import (
	"log"
	c "main/common"
	d "main/dataaccess"
	s "main/snoo"
	"net/http"
	"os"
)

func GetUserLinks(user *c.User, userLoggedIn bool, username string) (multipleLinkData MultipleLinkData, redirect bool) {
	userLinkData := make([]c.UserLinkData, 0)

	if userLoggedIn {
		multipleLinkData.User = user
	}
	if userLoggedIn && user.Username == username {
		userLinkData = d.GetRecentLoggedInLinksByUsername(1, user.UserId, username)
	} else {
		posts, userExists := d.GetRecentLinksByUsername(1, username)
		if !userExists {
			multipleLinkData.ErrorText = "user not found"
		} else {
			for _, post := range posts {
				userLinkDataItem := c.UserLinkData{Link: post}
				userLinkData = append(userLinkData, userLinkDataItem)
			}
		}
	}
	multipleLinkData.UserLinkData = userLinkData
	multipleLinkData.UserState = Posts
	return multipleLinkData, false
}

func GetUserSavedLinks(user *c.User, userLoggedIn bool, username string) (multipleLinkData MultipleLinkData, redirect bool) {
	userLinkData := make([]c.UserLinkData, 0)

	if userLoggedIn && user.Username == username {
		multipleLinkData.User = user
		userLinkData = d.GetRecentLoggedInSavedLinks(1, user.UserId)
	} else {
		redirect = true
	}
	multipleLinkData.UserLinkData = userLinkData
	multipleLinkData.UserState = Saved
	return multipleLinkData, redirect
}

func GetUserSettings(user *c.User, userLoggedIn bool, username string) (multipleLinkData MultipleLinkData, redirect bool) {
	if userLoggedIn && user.Username == username {
		multipleLinkData.User = user
		multipleLinkData.ErrorText = "todo"
	} else {
		redirect = true
	}
	multipleLinkData.UserState = Settings
	return multipleLinkData, redirect

}

func UserHandler(w http.ResponseWriter, r *http.Request, linkRetriever func(user *c.User, userLoggedIn bool, username string) (multipleLinkData MultipleLinkData, redirect bool)) {
	var multipleLinkData MultipleLinkData
	username, err := ExtractUsername(r)

	if err != nil {
		log.Printf("error extracting username, err=%v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	user, ok := ctx.Value(c.UserCtx).(*c.User)
	if !ok {
		user, ok = s.GetUserCookie(r)
	}

	multipleLinkData, redirect := linkRetriever(user, ok, username)
	if redirect {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	multipleLinkData.Path = username
	multipleLinkData.CommenteerUrl = os.Getenv("COMMENTEER_URL")

	if err := Templates.Get("user").ExecuteTemplate(w, "base", multipleLinkData); err != nil {
		log.Printf("userHandler err: %v", err)
	}
}
