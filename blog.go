package main

import (
	"log"
	"net/http"

	"github.com/CalebQ42/darkstorm-server/internal/backend"
)

func latestBlogsHandle(w http.ResponseWriter, r *http.Request) {
	latest, err := blogApp.LatestBlogs(r.Context(), 0)
	if err != nil {
		if err == backend.ErrNotFound {
			w.WriteHeader(404)
			sendContent(w, r, "Page not found", "", "")
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("error getting latest blogs:", err)
		sendContent(w, r, "Error getting page", "", "")
		return
	}
	var out string
	for _, b := range latest {
		out += b.HTMX(blogApp, r.Context())
	}
	sendContent(w, r, out, "", "")
}

func blogHandle(w http.ResponseWriter, r *http.Request, blog string) {
	bl, err := blogApp.Blog(r.Context(), blog)
	if err != nil {
		if err == backend.ErrNotFound {
			w.WriteHeader(404)
			sendContent(w, r, "Page not found", "", "")
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error getting blog %v: %v\n", blog, err)
		sendContent(w, r, "Error getting page", "", "")
		return
	}
	sendContent(w, r, bl.HTMX(blogApp, r.Context()), bl.Title, bl.Favicon)
}
