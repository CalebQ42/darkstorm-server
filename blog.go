package main

import (
	"net/http"

	"github.com/CalebQ42/darkstorm-server/internal/backend"
)

const (
	blogTitle  = "<h2 class='blog-title'>%v</h2>"
	blogAuthor = "<h4 class='blog-author'><i>%v</i></h4>"
	blogMain   = "<p class='blog'>%v</p>"
)

func latestBlogsHandle(w http.ResponseWriter, r *http.Request) {}

func blogHandle(w http.ResponseWriter, r *http.Request, blog string) {
	bl, err := blogApp.Blog(blog)
	if err != nil {
		if err == backend.ErrNotFound {
			w.WriteHeader(404)
			sendIndexWithContent(w, "Page not found")
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		sendIndexWithContent(w, "Error getting page")
		return
	}

}
