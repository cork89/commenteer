package http

import (
	"encoding/json"
	"fmt"
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
	multipleLinkData, redirect, err := UserDataHandler(w, r, linkRetriever)

	if err != nil {
		return
	}

	if redirect {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if err := Templates.Get("user").ExecuteTemplate(w, "base", multipleLinkData); err != nil {
		log.Printf("userHandler err: %v", err)
	}
}

func UserDataHandler(w http.ResponseWriter, r *http.Request, linkRetriever func(user *c.User, userLoggedIn bool, username string) (multipleLinkData MultipleLinkData, redirect bool)) (MultipleLinkData, bool, error) {
	username, err := ExtractUsername(r)

	var multipleLinkData MultipleLinkData
	var redirect bool

	if err != nil {
		log.Printf("error extracting username, err=%v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return multipleLinkData, redirect, err
	}

	ctx := r.Context()
	user, ok := ctx.Value(c.UserCtx).(*c.User)
	if !ok {
		user, ok = s.GetUserCookie(r)
	}
	multipleLinkData, redirect = linkRetriever(user, ok, username)
	multipleLinkData.Path = username
	multipleLinkData.CommenteerUrl = os.Getenv("COMMENTEER_URL")
	return multipleLinkData, redirect, nil
}

func UserDataApiHandler(w http.ResponseWriter, r *http.Request, linkRetriever func(user *c.User, userLoggedIn bool, username string) (multipleLinkData MultipleLinkData, redirect bool)) {
	multipleLinkData, redirect, err := UserDataHandler(w, r, linkRetriever)
	fmt.Println(multipleLinkData, redirect, err)
	if err != nil || redirect {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	data, err := json.Marshal(multipleLinkData)

	fmt.Println(string(data))

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}
