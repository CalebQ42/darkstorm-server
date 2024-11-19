package blog

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Blog struct {
	ID         string `json:"id" bson:"_id"`
	Author     string `json:"author" bson:"author"`
	Favicon    string `json:"favicon" bson:"favicon"`
	Title      string `json:"title" bson:"title"`
	RawBlog    string `json:"blog" bson:"blog"`
	HTMLBlog   string `json:"-" bson:"-"`
	StaticPage bool   `json:"staticPage" bson:"staticPage"`
	Draft      bool   `json:"draft" bson:"draft"`
	CreateTime int64  `json:"createTime" bson:"createTime"`
	UpdateTime int64  `json:"updateTime" bson:"updateTime"`
}

func (b *Backend) GetBlog(ctx context.Context, ID string, allowDraft bool) (Blog, error) {
	filter := bson.M{"_id": ID}
	if !allowDraft {
		filter["draft"] = false
	}
	res := b.blogCol.FindOne(ctx, filter)
	if res.Err() != nil {
		return Blog{}, res.Err()
	}
	var out Blog
	err := res.Decode(&out)
	return out, err
}

func (b *Backend) postBlogReq(w http.ResponseWriter, r *http.Request) {
	usr := b.verifyEditorCookie(r)
	if usr == nil {
		redirect(w, r, "/login")
		return
	}
	if usr.Perm["blog"] != "admin" && usr.Perm["blog"] != "writer" {
		w.Write([]byte("<p>Sorry, but you aren't authorized to do this action.</p>"))
		return
	}
	err := r.ParseForm()
	if err != nil {
		w.Write([]byte("<p>Error decoding form</p>"))
		return
	}
	newBlog := Blog{
		ID:         r.FormValue("id"),
		Title:      r.FormValue("title"),
		RawBlog:    r.FormValue("blog"),
		StaticPage: r.FormValue("staticPage") == "on",
		Draft:      r.FormValue("draft") == "on",
		UpdateTime: time.Now().Unix(),
	}
	if newBlog.Title == "" || newBlog.RawBlog == "" {
		w.Write([]byte("<p>Title and blog contents are required</p>"))
		return
	}
	if newBlog.ID == "" {
		newBlog.ID = strings.ToLower(strings.ReplaceAll(newBlog.Title, " ", "-"))
		newBlog.CreateTime = newBlog.UpdateTime
		newBlog.Author = usr.Username
		_, err = b.blogCol.InsertOne(r.Context(), newBlog)
		if mongo.IsDuplicateKeyError(err) {
			w.Write([]byte("<p>Title already exists</p>"))
			return
		} else if err != nil {
			log.Println("error inserting document")
			w.Write([]byte("<p>Server error inserting document</p>"))
			return
		}
		b.blogSuccessFullPageReplace(r.Context(), w, newBlog)
		return
	}
	res := b.blogCol.FindOne(r.Context(), bson.M{"_id": newBlog.ID})
	if res.Err() == mongo.ErrNoDocuments {
		w.Write([]byte("<p>Error finding old document</p>"))
		return
	} else if res.Err() != nil {
		log.Println("error getting old blog for update:", res.Err())
		w.Write([]byte("<p>Server error!</p>"))
		return
	}
	var oldBlog Blog
	err = res.Decode(&oldBlog)
	if err != nil {
		log.Println("error decoding old blog for update:", res.Err())
		w.Write([]byte("<p>Server error!</p>"))
		return
	}
	newBlog.CreateTime = oldBlog.CreateTime
	if oldBlog.Title != newBlog.Title {
		res = b.blogCol.FindOne(r.Context(), bson.M{"_id": strings.ToLower(strings.ReplaceAll(newBlog.Title, " ", "-"))})
		if res.Err() == nil {
			w.Write([]byte("<p>Title already exists</p>"))
			return
		} else if res.Err() != mongo.ErrNoDocuments {
			log.Println("error checking for title existance:", res.Err())
			w.Write([]byte("<p>Server error!</p>"))
			return
		}
		res = b.blogCol.FindOneAndDelete(r.Context(), bson.M{"_id": oldBlog.ID})
		if res.Err() != nil {
			log.Println("error deleting old blog:", res.Err())
			w.Write([]byte("<p>Server error!</p>"))
			return
		}
		newBlog.ID = strings.ToLower(strings.ReplaceAll(newBlog.Title, " ", "-"))
		_, err = b.blogCol.InsertOne(r.Context(), newBlog)
		if err != nil {
			log.Println("error inserting document")
			w.Write([]byte("<p>Server error inserting document</p>"))
			return
		}
		b.blogSuccessFullPageReplace(r.Context(), w, newBlog)
		return
	}
	_, err = b.blogCol.UpdateByID(r.Context(), newBlog.ID, bson.M{
		"updateTime": newBlog.UpdateTime,
		"blog":       newBlog.RawBlog,
		"staticPage": newBlog.StaticPage,
		"draft":      newBlog.Draft,
	})
	if err != nil {
		log.Println("error updating blog:", err)
		w.Write([]byte("<p>Server error inserting document</p>"))
	} else {
		w.Write([]byte("<p>Successfully updated</p>"))
	}
}

func (b *Backend) blogSuccessFullPageReplace(ctx context.Context, w http.ResponseWriter, blog Blog) {
	form, err := b.BlogEditForm(blog)
	if err != nil {
		log.Println("error with blogForm template:", err)
		w.Write([]byte("<p>Success, but error reloading page</p>"))
		return
	}
	out, err := b.BlogEditPage(ctx, blog.ID, form)
	if err != nil {
		log.Println("error with getting blog list:", err)
		w.Write([]byte("<p>Success, but error reloading page</p>"))
		return
	}
	w.Header().Set("Hx-Retarget", "#editPage")
	w.Write([]byte(out))
}

func (b *Backend) BlogEditForm(blog Blog) (string, error) {
	form := new(bytes.Buffer)
	err := b.tmpl.ExecuteTemplate(form, "blogForm", blogFormStruct{
		Blog:   blog,
		Result: "<p>Success!!</p>",
	})
	return form.String(), err
}

func (b *Backend) blogFormReq(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Hx-Request") != "true" {
		redirect(w, r, "/editor")
		return
	}
	blogID := r.URL.Query().Get("blog")
	var blog Blog
	if blogID != "" {
		var err error
		blog, err = b.GetBlog(r.Context(), blogID, true)
		if err != nil {
			log.Println("error getting blog:", err)
			w.Write([]byte("<p>Server error!</p>"))
			return
		}
	}
	form, err := b.BlogEditForm(blog)
	if err != nil {
		log.Println("error using blogForm template:", err)
		w.Write([]byte("<p>Server error!</p>"))
		return
	}
	w.Write([]byte(form))
}

func (b *Backend) BlogEditPage(ctx context.Context, selectedID, editor string) (string, error) {
	blogs, err := b.FullBlogList(ctx)
	if err != nil {
		return "", err
	}
	out := new(bytes.Buffer)
	err = b.tmpl.ExecuteTemplate(out, "blogPage", blogPageStruct{
		Selected: selectedID,
		Editor:   editor,
		Blogs:    blogs,
	})
	return out.String(), err
}

func (b *Backend) blogPageReq(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Hx-Request") != "true" {
		redirect(w, r, "/editor")
		return
	}
	form, err := b.BlogEditForm(Blog{})
	if err != nil {
		log.Println("error using blogForm template:", err)
		w.Write([]byte("<p>Server error!</p>"))
		return
	}
	page, err := b.BlogEditPage(r.Context(), "", form)
	if err != nil {
		log.Println("error using blogPage template:", err)
		w.Write([]byte("<p>Server error!</p>"))
		return
	}
	w.Write([]byte(page))
}
