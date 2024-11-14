package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/CalebQ42/darkstorm-server/internal/backend"
	"github.com/CalebQ42/darkstorm-server/internal/blog"
)

// //go:embed embed
// var editorFS embed.FS

const (
	loginPage = `
<form id="loginForm" hx-post="/login" hx-target="#formResult">
	<label for="username">Username:</label>
	<input name="username" id="usernameInput" onkeydown="return event.key != 'Enter';" type="text"></input>
	<label for="password">Password:</label>
	<input name="password" type="password" id="passwordInput"></input>
	<div id="formResult"></div>
	<button class="formButton" type="submit">Login</button>
</form>`
	editorPage = `
<p>
	<label for="blog" style="margin-right:10px">Blog:</label>
	<select id="blogSelect"
			name='blog'
			hx-get='/editor/edit'
			hx-target='#editor'>
		<option value=""></option>
		<option value='new'>New Blog</option>
	{{ range $blog := . }}
		<option value='{{$blog.ID}}'>{{$blog.Title}}</option>
	{{end}}
	</select>
</p>
<div id="editor" hx-on::after-swap="blogEditorResize()"><p>Select a blog!</p></div>
`
	editorForm = `
<form id="editorForm" hx-post="/editor/post">
	<input name="id" type="hidden" value="{{.ID}}"></input>
	<p>
		<label for="static" style="margin-right:10px">Static Page:</label><input type="checkbox" name="static"{{if .StaticPage}} checked {{end}}/>
		<span class="vertical-seperator"></span>
		<label for="draft" style="margin-right:10px">Draft:</label><input type="checkbox" name="draft"{{if .Draft}} checked {{end}}/>
	</p>
	<label for="title">Title</label>
	<input id="titleInput" name="title" value="{{.Title}}" type="text"/>
	<textarea id="blogEditor" name="blog" oninput="blogEditorResize()">{{.RawBlog}}</textarea>
	<p style="margin-right:0px;">
		<button class="formButton" type="submit" style="margin-right:0px;">{{if eq .ID ""}}Create{{else}}Update{{end}}</button>
		<button class="formButton" type="submit" style="margin-right:0px;"
				hx-get="/editor/edit"
				hx-include="#blogSelect"
				hx-target="#editor"
				hx-confirm="Undo all your changes??">
			Cancel
		</button>
	<p>
</form>`
)

func loginPageRequest(w http.ResponseWriter, r *http.Request) {
	sendContent(w, r, loginPage, "", "")
}

func trueLoginRequest(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("HX-Request") != "true" {
		sendContent(w, r, "<p>Bad request</p>", "", "")
		return
	}
	err := r.ParseForm()
	if err != nil {
		sendContent(w, r, "<p>Bad request</p>", "", "")
		return
	}
	u, err := back.TryLogin(r.Context(), r.FormValue("username"), r.FormValue("password"))
	if err != nil {
		if err == backend.ErrLoginTimeout {
			sendContent(w, r, fmt.Sprint("<p>Timed out for", time.Unix(u.Timeout, 0).Sub(time.Now()), "</p>"), "", "")
		} else if err == backend.ErrLoginIncorrect {
			sendContent(w, r, "<p>Username or password invalid</p>", "", "")
		} else {
			log.Println("error trying to login:", err)
			sendContent(w, r, "<p>Server error</p>", "", "")
		}
		return
	}
	tok, err := back.GenerateJWT(u.ToReqUser())
	if err != nil {
		log.Println("error trying to generate JWT:", err)
		sendContent(w, r, "<p>Server error</p>", "", "")
		return
	}
	w.Header().Set("Set-Cookie", "blogAuthToken="+tok+"; Secure; Max-Age=43170") // Max-Age is 11.5 hours. JWTs are valid for 12 hours.
	sendContent(w, r, "<p hx-get='/editor' hx-push-url='true' hx-trigger='load' hx-target='#content'>Successful Login</p>", "", "")
}

func editorRequest(w http.ResponseWriter, r *http.Request) {
	if !verifyEditorCookie(r) {
		editorRedirect(w, r, "/login")
		return
	}
	tmpl, err := template.New("page").Parse(editorPage)
	if err != nil {
		log.Println("error parsing editor template:", err)
		sendContent(w, r, "ERROR", "", "")
		return
	}
	blogs, _ := blogApp.LatestBlogs(r.Context(), 0)
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, blogs)
	if err != nil {
		log.Println("error executing editor page template:", err)
		sendContent(w, r, "ERROR", "", "")
		return
	}
	sendContent(w, r, buf.String(), "", "")
}

func editorEdit(w http.ResponseWriter, r *http.Request) {
	if !verifyEditorCookie(r) {
		editorRedirect(w, r, "/login")
		return
	}
	tmpl, err := template.New("editor").Parse(editorForm)
	if err != nil {
		log.Println("error parsing editor template:", err)
		sendContent(w, r, "ERROR", "", "")
		return
	}
	var bl *blog.Blog
	blogID := r.URL.Query().Get("blog")
	if blogID == "" {
		sendContent(w, r, "<p>Select a blog!</p>", "", "")
		return
	}
	if blogID == "new" {
		bl = &blog.Blog{}
	} else {
		bl, err = blogApp.Blog(r.Context(), r.URL.Query().Get("blog"))
		if err != nil {
			log.Println("error getting blog for editor:", err)
			sendContent(w, r, "ERROR", "", "")
			return
		}
	}
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, bl)
	if err != nil {
		log.Println("error executing editor template:", err)
		sendContent(w, r, "ERROR", "", "")
		return
	}
	sendContent(w, r, buf.String(), "", "")
}

func verifyEditorCookie(r *http.Request) bool {
	authCookie, err := r.Cookie("blogAuthToken")
	if err != nil {
		if err != http.ErrNoCookie {
			log.Println("error getting auth cookie:", err)
		}
		return false
	}
	_, err = back.VerifyUser(r.Context(), authCookie.Value)
	if err != nil {
		if err != backend.ErrTokenUnauthorized {
			log.Println("error authorizing JWT token:", err)
		}
		return false
	}
	return true
}

func editorRedirect(w http.ResponseWriter, r *http.Request, path string) {
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Location", `{"path": "`+path+`", "target":"#content"}`)
		return
	}
	http.Redirect(w, r, "https://darkstorm.tech"+path, http.StatusFound)
}
