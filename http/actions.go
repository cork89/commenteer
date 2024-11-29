package http

import (
	"fmt"
	"io"
	"log"
	c "main/common"
	d "main/dataaccess"
	s "main/snoo"
	"net/http"
)

func ImageHandler(w http.ResponseWriter, r *http.Request) {
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

func LikeHandler(w http.ResponseWriter, r *http.Request) {
	redditRequest, err := ExtractRedditRequest(r)
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
