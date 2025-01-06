package main

import (
	_ "embed"
	"log"
	h "main/http"
	"main/middleware"
	"net/http"
	"text/template"

	"github.com/joho/godotenv"
)

//go:embed static/AllowedSubreddits.txt
var subreddits string

func main() {
	err := godotenv.Load("/run/secrets/.env.local")
	if err != nil {
		log.Println(err)
	}

	h.Initialize(subreddits)

	// templates
	templates := h.Templates

	homeTemp := template.New("home").Funcs(template.FuncMap{"LinkWrap": h.LinkWrap})
	templates.Set("home", template.Must(homeTemp.ParseFiles("static/home.html", "static/base.html", "static/links.html", "static/linkActions.html")))

	homeDataTemp := template.New("homedata").Funcs(template.FuncMap{"LinkWrap": h.LinkWrap})
	templates.Set("homedata", template.Must(homeDataTemp.ParseFiles("static/links.html", "static/linkActions.html")))
	templates.Set("edit", template.Must(template.ParseFiles("static/edit.html", "static/base.html")))

	viewTemp := template.New("view").Funcs(template.FuncMap{"LinkWrap": h.LinkWrap})
	templates.Set("view", template.Must(viewTemp.ParseFiles("static/view.html", "static/base.html", "static/linkActions.html")))
	templates.Set("login", template.Must(template.ParseFiles("static/login.html", "static/base.html")))
	templates.Set("faq", template.Must(template.ParseFiles("static/faq.html", "static/base.html")))

	userTemp := template.New("user").Funcs(template.FuncMap{"LinkWrap": h.LinkWrap})
	templates.Set("user", template.Must(userTemp.ParseFiles("static/user.html", "static/base.html", "static/links.html", "static/linkActions.html")))
	templates.Set("userSaved", template.Must(userTemp.ParseFiles("static/user.html", "static/base.html", "static/links.html", "static/linkActions.html")))

	// router
	router := http.NewServeMux()

	router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	router.HandleFunc("/", h.HomeHandler)
	router.HandleFunc("/data/", h.HomeDataHandler)
	router.HandleFunc("/api/data/", h.HomeDataApiHandler)
	router.HandleFunc("GET /login/", h.LoginHandler)
	router.HandleFunc("POST /login/", h.LoginPostHandler)
	router.HandleFunc("GET /u/{username}/", func(w http.ResponseWriter, r *http.Request) {
		h.UserHandler(w, r, h.GetUserLinks)
	})
	router.HandleFunc("POST /u/{username}/", func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.URL.Path)
		http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
	})
	router.HandleFunc("/api/u/{username}/", func(w http.ResponseWriter, r *http.Request) {
		h.UserDataApiHandler(w, r, h.GetUserLinks)
	})

	loggedInRouter := http.NewServeMux()
	loggedInRouter.HandleFunc("GET /r/{id}/submit/", h.EditHandler)
	loggedInRouter.HandleFunc("POST /r/{id}/submit/", h.SaveHandler)
	loggedInRouter.HandleFunc("GET /u/{username}/saved/", func(w http.ResponseWriter, r *http.Request) {
		h.UserHandler(w, r, h.GetUserSavedLinks)
	})
	loggedInRouter.HandleFunc("POST /u/{username}/saved/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
	})
	loggedInRouter.HandleFunc("/api/u/{username}/saved/", func(w http.ResponseWriter, r *http.Request) {
		h.UserDataApiHandler(w, r, h.GetUserSavedLinks)
	})
	loggedInRouter.HandleFunc("GET /u/{username}/settings/", func(w http.ResponseWriter, r *http.Request) {
		h.UserHandler(w, r, h.GetUserSettings)
	})
	loggedInRouter.HandleFunc("POST /u/{username}/settings/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
	})
	router.HandleFunc("/r/{id}/", h.ViewHandler)
	router.HandleFunc("POST /logout/", h.LogoutHandler)
	router.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/robots.txt")
	})
	router.HandleFunc("/image/", h.ImageHandler)
	router.HandleFunc("/faq/", h.FaqHandler)
	router.HandleFunc("POST /r/{id}/like/", h.LikeHandler)
	router.HandleFunc("/edit/{url}/", h.RedirectHandler)

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
	router.Handle("/api/u/{username}/saved/", strict(loggedInRouter))

	server := http.Server{
		Addr:    ":8090",
		Handler: stack(router),
	}

	server.ListenAndServe()
}
