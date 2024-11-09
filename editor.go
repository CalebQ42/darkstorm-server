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

func (e Editor) Editor(w http.ResponseWriter, r *http.Request) {}
