package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/CalebQ42/darkstorm-server/internal/backend"
	"github.com/CalebQ42/darkstorm-server/internal/blog"
)

const (
	blogTitle  = "<h1 class='blog-title'><a href='https://darkstorm.tech/%v' onclick=\"return setToPath('%v')\" style='text-decoration:none'>%v</a></h1>"
	blogAuthor = "<h4 class='blog-author'><i><b>By %v</b></i></h4>"
	blogCreate = "<h5 class='blog-time'><i>Written on: %v</i></h5>"
	blogMain   = "<div class='blog'>%v</div>"

	authorInfo = `
<h2 class='blog-author-info'>About the author:</h2>
<table><tr>
	<td><img src="%v" alt="%v" class='author-pic'></td>
	<td><h3 class="author-title">%v</h3>%v</td>
</tr></table>`
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
		out += blogElement(r.Context(), b)
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
	sendContent(w, r, blogElement(r.Context(), bl), bl.Title, bl.Favicon)
}

func blogElement(ctx context.Context, b *blog.Blog) (out string) {
	if b.StaticPage {
		return b.Blog
	}
	out = fmt.Sprintf(blogTitle, b.ID, b.ID, b.Title)
	auth, err := blogApp.GetAuthor(ctx, b)
	if err == nil {
		out += fmt.Sprintf(blogAuthor, auth.Name)
	} else {
		out += fmt.Sprintf(blogAuthor, "unknown")
	}
	cTime := time.Unix(b.CreateTime, 0).Format(time.DateOnly)
	if b.UpdateTime > b.CreateTime {
		out += fmt.Sprintf(blogCreate, cTime+"; Last updated on: "+time.Unix(b.UpdateTime, 0).Format(time.DateOnly))
	} else {
		out += fmt.Sprintf(blogCreate, cTime)
	}
	out += fmt.Sprintf(blogMain, b.Blog)
	if err == nil {
		out += authorSection(auth)
	}
	return
}

func authorSection(a *blog.Author) string {
	return fmt.Sprintf(authorInfo, a.PicURL, a.Name+"'s profile picture", a.Name, a.About)
}
