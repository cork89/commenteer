package middleware

import (
	"context"
	"log"
	"main/common"
	"main/snoo"
	"net/http"
	"strings"
	"time"
)

type Middleware func(http.Handler) http.Handler

func CreateStack(mw ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(mw) - 1; i >= 0; i-- {
			next = mw[i](next)
		}
		return next
	}
}

type scWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *scWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.statusCode = statusCode
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		opWriter := &scWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(opWriter, r)
		log.Println(opWriter.statusCode, r.Method, r.URL.Path, time.Since(start))
	})
}

func IsLoggedInStrict(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := snoo.GetUserCookie(r)

		if ok {
			ctx := context.WithValue(r.Context(), common.UserCtx, user)
			r = r.Clone(ctx)
			next.ServeHTTP(w, r)
		} else {
			http.Redirect(w, r, "/login/", http.StatusTemporaryRedirect)
		}
	})
}

func CheckRemainingUploads(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user, ok := ctx.Value(common.UserCtx).(*common.User)

		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		shouldRefresh := time.Now().UTC().After(user.UploadRefreshDtTm)

		if shouldRefresh {
			refreshed := snoo.RefreshUserUploadCount(user)
			if !refreshed {
				log.Printf("user failed to refresh upload count, user id: %d", user.UserId)
			}
			next.ServeHTTP(w, r)
		} else if user.RemainingUploads > 0 {
			next.ServeHTTP(w, r)
		} else {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		}
	})
}

func IsLoggedIn(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.RequestURI, "/static/") {
			next.ServeHTTP(w, r)
			return
		}

		ctx := r.Context()
		user, ok := ctx.Value(common.UserCtx).(*common.User)

		if !ok {
			user, ok = snoo.GetUserCookie(r)
		}

		if ok {
			if time.Now().UTC().Compare(user.RefreshExpireDtTm) > -1 {
				log.Printf("%s needs to re-auth", user.Username)
				user, _ = snoo.RefreshRedditAccessToken(user)
			}
			ctx := context.WithValue(r.Context(), common.UserCtx, user)
			r = r.Clone(ctx)
			next.ServeHTTP(w, r)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

func CacheControl(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=604800")
		next.ServeHTTP(w, r)
	})
}

func StaticMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/static/") {
			// Call the next handler to process other routes
			next.ServeHTTP(w, r)
			return
		}

		http.ServeFile(w, r, r.URL.Path[9:])
		next.ServeHTTP(w, r)
	})
}
