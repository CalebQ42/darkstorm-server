package main

import (
	"embed"
	"io"
	"log"
	"net/http"

	"github.com/CalebQ42/darkstorm-server/internal/backend"
	"github.com/CalebQ42/darkstorm-server/internal/blog"
)

//go:embed embed
var editorFS embed.FS

type Editor struct {
	blogApp *blog.BlogApp
	back    *backend.Backend
}

func NewBlogEditor(blogApp *blog.BlogApp, back *backend.Backend) Editor {
	return Editor{blogApp: blogApp, back: back}
}

func (e Editor) LoginPage(w http.ResponseWriter, r *http.Request) {
	page, err := editorFS.Open("embed/login.html")
	defer page.Close()
	if err != nil {
		log.Println("error getting login.html:", err)
		sendContent(w, r, "error getting page", "", "")
		return
	}
	dat, err := io.ReadAll(page)
	if err != nil {
		log.Println("error reading login.html:", err)
		sendContent(w, r, "error getting page", "", "")
		return
	}
	sendContent(w, r, string(dat), "", "")
}

func (e Editor) Editor(w http.ResponseWriter, r *http.Request) {
	hdr, err := back.ParseHeader(r)
	if err == backend.ErrApiKeyUnauthorized || err == backend.ErrTokenUnauthorized || hdr == nil || hdr.User == nil {
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Location", `{"path":"/login", "target":"#content"}`)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		http.Redirect(w, r, "https://darkstorm.tech/login", http.StatusSeeOther)
		return
	}
	page, err := editorFS.Open("embed/editor.html")
	defer page.Close()
	if err != nil {
		log.Println("error getting editor.html:", err)
		sendContent(w, r, "error getting page", "", "")
		return
	}
	dat, err := io.ReadAll(page)
	if err != nil {
		log.Println("error reading editor.html:", err)
		sendContent(w, r, "error getting page", "", "")
		return
	}
	sendContent(w, r, string(dat), "", "")
}
