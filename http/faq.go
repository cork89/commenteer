package http

import (
	"log"
	c "main/common"
	"net/http"
)

type faqData struct {
	User *c.User
}

func FaqHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value(c.UserCtx).(*c.User)
	if !ok {
		log.Println("user context missing")
	}
	faqData := faqData{User: user}

	if err := Templates.Get("faq").ExecuteTemplate(w, "base", faqData); err != nil {
		log.Printf("faqHandler err: %v", err)
	}
}
