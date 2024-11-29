package http

import (
	"log"
	c "main/common"
	"net/http"
)

func FaqHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value(c.UserCtx).(*c.User)
	if !ok {
		log.Println("user context missing")
	}

	if err := Templates.Get("faq").ExecuteTemplate(w, "base", user); err != nil {
		log.Printf("faqHandler err: %v", err)
	}
}
