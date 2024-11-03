package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	c "main/common"
	"main/dataaccess"
	d "main/dataaccess"
	"main/middleware"
	"main/snoo"
	s "main/snoo"
	"net/http"
	"os"
	"regexp"
	"strings"
	"text/template"

	"github.com/joho/godotenv"
)

var tmpl map[string]*template.Template

var validPathValue = regexp.MustCompile("^[a-zA-Z0-9_]+-[a-zA-Z0-9]{7}-[a-zA-Z0-9]{7}$")

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
	go d.UpdateCdnUrl(*redditRequest, cdnUrl)
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
	if data.UserId != user.UserId {
		log.Printf("data.UserId: %d, user.UserId: %d", data.UserId, user.UserId)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	singleLinkData.Link = data

	singleLinkData.UserCookie = &user.UserCookie
	singleLinkData.RedditRequest = redditRequest.AsString()

	if err := tmpl["edit"].ExecuteTemplate(w, "base", singleLinkData); err != nil {
		log.Printf("editHandler err: %v", err)
	}
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
	singleLinkData.CommenteerUrl = os.Getenv("COMMENTEER_URL")

	// user, _ := s.GetUserCookie(r)
	if err = tmpl["view"].ExecuteTemplate(w, "base", singleLinkData); err != nil {
		log.Printf("viewHandler err: %v", err)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	var homeData c.HomeData
	homeData.Posts = d.GetRecentLinks(1)

	user, ok := s.GetUserCookie(r)
	if ok {
		homeData.UserCookie = &user.UserCookie
	}
	homeData.CommenteerUrl = os.Getenv("COMMENTEER_URL")

	if err := tmpl["home"].ExecuteTemplate(w, "base", homeData); err != nil {
		log.Printf("homeHandler err: %v", err)
	}
}

func loginPostHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, s.RedditAuthUrl, http.StatusSeeOther)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	state := queryParams.Get("state")
	code := queryParams.Get("code")

	_, ok := s.GetUserCookie(r)
	if !ok {
		if state == "" || code == "" {
			if err := tmpl["login"].ExecuteTemplate(w, "base", nil); err != nil {
				log.Printf("loginHandler state/code err: %v", err)
			}
			return
		}
		accessToken, ok := s.GetRedditAccessToken(state, code)
		if !ok {
			fmt.Println("something went wrong :(")
			if err := tmpl["login"].ExecuteTemplate(w, "base", nil); err != nil {
				log.Printf("loginHandler accesstoken err: %v", err)
			}
			return
		}
		userData := s.GetUserData(*accessToken)
		_, ok = d.GetUser(userData.Username)
		if !ok {
			userAdded := d.AddUser(userData)
			if !userAdded {
				fmt.Println("something went wrong :(")
				if err := tmpl["login"].ExecuteTemplate(w, "base", nil); err != nil {
					log.Printf("loginHandler adduser err: %v", err)
				}
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
	http.Redirect(w, r, "/", http.StatusFound)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("logging out")
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

func imageHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	srcImg := queryParams.Get("src")

	if srcImg == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	destImgUrl := s.ConvertImageToPng(srcImg)

	destImg, err := http.Get(destImgUrl)
	if err != nil {
		fmt.Fprintf(w, "Error %d", err)
		return
	}
	defer destImg.Body.Close()

	w.Header().Set("Content-Length", fmt.Sprint(destImg.ContentLength))
	w.Header().Set("Content-Type", destImg.Header.Get("Content-Type"))
	if _, err = io.Copy(w, destImg.Body); err != nil {
		fmt.Fprintf(w, "Error %d", err)
		return
	}

	// buffer := make([]byte, destImg.ContentLength)
	// destImg.Body.Read(buffer)
	// http.ServeContent(w, r, "test", time.Now(), strings.NewReader(destImg))
}

func faqHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value(c.UserCtx).(*c.User)
	if !ok {
		log.Println("user context missing")
	}

	if err := tmpl["faq"].ExecuteTemplate(w, "base", user); err != nil {
		log.Printf("faqHandler err: %v", err)
	}
}

func likeHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("in like handler")
	redditRequest, err := extractRedditRequest(r)

	if err != nil {
		log.Printf("error getting reddit request, %v\n", err)
		return
	}

	ctx := r.Context()
	user, ok := ctx.Value(c.UserCtx).(*c.User)

	if !ok {
		log.Printf("error getting user, %v\n", err)
		return
	}

	link, ok := d.GetLink(*redditRequest)

	if !ok {
		log.Printf("error getting link, %v\n", err)
		return
	}

	userAction := c.UserAction{UserId: user.UserId, TargetType: c.LinkTarget, ActionType: c.Like, TargetId: link.LinkId}

	ok = d.AddUserAction(userAction)

	if !ok {
		log.Printf("error adding user action, %v\n", err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func LinkWrap(commenteerUrl string, link c.Link) map[string]interface{} {
	return map[string]interface{}{
		"CommenteerUrl": commenteerUrl,
		"Link":          link,
	}
}

func main() {
	dataaccess.Initialize("")
	err := godotenv.Load("/run/secrets/.env.local")
	if err != nil {
		log.Println(err)
	}

	tmpl = make(map[string]*template.Template)

	homeTemp := template.New("home").Funcs(template.FuncMap{"LinkWrap": LinkWrap})
	tmpl["home"] = template.Must(homeTemp.ParseFiles("static/home.html", "static/base.html", "static/linkActions.html"))
	tmpl["edit"] = template.Must(template.ParseFiles("static/edit.html", "static/base.html"))
	viewTemp := template.New("view").Funcs(template.FuncMap{"LinkWrap": LinkWrap})
	tmpl["view"] = template.Must(viewTemp.ParseFiles("static/view.html", "static/base.html", "static/linkActions.html"))
	tmpl["login"] = template.Must(template.ParseFiles("static/login.html", "static/base.html"))
	tmpl["faq"] = template.Must(template.ParseFiles("static/faq.html", "static/base.html"))

	router := http.NewServeMux()

	router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	router.HandleFunc("/", homeHandler)
	router.HandleFunc("GET /login/", loginHandler)
	router.HandleFunc("POST /login/", loginPostHandler)
	loggedInRouter := http.NewServeMux()
	loggedInRouter.HandleFunc("GET /r/{id}/submit/", editHandler)
	loggedInRouter.HandleFunc("POST /r/{id}/submit/", saveHandler)
	router.HandleFunc("/r/{id}/", viewHandler)
	router.HandleFunc("POST /logout/", logoutHandler)
	router.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/robots.txt")
	})
	router.HandleFunc("/image/", imageHandler)
	router.HandleFunc("/faq/", faqHandler)
	router.HandleFunc("POST /r/{id}/like/", likeHandler)

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
