package blog

import (
	"net/http"

	"github.com/CalebQ42/darkstorm-server/internal/backend"
	"go.mongodb.org/mongo-driver/mongo"
)

type BlogApp struct {
	back    *backend.Backend
	blogCol *mongo.Collection
	authCol *mongo.Collection
}

func NewBlogApp(b *backend.Backend, db *mongo.Database, mux *http.ServeMux) *BlogApp {
	out := &BlogApp{
		back:    b,
		blogCol: db.Collection("blog"),
		authCol: db.Collection("author"),
	}
	// setup mux
	mux.HandleFunc("GET /blog", out.LatestBlogs)
	mux.HandleFunc("GET /blog/list", out.BlogList)
	mux.HandleFunc("GET /blog/{blogID}", out.Blog)

	mux.HandleFunc("POST /blog", out.CreateBlog)
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
