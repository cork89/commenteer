package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"log/slog"
	c "main/common"
	d "main/dataaccess"
	"main/middleware"
	"main/snoo"
	s "main/snoo"
	"net/http"
	"regexp"
	"strings"
	"text/template"
)

var tmpl map[string]*template.Template

var validPathValue = regexp.MustCompile("^[a-zA-Z0-9]+-[a-zA-Z0-9]{7}-[a-zA-Z0-9]{7}$")

func extractRedditRequest(r *http.Request) (*c.RedditRequest, error) {
	m := validPathValue.FindStringSubmatch(r.PathValue("id"))
	log.Println("m: ", m)
	parts := strings.Split(m[0], "-")
	if len(parts) != 3 {
		return nil, errors.New("invalid url")
	}

	return &c.RedditRequest{Subreddit: parts[0], Article: parts[1], Comment: parts[2]}, nil
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	redditRequest, err := extractRedditRequest(r)
	if err != nil {
		log.Printf("error getting filename: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	user, ok := ctx.Value(c.UserCtx).(*c.User)
	if !ok {
		log.Println("user context missing")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	contentType := r.Header.Get("Content-Type")
	imageType := strings.Split(contentType, "/")[1]
	fileName := fmt.Sprintf("%s.%s", redditRequest.AsString(), imageType)

	b64 := base64.NewDecoder(base64.StdEncoding, r.Body)
	// fmt.Println(contentType, imageType, fileName)
	slog.Info("saveHandler", "ContentType", contentType, "imageType", imageType, "filename", fileName)
	// fmt.Println(r.Body, b64)
	// w.WriteHeader(http.StatusInternalServerError)
	// return

	_, err = d.UploadImage(b64, fileName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	cdnUrl := fmt.Sprintf("%s/%s.%s", s.CdnBaseUrl, redditRequest.AsString(), imageType)
	d.UpdateCdnUrl(*redditRequest, cdnUrl)
	go snoo.DecrementUserUploadCount(user)

	http.Redirect(w, r, fmt.Sprintf("/r/%s", redditRequest.AsString()), http.StatusSeeOther)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	var singleLinkData c.SingleLinkData

	redditRequest, err := extractRedditRequest(r)
	var data *c.Link

	ctx := r.Context()
	user, ok := ctx.Value(c.UserCtx).(*c.User)
	if !ok {
		log.Println("user context missing")
		return
	}

	if err != nil {
		log.Println(err)
		data = s.CreateErrorLink()
	} else {
		link := make(chan *c.Link)
		go s.GetRedditDetails(*redditRequest, link, user)
		data = <-link
	}
	singleLinkData.Link = data

	singleLinkData.UserCookie = &user.UserCookie
	singleLinkData.RedditRequest = redditRequest.AsString()

	tmpl["edit"].ExecuteTemplate(w, "base", singleLinkData)
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	var singleLinkData c.SingleLinkData
	redditRequest, err := extractRedditRequest(r)

	var data *c.Link

	ctx := r.Context()
	user, ok := ctx.Value(c.UserCtx).(*c.User)
	if !ok {
		log.Println("user context missing")
	} else {
		singleLinkData.UserCookie = &user.UserCookie
	}

	if err != nil {
		log.Println(err)
		data = s.CreateErrorLink()
	} else {
		link := make(chan *c.Link)
		go s.GetRedditDetails(*redditRequest, link, user)
		data = <-link
	}

	singleLinkData.Link = data
	singleLinkData.RedditRequest = redditRequest.AsString()

	// user, _ := s.GetUserCookie(r)

	tmpl["view"].ExecuteTemplate(w, "base", singleLinkData)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	var homeData c.HomeData
	currentPosts := d.GetRecentLinks(1)
	homeData.Posts = currentPosts

	user, ok := s.GetUserCookie(r)
	if ok {
		homeData.UserCookie = &user.UserCookie
	}
	tmpl["home"].ExecuteTemplate(w, "base", homeData)
}

func loginPostHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, s.RedditAuthUrl, http.StatusSeeOther)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	test := "hello world"
	fmt.Println(r.Header)
	fmt.Println(r.Body)
	fmt.Println(r.URL.Query())
	queryParams := r.URL.Query()

	state := queryParams.Get("state")
	code := queryParams.Get("code")

	_, ok := s.GetUserCookie(r)
	if !ok {
		if state == "" || code == "" {
			tmpl["login"].ExecuteTemplate(w, "base", test)
			return
		}
		accessToken, ok := s.GetRedditAccessToken(state, code)
		if !ok {
			test = "something went wrong :("
			tmpl["login"].ExecuteTemplate(w, "base", test)
			return
		}
		userData := s.GetUserData(*accessToken)
		_, ok = d.GetUser(userData.Username)
		if !ok {
			userAdded := d.AddUser(userData)
			if !userAdded {
				test = "something went wrong :("
				tmpl["login"].ExecuteTemplate(w, "base", test)
				return
			}
		}
		cookie := s.CreateUserCookie(userData.UserCookie)
		http.SetCookie(w, &cookie)
	}

	// if !ok {
	// 	accessToken := s.GetRedditAccessToken(state, code)

	// 	cookie := createCookie(accessToken)
	// 	http.SetCookie(w, &cookie)
	// }

	// if userCookie.RefreshExpireDtTm < time.Now().Format(time.RFC3339) {
	// 	user := d.GetUser(userCookie.Username)

	// }

	// tmpl["home"].ExecuteTemplate(w, "base")
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func main() {
	// snoo.Main()
	tmpl = make(map[string]*template.Template)
	tmpl["home"] = template.Must(template.ParseFiles("static/home.html", "static/base.html"))
	tmpl["edit"] = template.Must(template.ParseFiles("static/edit.html", "static/base.html"))
	tmpl["view"] = template.Must(template.ParseFiles("static/view.html", "static/base.html"))
	tmpl["comments"] = template.Must(template.ParseFiles("static/comments.html", "static/base.html"))
	tmpl["login"] = template.Must(template.ParseFiles("static/login.html", "static/base.html"))

	router := http.NewServeMux()

	router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	router.HandleFunc("/", homeHandler)
	router.HandleFunc("GET /login/", loginHandler)
	router.HandleFunc("POST /login/", loginPostHandler)
	loggedInRouter := http.NewServeMux()
	loggedInRouter.HandleFunc("GET /r/{id}/submit/", editHandler)
	loggedInRouter.HandleFunc("POST /r/{id}/submit/", saveHandler)
	router.HandleFunc("/r/{id}/", viewHandler)

	stack := middleware.CreateStack(
		middleware.Logging,
		middleware.CacheControl,
		middleware.IsLoggedIn,
	)

	strict := middleware.CreateStack(
		middleware.CheckRemainingUploads,
		middleware.IsLoggedInStrict,
	)

	router.Handle("/r/{id}/submit/", strict(loggedInRouter))

	server := http.Server{
		Addr:    ":8090",
		Handler: stack(router),
	}

	server.ListenAndServe()
}
