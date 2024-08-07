package blog

import (
	"net/http"

	"github.com/CalebQ42/bbConvert"
	"github.com/CalebQ42/darkstorm-server/internal/backend"
	"go.mongodb.org/mongo-driver/mongo"
)

type BlogApp struct {
	back         *backend.Backend
	blogCol      *mongo.Collection
	authCol      *mongo.Collection
	portfolioCol *mongo.Collection
	conv         *bbConvert.HTMLConverter
}

func NewBlogApp(db *mongo.Database) *BlogApp {
	out := &BlogApp{
		blogCol:      db.Collection("blog"),
		authCol:      db.Collection("author"),
		portfolioCol: db.Collection("portfolio"),
		conv:         &bbConvert.HTMLConverter{},
	}
	out.conv.ImplementDefaults()
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

func (b *BlogApp) AddBackend(back *backend.Backend) {
	b.back = back
}

func (b *BlogApp) Extension(mux *http.ServeMux) {
	mux.HandleFunc("GET /blog", b.reqLatestBlogs)
	mux.HandleFunc("GET /blog/list", b.reqBlogList)
	mux.HandleFunc("GET /blog/{blogID}", b.reqBlog)
	mux.HandleFunc("POST /blog", b.createBlog)
	mux.HandleFunc("POST /blog/{blogID}", b.updateBlog)

	mux.HandleFunc("GET /blog/author/{authorID}", b.reqAuthorInfo)
	mux.HandleFunc("POST /blog/author", b.addAuthorInfo)
	mux.HandleFunc("POST /blog/author/{authorID}", b.updateAuthorInfo)

	mux.HandleFunc("GET /blog/portfolio", b.reqPortfolio)
}
