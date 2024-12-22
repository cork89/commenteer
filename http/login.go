package http

import (
	"log"
	d "main/dataaccess"
	s "main/snoo"
	"net/http"
)

func LoginPostHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, s.RedditAuthUrl, http.StatusSeeOther)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	state := queryParams.Get("state")
	code := queryParams.Get("code")

	_, ok := s.GetUserCookie(r)
	if !ok {
		if state == "" || code == "" {
			if err := Templates.Get("login").ExecuteTemplate(w, "base", nil); err != nil {
				log.Printf("loginHandler state/code err: %v", err)
			}
			return
		}
		accessToken, ok := s.GetRedditAccessToken(state, code)
		if !ok {
			log.Println("something went wrong :(")
			if err := Templates.Get("login").ExecuteTemplate(w, "base", nil); err != nil {
				log.Printf("loginHandler accesstoken err: %v", err)
			}
			return
		}
		userData := s.GetUserData(*accessToken)
		_, ok = d.GetUser(userData.Username)
		if !ok {
			userAdded := d.AddUser(userData)
			if !userAdded {
				log.Println("something went wrong :(")
				if err := Templates.Get("login").ExecuteTemplate(w, "base", nil); err != nil {
					log.Printf("loginHandler adduser err: %v", err)
				}
				return
			}
		}
		cookie := s.CreateUserCookie(userData.UserCookie)
		http.SetCookie(w, &cookie)
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("logging out")
	c := &http.Cookie{
		Name:     s.CookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, c)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
