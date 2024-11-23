package main

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	c "main/common"
	d "main/dataaccess"
	"main/middleware"
	s "main/snoo"
	"net/http"
	"os"
	"regexp"
	"slices"
	"strings"
	"text/template"

	"github.com/joho/godotenv"
)

var tmpl map[string]*template.Template

//go:embed static/AllowedSubreddits.txt
var subreddits string
var allowedSubreddits []string

var validPathValue = regexp.MustCompile("^[a-zA-Z0-9_]+-[a-zA-Z0-9]{7}-[a-zA-Z0-9]{7}$")
var validUsername = regexp.MustCompile("^[-_a-zA-Z0-9]{3,20}$")

func extractRedditRequest(r *http.Request) (*c.RedditRequest, error) {
	m := validPathValue.FindStringSubmatch(r.PathValue("id"))
	log.Println("m: ", m)
	parts := strings.Split(m[0], "-")
	if len(parts) != 3 {
		return nil, errors.New("invalid url")
	}

	return &c.RedditRequest{Subreddit: parts[0], Article: parts[1], Comment: parts[2]}, nil
}

func extractUsername(r *http.Request) (string, error) {
	m := validUsername.FindStringSubmatch(r.PathValue("username"))
	log.Println("m: ", m)
	if len(m) != 1 {
		return "", errors.New("invalid username")
	}
	return fmt.Sprintf("u/%s", m[0]), nil
}

var imageInfo struct {
	ImgData string `json:"imgData"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
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
	_, ok := ctx.Value(c.UserCtx).(*c.User)
	if !ok {
		log.Println("user context missing")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	contentType := r.Header.Get("Content-Type")
	imageType := strings.Split(contentType, "/")[1]
	fileName := fmt.Sprintf("%s.%s", redditRequest.AsString(), imageType)

	if err = json.NewDecoder(r.Body).Decode(&imageInfo); err != nil {
		log.Printf("failed to unmarshal request body, err=%v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

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

func editHandler(w http.ResponseWriter, r *http.Request) {
	var singleLinkData c.SingleLinkData

	redditRequest, err := extractRedditRequest(r)

	if !slices.Contains(allowedSubreddits, redditRequest.Subreddit) {
		r.Header.Add("ErrorText", fmt.Sprintf("r/%s is not a supported subreddit.", redditRequest.Subreddit))
		homeHandler(w, r)
		return
	}

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

	singleLinkData.UserLinkData = c.UserLinkData{Link: *data}
	singleLinkData.User = user
	singleLinkData.RedditRequest = redditRequest.AsString()
	singleLinkData.CommenteerUrl = os.Getenv("COMMENTEER_URL")

	if err := tmpl["edit"].ExecuteTemplate(w, "base", singleLinkData); err != nil {
		log.Printf("editHandler err: %v", err)
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	var singleLinkData c.SingleLinkData

	var data *c.Link
	var userLinkData *c.UserLinkData

	ctx := r.Context()
	user, ok := ctx.Value(c.UserCtx).(*c.User)
	if !ok {
		log.Println("user context missing")
	}

	redditRequest, err := extractRedditRequest(r)
	if err != nil {
		log.Println(err)
		data = s.CreateErrorLink()
	} else {
		link := make(chan *c.Link)
		go s.GetRedditDetails(*redditRequest, link, user)
		data = <-link
	}

	if ok {
		singleLinkData.User = user
		userLinkData, ok = d.GetLoggedInLink(*redditRequest, user.UserId)
		if !ok {
			userLinkData = &c.UserLinkData{Link: *data}
		}
	} else {
		userLinkData = &c.UserLinkData{Link: *data}
	}

	singleLinkData.UserLinkData = *userLinkData
	singleLinkData.RedditRequest = redditRequest.AsString()
	singleLinkData.CommenteerUrl = os.Getenv("COMMENTEER_URL")

	// user, _ := s.GetUserCookie(r)
	if err = tmpl["view"].ExecuteTemplate(w, "base", singleLinkData); err != nil {
		log.Printf("viewHandler err: %v", err)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	var multipleLinkData c.MultipleLinkData
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

	if err := tmpl["home"].ExecuteTemplate(w, "base", multipleLinkData); err != nil {
		log.Printf("homeHandler err: %v", err)
	}
}

func getUserLinks(user *c.User, userLoggedIn bool, username string) (multipleLinkData c.MultipleLinkData) {
	userLinkData := make([]c.UserLinkData, 0)

	if userLoggedIn {
		multipleLinkData.User = user
	}
	if userLoggedIn && user.Username == username {
		userLinkData = d.GetRecentLoggedInLinks(1, user.UserId)
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
	multipleLinkData.UserState = c.Posts
	return multipleLinkData
}

func getUserSavedLinks(user *c.User, userLoggedIn bool, username string) (multipleLinkData c.MultipleLinkData) {
	userLinkData := make([]c.UserLinkData, 0)

	if userLoggedIn && user.Username == username {
		multipleLinkData.User = user
		userLinkData = d.GetRecentLoggedInSavedLinks(1, user.UserId)
	}
	multipleLinkData.UserLinkData = userLinkData
	multipleLinkData.UserState = c.Saved
	return multipleLinkData
}

func getUserSettings(user *c.User, userLoggedIn bool, username string) (multipleLinkData c.MultipleLinkData) {
	if userLoggedIn && user.Username == username {
		multipleLinkData.User = user
		multipleLinkData.ErrorText = "todo"
	}
	multipleLinkData.UserState = c.Settings
	return multipleLinkData

}

func userHandler(w http.ResponseWriter, r *http.Request, linkRetriever func(user *c.User, userLoggedIn bool, username string) (multipleLinkData c.MultipleLinkData)) {
	var multipleLinkData c.MultipleLinkData
	username, err := extractUsername(r)

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

	if ok && user.Username != username {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	multipleLinkData = linkRetriever(user, ok, username)
	multipleLinkData.Path = username
	multipleLinkData.CommenteerUrl = os.Getenv("COMMENTEER_URL")

	if err := tmpl["user"].ExecuteTemplate(w, "base", multipleLinkData); err != nil {
		log.Printf("userHandler err: %v", err)
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

func LinkWrap(commenteerUrl string, userLinkData c.UserLinkData, user *c.User) map[string]interface{} {
	return map[string]interface{}{
		"CommenteerUrl": commenteerUrl,
		"UserLinkData":  userLinkData,
		"User":          user,
	}
}

func main() {
	err := godotenv.Load("/run/secrets/.env.local")
	if err != nil {
		log.Println(err)
	}

	allowedSubreddits = strings.Split(subreddits, "\n")

	tmpl = make(map[string]*template.Template)

	homeTemp := template.New("home").Funcs(template.FuncMap{"LinkWrap": LinkWrap})
	tmpl["home"] = template.Must(homeTemp.ParseFiles("static/home.html", "static/base.html", "static/links.html", "static/linkActions.html"))
	tmpl["edit"] = template.Must(template.ParseFiles("static/edit.html", "static/base.html"))
	viewTemp := template.New("view").Funcs(template.FuncMap{"LinkWrap": LinkWrap})
	tmpl["view"] = template.Must(viewTemp.ParseFiles("static/view.html", "static/base.html", "static/linkActions.html"))
	tmpl["login"] = template.Must(template.ParseFiles("static/login.html", "static/base.html"))
	tmpl["faq"] = template.Must(template.ParseFiles("static/faq.html", "static/base.html"))
	userTemp := template.New("home").Funcs(template.FuncMap{"LinkWrap": LinkWrap})
	tmpl["user"] = template.Must(userTemp.ParseFiles("static/user.html", "static/base.html", "static/links.html", "static/linkActions.html"))
	tmpl["userSaved"] = template.Must(userTemp.ParseFiles("static/user.html", "static/base.html", "static/links.html", "static/linkActions.html"))

	router := http.NewServeMux()

	router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	router.HandleFunc("/", homeHandler)
	router.HandleFunc("GET /login/", loginHandler)
	router.HandleFunc("POST /login/", loginPostHandler)
	router.HandleFunc("GET /u/{username}/", func(w http.ResponseWriter, r *http.Request) {
		userHandler(w, r, getUserLinks)
	})
	router.HandleFunc("POST /u/{username}/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
	})
	loggedInRouter := http.NewServeMux()
	loggedInRouter.HandleFunc("GET /r/{id}/submit/", editHandler)
	loggedInRouter.HandleFunc("POST /r/{id}/submit/", saveHandler)
	loggedInRouter.HandleFunc("GET /u/{username}/saved/", func(w http.ResponseWriter, r *http.Request) {
		userHandler(w, r, getUserSavedLinks)
	})
	loggedInRouter.HandleFunc("POST /u/{username}/saved/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
	})
	loggedInRouter.HandleFunc("GET /u/{username}/settings/", func(w http.ResponseWriter, r *http.Request) {
		userHandler(w, r, getUserSettings)
	})
	loggedInRouter.HandleFunc("POST /u/{username}/settings/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
	})
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
	router.Handle("GET /u/{username}/saved/", strict(loggedInRouter))
	router.Handle("POST /u/{username}/saved/", strict(loggedInRouter))
	router.Handle("GET /u/{username}/settings/", strict(loggedInRouter))
	router.Handle("POST /u/{username}/settings/", strict(loggedInRouter))

	server := http.Server{
		Addr:    ":8090",
		Handler: stack(router),
	}

	server.ListenAndServe()
}
