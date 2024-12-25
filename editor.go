package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/CalebQ42/darkstorm-server/internal/backend"
	"github.com/CalebQ42/darkstorm-server/internal/blog"
	"go.mongodb.org/mongo-driver/bson"
)

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
		<option value=""{{if eq .Selected ""}} selected{{end}}></option>
		<option value='new'>New Blog</option>
	{{ range $blog := .Blogs }}
		<option value='{{.ID}}'{{if eq $.Selected .ID}} selected{{end}}>{{.Title}}</option>
	{{end}}
	</select>
</p>
<div id="editor" hx-on::after-settle="blogEditorResize()">{{.Editor}}</div>`
	editorForm = `
<form id="editorForm" hx-post="/editor/post" hx-target="#formResult" hx-confirm="Save changes, overwritting previous values??">
	<input name="id" type="hidden" value="{{.Blog.ID}}"></input>
	<p>
		<label for="staticPage" style="margin-right:10px">Static Page:</label><input type="checkbox" name="staticPage"{{if .Blog.StaticPage}} checked {{end}}/>
		<span class="vertical-seperator"></span>
		<label for="draft" style="margin-right:10px">Draft:</label><input type="checkbox" name="draft"{{if or .Blog.Draft (not .Blog.ID)}} checked {{end}}/>
	</p>
	<label for="title">Title</label>
	<input id="titleInput" name="title" value="{{.Blog.Title}}" type="text" onkeydown="return event.key != 'Enter';"/>
	<textarea id="blogEditor" name="blog" oninput="blogEditorResize()">{{.Blog.RawBlog}}</textarea>
	<div id="formResult">{{.Result}}</div>
	<p style="margin-right:0px;">
		<button class="formButton" type="submit">{{if eq .Blog.ID ""}}Create{{else}}Update{{end}}</button>
		<button class="formButton"
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
		switch err {
		case backend.ErrLoginTimeout:
			sendContent(w, r, fmt.Sprint("<p>Timed out for", time.Until(time.Unix(u.Timeout, 0)), "</p>"), "", "")
		case backend.ErrLoginIncorrect:
			sendContent(w, r, "<p>Username or password invalid</p>", "", "")
		default:
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
	w.Header().Set("Set-Cookie", "blogAuthToken="+tok+"; Secure; Max-Age=43170; SameSite=Lax") // Max-Age is 11.5 hours. JWTs are valid for 12 hours.
	sendContent(w, r, "<p hx-get='/editor' hx-push-url='true' hx-trigger='load' hx-target='#content'>Successful Login</p>", "", "")
}

var (
	pageTmpl *template.Template
	formTmpl *template.Template
)

type pageTmplStruct struct {
	Selected string
	Blogs    []blog.BlogListResult
	Editor   string
}

type formTmplStruct struct {
	Blog   blog.Blog
	Result string
}

func setupEditorTemplates() error {
	var err error
	pageTmpl, err = template.New("page").Parse(editorPage)
	if err != nil {
		return err
	}
	formTmpl, err = template.New("form").Parse(editorForm)
	if err != nil {
		return err
	}
	return nil
}

func editorRequest(w http.ResponseWriter, r *http.Request) {
	if verifyEditorCookie(r) == nil {
		editorRedirect(w, r, "/login")
		return
	}
	blogs, err := blogApp.AllBlogsList(r.Context())
	if err != nil {
		log.Println("error getting all blogs:", err)
		sendContent(w, r, "ERROR", "", "")
		return
	}
	buf := new(bytes.Buffer)
	err = pageTmpl.Execute(buf, pageTmplStruct{Blogs: blogs})
	if err != nil {
		log.Println("error executing editor page template:", err)
		sendContent(w, r, "ERROR", "", "")
		return
	}
	sendContent(w, r, buf.String(), "", "")
}

func editorEdit(w http.ResponseWriter, r *http.Request) {
	if verifyEditorCookie(r) == nil {
		editorRedirect(w, r, "/login")
		return
	}
	var bl *blog.Blog
	blogID := r.URL.Query().Get("blog")
	if blogID == "" {
		sendContent(w, r, "<p>Select a blog!</p>", "", "")
		return
	}
	var err error
	if blogID == "new" {
		bl = &blog.Blog{}
	} else {
		bl, err = blogApp.AnyBlog(r.Context(), blogID)
		if err != nil {
			log.Println("error getting blog for editor:", err)
			log.Println(blogID)
			sendContent(w, r, "ERROR", "", "")
			return
		}
	}
	buf := new(bytes.Buffer)
	err = formTmpl.Execute(buf, formTmplStruct{Blog: *bl})
	if err != nil {
		log.Println("error executing editor template:", err)
		sendContent(w, r, "ERROR", "", "")
		return
	}
	sendContent(w, r, buf.String(), "", "")
}

func editorPost(w http.ResponseWriter, r *http.Request) {
	usr := verifyEditorCookie(r)
	if usr == nil {
		editorRedirect(w, r, "/login")
		return
	}
	if usr.Perm["blog"] != "admin" {
		sendContent(w, r, "<p>You are not allowed to perform this action. Sorry, not sorry.</p>", "", "")
		return
	}
	err := r.ParseForm()
	if err != nil {
		sendContent(w, r, "<p>Bad request</p>", "", "")
		return
	}
	newBlog := blog.Blog{
		ID:         r.FormValue("id"),
		Title:      r.FormValue("title"),
		RawBlog:    r.FormValue("blog"),
		Draft:      r.FormValue("draft") == "on",
		StaticPage: r.FormValue("staticPage") == "on",
	}
	if newBlog.Title == "" || newBlog.RawBlog == "" {
		sendContent(w, r, "<p>Title and Blog content required</p>", "", "")
		return
	}
	if newBlog.ID == "" {
		newBlog.ID = strings.TrimSpace(strings.ToLower(strings.ReplaceAll(newBlog.Title, " ", "-")))
		if blogApp.Contains(r.Context(), newBlog.ID) {
			sendContent(w, r, "<p>Title is not unique!</p>", "", "")
			return
		}
		now := time.Now()
		newBlog.CreateTime = now.Unix()
		newBlog.Author = usr.Username
		err = blogApp.InsertBlog(r.Context(), newBlog)
		if err != nil {
			log.Println("error creating new blog ID:", err)
			sendContent(w, r, "<p>Error inserting into DB</p>", "", "")
			return
		}
		var blogs []blog.BlogListResult
		blogs, err = blogApp.AllBlogsList(r.Context())
		if err != nil {
			log.Println("error getting all blogs list:", err)
			sendContent(w, r, "<p>Successfully save, but page reload failed</p>", "", "")
			return
		}
		w.Header().Set("HX-Retarget", "#content")
		newForm := new(bytes.Buffer)
		formTmpl.Execute(newForm, formTmplStruct{Blog: newBlog, Result: "<p>Successfully Created</p>"})
		pageTmpl.Execute(w, pageTmplStruct{Selected: newBlog.ID, Blogs: blogs, Editor: newForm.String()})
		return
	}
	err = blogApp.UpdateBlog(r.Context(), newBlog.ID,
		bson.M{
			"updateTime": time.Now().Unix(),
			"title":      newBlog.Title,
			"blog":       newBlog.RawBlog,
			"draft":      newBlog.Draft,
			"staticPage": newBlog.StaticPage})
	if err != nil {
		log.Println("error updating blog:", err)
		sendContent(w, r, "<p>Server error updating blog</p>", "", "")
		return
	}
	old, err := blogApp.AnyBlog(r.Context(), newBlog.ID)
	if err != nil {
		log.Println("error getting old blog to be updated:", err)
		sendContent(w, r, "<p>Updated!</p>", "", "")
		return
	}
	if old.Title == newBlog.Title {
		sendContent(w, r, "<p>Updated!</p>", "", "")
		return
	}
	var blogs []blog.BlogListResult
	blogs, err = blogApp.AllBlogsList(r.Context())
	if err != nil {
		log.Println("error getting all blogs list:", err)
		sendContent(w, r, "<p>Updated!</p>", "", "")
		return
	}
	w.Header().Set("HX-Retarget", "#content")
	newForm := new(bytes.Buffer)
	formTmpl.Execute(newForm, formTmplStruct{Blog: newBlog, Result: "<p>Successfully Created</p>"})
	pageTmpl.Execute(w, pageTmplStruct{Selected: newBlog.ID, Blogs: blogs, Editor: newForm.String()})
}

func verifyEditorCookie(r *http.Request) *backend.User {
	authCookie, err := r.Cookie("blogAuthToken")
	if err != nil {
		if err != http.ErrNoCookie {
			log.Println("error getting auth cookie:", err)
		}
		return nil
	}
	usr, err := back.VerifyUser(r.Context(), authCookie.Value)
	if err != nil {
		if err != backend.ErrTokenUnauthorized {
			log.Println("error authorizing JWT token:", err)
		}
		return nil
	}
	return usr
}

func editorRedirect(w http.ResponseWriter, r *http.Request, path string) {
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Location", `{"path": "`+path+`", "target":"#content"}`)
		return
	}
	http.Redirect(w, r, "https://darkstorm.tech"+path, http.StatusFound)
}
