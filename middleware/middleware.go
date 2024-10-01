package middleware

import (
	"log"
	"main/snoo"
	"net/http"
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

func IsLoggedIn(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userCookie, ok := snoo.GetUserCookie(r)

		if ok {
			if time.Now().UTC().Compare(userCookie.RefreshExpireDtTm) > -1 {
				log.Printf("%s needs to re-auth", userCookie.Username)
			}
			next.ServeHTTP(w, r)
		} else {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		}
		// slog.Info("IsLoggedIn", "method", r.Method, "path", r.URL.Path)
		// next.ServeHTTP(w, r)
	})
}
