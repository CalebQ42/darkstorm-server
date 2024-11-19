package blog

import (
	"log"
	"net/http"
	"sync"
	"text/template"

	"github.com/CalebQ42/darkstorm-server/internal/backend"
	"go.mongodb.org/mongo-driver/mongo"
)

type WrapperFunc func(w http.ResponseWriter, r *http.Request, title, content string)

type Backend struct {
	blogCol *mongo.Collection
	authCol *mongo.Collection
	projCol *mongo.Collection

	tmpl    *template.Template
	wrapper WrapperFunc

	back *backend.Backend

	cacheMutex sync.RWMutex
	cache      map[string]string
}

func New(c *mongo.Client, back *backend.Backend, wrapper WrapperFunc) (*Backend, error) {
	var b = &Backend{
		blogCol: c.Database("blog").Collection("blog"),
		authCol: c.Database("blog").Collection("blog"),
		projCol: c.Database("blog").Collection("blog"),
		wrapper: wrapper,
		back:    back,
		cache:   make(map[string]string),
	}
	return b, b.parseTemplates()
}

func (b *Backend) RegisterToMux(mux *http.ServeMux) {
	mux.HandleFunc("GET /editor", b.editorReq)

	mux.HandleFunc("GET /editor/blog", b.blogPageReq)
	mux.HandleFunc("GET /editor/blog/edit", b.blogFormReq)
	mux.HandleFunc("POST /editor/blog/post", b.postBlogReq)

	// mux.HandleFunc("GET /editor/portfolio", b.portfolioPageReq)
	// mux.HandleFunc("GET /editor/portfolio/edit", b.portfolioFormReq)
	// mux.HandleFunc("POST /editor/portfolio/post", b.postPortfolioReq)

	// mux.HandleFunc("GET /editor/author", b.authorPageReq)
	// mux.HandleFunc("GET /editor/author/edit", b.authorFormReq)
	// mux.HandleFunc("POST /editor/author/post", b.postAuthorReq)
}

func (b *Backend) verifyEditorCookie(r *http.Request) *backend.User {
	authCookie, err := r.Cookie("blogAuthToken")
	if err != nil {
		if err != http.ErrNoCookie {
			log.Println("error getting auth cookie:", err)
		}
		return nil
	}
	usr, err := b.back.VerifyUser(r.Context(), authCookie.Value)
	if err != nil {
		if err != backend.ErrTokenUnauthorized {
			log.Println("error authorizing JWT token:", err)
		}
		return nil
	}
	return usr
}

func redirect(w http.ResponseWriter, r *http.Request, path string) {
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Location", `{"path": "`+path+`", "target":"#content"}`)
		return
	}
	http.Redirect(w, r, "https://darkstorm.tech"+path, http.StatusFound)
}
