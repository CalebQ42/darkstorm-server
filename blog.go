package main

import (
	"log"
	"net/http"
	"strconv"

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
	bl, err := blogApp.Blog(r.Context(), blog, true)
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

func blogListHandle(w http.ResponseWriter, r *http.Request) {
	pag, _ := strconv.Atoi(r.URL.Query().Get("page"))
	list, err := blogApp.BlogList(r.Context(), int64(pag))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		sendContent(w, r, "Error getting page", "", "")
		return
	}
	out := ""
	for i := range list {
		out += "<p>" + list[i].HTMX() + "</p>"
	}
	if pag > 0 || len(list) == 50 {
		out += "<div id='blog-list-page-selector'>"
		if pag > 0 {
			pagNum := strconv.Itoa(pag - 1)
			out += "<a href='https://darkstorm.tech/list?page=" + pagNum + "' hx-get='/list?page='" + pagNum + "' hx-push-url='true' hx-target='#content'>&lt;Previous</a>"
		}
		if len(list) == 50 {
			pagNum := strconv.Itoa(pag + 1)
			out += "<a href='https://darkstorm.tech/list?page=" + pagNum + "' hx-get='/list?page='" + pagNum + "' hx-push-url='true' hx-target='#content'>Next&gt;</a>"
		}
		out += "</div>"
	}
	sendContent(w, r, out, "", "")
}
