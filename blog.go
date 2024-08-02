package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/CalebQ42/darkstorm-server/internal/backend"
	"github.com/CalebQ42/darkstorm-server/internal/blog"
)

const (
	blogTitle  = "<h1 class='blog-title'>%v</h1>"
	blogAuthor = "<h4 class='blog-author'><i><b>By %v</b></i></h4>"
	blogCreate = "<h5 class='blog-time'><i>Written on: %v</i></h5"
	blogMain   = "<div class='blog'>%v</div>"

	authorInfo = `
<table><tr>
	<td><img src="%v" alt="%v" class='author-pic'></td>
	<td><h2 class="author-title">%v</h2>%v</td>
</tr></table>`
)

func latestBlogsHandle(w http.ResponseWriter, _ *http.Request) {
	latest, err := blogApp.LatestBlogs(0)
	if err != nil {
		if err == backend.ErrNotFound {
			w.WriteHeader(404)
			sendIndexWithContent(w, "Page not found", "", "")
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("error getting latest blogs:", err)
		sendIndexWithContent(w, "Error getting page", "", "")
		return
	}
	var out string
	for _, b := range latest {
		out += blogElement(b)
	}
	sendIndexWithContent(w, out, "", "")
}

func blogHandle(w http.ResponseWriter, blog string) {
	bl, err := blogApp.Blog(blog)
	if err != nil {
		if err == backend.ErrNotFound {
			w.WriteHeader(404)
			sendIndexWithContent(w, "Page not found", "", "")
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error getting blog %v: %v\n", blog, err)
		sendIndexWithContent(w, "Error getting page", "", "")
		return
	}
	sendIndexWithContent(w, blogElement(bl), bl.Title, bl.Favicon)
}

func blogElement(b *blog.Blog) (out string) {
	out = fmt.Sprintf(blogTitle, b.Title)
	out += fmt.Sprintf(blogAuthor, b.Author)
	cTime := time.Unix(b.CreateTime, 0).Format(time.DateOnly)
	if b.UpdateTime > b.CreateTime {
		out += fmt.Sprintf(blogCreate, cTime+"; Last updated on: "+time.Unix(b.UpdateTime, 0).Format(time.DateOnly))
	} else {
		out += fmt.Sprintf(blogCreate, cTime)
	}
	out += fmt.Sprintf(blogMain, b.Blog)
	auth, err := blogApp.GetAuthor(b)
	if err == nil {
		out += authorSection(auth)
	}
	return
}

func authorSection(a *blog.Author) string {
	return fmt.Sprintf(authorInfo, a.PicURL, a.Name+"'s profile picture", a.Name, a.About)
}
