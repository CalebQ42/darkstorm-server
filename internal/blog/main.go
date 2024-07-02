package blog

import (
	"net/http"

	"github.com/CalebQ42/bbConvert"
	"github.com/CalebQ42/darkstorm-server/internal/backend"
	"go.mongodb.org/mongo-driver/mongo"
)

type BlogApp struct {
	back    *backend.Backend
	blogCol *mongo.Collection
	authCol *mongo.Collection
	conv    *bbConvert.HTMLConverter
}

func NewBlogApp(b *backend.Backend, db *mongo.Database, mux *http.ServeMux) *BlogApp {
	out := &BlogApp{
		back:    b,
		blogCol: db.Collection("blog"),
		authCol: db.Collection("author"),
		conv:    &bbConvert.HTMLConverter{},
	}
	out.conv.ImplementDefaults()
	// setup mux
	mux.HandleFunc("GET /blog", out.reqLatestBlogs)
	mux.HandleFunc("GET /blog/list", out.reqBlogList)
	mux.HandleFunc("GET /blog/{blogID}", out.reqBlog)
	mux.HandleFunc("POST /blog", out.createBlog)
	mux.HandleFunc("POST /blog/{blogID}", out.updateBlog)

	mux.HandleFunc("GET /author/{authorID}", out.reqAuthorInfo)
	mux.HandleFunc("POST /author", out.addAuthorInfo)
	mux.HandleFunc("POST /author/{authorID}", out.updateAuthorInfo)
	//TODO
	return out
}

func (b *BlogApp) AppID() string {
	return "blog"
}

func (b *BlogApp) CountTable() backend.CountTable {
	return nil
}

func (b *BlogApp) CrashTable() backend.CrashTable {
	return nil
}
