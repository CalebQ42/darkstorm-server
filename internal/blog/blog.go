package blog

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/CalebQ42/darkstorm-server/internal/backend"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	blogTitle  = "<h1 class='blog-title'><a hx-push-url='true' hx-get='/%v' hx-target='#content' href='/%v' style='text-decoration:none'>%v</a></h1>"
	blogAuthor = "<h4 class='blog-author'><i><b>By %v</b></i></h4>"
	blogCreate = "<h5 class='blog-time'><i>Written on: %v</i></h5>"
	blogMain   = "<div class='blog'>%v</div>"
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

func (b *Blog) HTMX(blogApp *BlogApp, ctx context.Context) string {
	if b.StaticPage {
		return b.RawBlog
	}
	out := fmt.Sprintf(blogTitle, b.ID, b.ID, b.Title)
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
	out += fmt.Sprintf(blogMain, b.HTMLBlog)
	if err == nil {
		out += "<h2 class='blog-author-info'>About the author:</h2>" + auth.HTML()
	}
	return out
}

func (b *BlogApp) ConvertBlog(blog *Blog) {
	//TODO: fix bbConvert
	// if !blog.StaticPage {
	// 	blog.HTMLBlog = b.conv.HTMLConvert(blog.RawBlog)
	// }
}

func (b *BlogApp) GetAuthor(ctx context.Context, blog *Blog) (*Author, error) {
	res := b.authCol.FindOne(ctx, bson.M{"_id": blog.Author})
	if res.Err() != nil {
		if res.Err() == mongo.ErrNoDocuments {
			return nil, backend.ErrNotFound
		}
		return nil, res.Err()
	}
	var author Author
	err := res.Decode(&author)
	return &author, err
}

func (b *BlogApp) Blog(ctx context.Context, ID string) (*Blog, error) {
	b.cacheMutex.RLock()
	blog, has := b.blogCache[ID]
	b.cacheMutex.RUnlock()
	if has {
		return &blog, nil
	}
	res := b.blogCol.FindOne(ctx, bson.M{"_id": ID, "draft": false})
	if res.Err() != nil {
		if res.Err() == mongo.ErrNoDocuments {
			return nil, backend.ErrNotFound
		}
		return nil, res.Err()
	}
	err := res.Decode(&blog)
	if err != nil {
		return nil, err
	}
	b.ConvertBlog(&blog)
	b.cacheMutex.Lock()
	b.blogCache[ID] = blog
	b.cacheMutex.Unlock()
	go b.CleanCache(ID)
	return &blog, nil
}

func (b *BlogApp) AnyBlog(ctx context.Context, ID string) (*Blog, error) {
	res := b.blogCol.FindOne(ctx, bson.M{"_id": ID})
	if res.Err() != nil {
		if res.Err() == mongo.ErrNoDocuments {
			return nil, backend.ErrNotFound
		}
		return nil, res.Err()
	}
	var blog Blog
	err := res.Decode(&blog)
	if err != nil {
		return nil, err
	}
	b.ConvertBlog(&blog)
	return &blog, nil
}

func (b *BlogApp) Contains(ctx context.Context, ID string) bool {
	res := b.blogCol.FindOne(ctx, bson.M{"_id": ID})
	return res.Err() == nil
}

func (b *BlogApp) CleanCache(ID string) {
	time.Sleep(5 * time.Minute)
	b.cacheMutex.Lock()
	delete(b.blogCache, ID)
	b.cacheMutex.Unlock()
}

func (b *BlogApp) reqBlog(w http.ResponseWriter, r *http.Request) {
	blogID := r.PathValue("blogID")
	if blogID == "" {
		backend.ReturnError(w, http.StatusBadRequest, "badRequest", "Must provide a blogID")
		return
	}
	blog, err := b.Blog(r.Context(), blogID)
	if err != nil {
		if err == backend.ErrNotFound {
			backend.ReturnError(w, http.StatusNotFound, "notFound", "Not blog found with the given ID")
			return
		}
		log.Println("error getting blog:", err)
		backend.ReturnError(w, http.StatusInternalServerError, "internal", "Server error")
		return
	}
	if r.Header.Get("Hx-Request") == "true" {
		w.Write([]byte(blog.HTMX(b, r.Context())))
	} else {
		json.NewEncoder(w).Encode(blog)
	}
}

func (b *BlogApp) createBlog(w http.ResponseWriter, r *http.Request) {
	hdr, err := b.back.VerifyHeader(w, r, "blogManagement", false)
	if hdr == nil {
		if err != nil {
			log.Println("request key parsing error:", err)
		}
		return
	} else if hdr.Key.AppID != "blog" {
		backend.ReturnError(w, http.StatusUnauthorized, "invalidKey", "Application is unauthorized")
		return
	}
	if hdr.User == nil || hdr.User.Perm["blog"] != "admin" {
		backend.ReturnError(w, http.StatusUnauthorized, "unauthorized", "Application is unauthorized")
		return
	}
	var newBlog Blog
	err = json.NewDecoder(r.Body).Decode(&newBlog)
	r.Body.Close()
	if err != nil {
		backend.ReturnError(w, http.StatusBadRequest, "badRequest", "Bad request")
		return
	}
	id, err := uuid.NewV7()
	if err != nil {
		backend.ReturnError(w, http.StatusInternalServerError, "internal", "Server Error")
		return
	}
	tim := time.Now().Unix()
	newBlog.ID = id.String()
	newBlog.CreateTime = tim
	newBlog.UpdateTime = tim
	newBlog.Author = hdr.User.Username
	_, err = b.blogCol.InsertOne(r.Context(), newBlog)
	if err != nil {
		log.Println("error when inserting new blog:", err)
		backend.ReturnError(w, http.StatusInternalServerError, "internal", "Server Error")
		return
	}
	w.WriteHeader(http.StatusCreated)

}

func (b *BlogApp) updateBlog(w http.ResponseWriter, r *http.Request) {
	hdr, err := b.back.VerifyHeader(w, r, "blogManagement", false)
	if hdr == nil {
		if err != nil {
			log.Println("request key parsing error:", err)
		}
		return
	} else if hdr.Key.AppID != "blog" {
		backend.ReturnError(w, http.StatusUnauthorized, "invalidKey", "Application is unauthorized")
		return
	}
	if hdr.User == nil || hdr.User.Perm["blog"] != "admin" {
		backend.ReturnError(w, http.StatusUnauthorized, "unauthorized", "Application is unauthorized")
		return
	}
	if r.PathValue("blogID") == "" {
		backend.ReturnError(w, http.StatusBadRequest, "badRequest", "Bad request")
		return
	}
	var reqUpdRaw map[string]string
	err = json.NewDecoder(r.Body).Decode(&reqUpdRaw)
	r.Body.Close()
	if err != nil {
		backend.ReturnError(w, http.StatusBadRequest, "badRequest", "Bad request")
		return
	}
	reqUpd := bson.M{}
	if fav, ok := reqUpdRaw["favicon"]; ok && fav != "" {
		reqUpd["favicon"] = fav
	}
	if titl, ok := reqUpdRaw["title"]; ok && titl != "" {
		reqUpd["title"] = titl
	}
	if blog, ok := reqUpdRaw["blog"]; ok && blog != "" {
		reqUpd["blog"] = blog
	}
	reqUpd["updateTime"] = time.Now().Unix()
	res, err := b.blogCol.UpdateByID(r.Context(), r.PathValue("blogID"), reqUpd)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			backend.ReturnError(w, http.StatusNotFound, "notFound", "Blog with ID "+r.PathValue("blogID")+" not found")
		} else {
			backend.ReturnError(w, http.StatusInternalServerError, "internal", "Server Error")
		}
		return
	}
	if res.MatchedCount == 0 {
		backend.ReturnError(w, http.StatusNotFound, "notFound", "Blog with ID "+r.PathValue("blogID")+" not found")
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (b *BlogApp) InsertBlog(ctx context.Context, blog Blog) error {
	_, err := b.blogCol.InsertOne(ctx, blog)
	return err
}

func (b *BlogApp) UpdateBlog(ctx context.Context, ID string, updates bson.M) error {
	_, err := b.blogCol.UpdateByID(ctx, ID, bson.M{"$set": updates})
	return err
}

func (b *BlogApp) LatestBlogs(ctx context.Context, page int64) ([]*Blog, error) {
	res, err := b.blogCol.Find(ctx, bson.M{"staticPage": false, "draft": false}, options.Find().
		SetSort(bson.M{"createTime": -1}).
		SetLimit(5).
		SetSkip(page*5))
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, backend.ErrNotFound
		}
		return nil, err
	}
	var out []*Blog
	err = res.All(ctx, &out)
	if err != nil {
		return nil, err
	}
	for i := range out {
		b.ConvertBlog(out[i])
	}
	return out, nil
}

func (b *BlogApp) reqLatestBlogs(w http.ResponseWriter, r *http.Request) {
	var page int
	var err error
	pagQuery := r.URL.Query().Get("page")
	if pagQuery != "" {
		page, err = strconv.Atoi(pagQuery)
		if err != nil {
			page = 0
		}
	}
	blogs, err := b.LatestBlogs(r.Context(), int64(page))
	if err != nil && err != backend.ErrNotFound {
		log.Println("error getting latest blogs:", err)
		backend.ReturnError(w, http.StatusInternalServerError, "internal", "internal error")
		return
	}
	var ret struct {
		Blogs []*Blog `json:"blogs"`
		Num   int     `json:"num"`
	}
	ret.Num = len(blogs)
	ret.Blogs = blogs
	json.NewEncoder(w).Encode(ret)
}

type BlogListResult struct {
	ID         string `json:"id" bson:"_id"`
	Title      string `json:"title" bson:"title"`
	CreateTime int    `json:"createTime" bson:"createTime"`
}

func (b BlogListResult) HTMX() string {
	return "<a class='blog-list-item' href='https://darkstorm.tech/" +
		b.ID +
		"' hx-push-url='true' hx-target='#content' hx-get='/" +
		b.ID +
		"'>" + b.Title + "</a>"
}

func (b *BlogApp) BlogList(ctx context.Context, page int64) ([]BlogListResult, error) {
	res, err := b.blogCol.Find(ctx, bson.M{"staticPage": false, "draft": false}, options.Find().
		SetProjection(bson.M{"_id": 1, "createTime": 1, "title": 1}).
		SetSort(bson.M{"createTime": -1}).
		SetLimit(50).
		SetSkip(page*50))
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, backend.ErrNotFound
		}
		return nil, err
	}
	var out []BlogListResult
	err = res.All(ctx, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (b *BlogApp) AllBlogsList(ctx context.Context) ([]BlogListResult, error) {
	res, err := b.blogCol.Find(ctx, bson.M{}, options.Find().
		SetProjection(bson.M{"_id": 1, "createTime": 1, "title": 1}).
		SetSort(bson.M{"createTime": -1}))
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, backend.ErrNotFound
		}
		return nil, err
	}
	var out []BlogListResult
	err = res.All(ctx, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (b *BlogApp) reqBlogList(w http.ResponseWriter, r *http.Request) {
	var page int
	var err error
	pagQuery := r.URL.Query().Get("page")
	if pagQuery != "" {
		page, err = strconv.Atoi(pagQuery)
		if err != nil {
			page = 0
		}
	}
	blogList, err := b.BlogList(r.Context(), int64(page))
	if err != nil && err != backend.ErrNotFound {
		log.Println("error getting blog list:", err)
		backend.ReturnError(w, http.StatusInternalServerError, "internal", "internal error")
		return
	}
	var ret struct {
		BlogList []BlogListResult `json:"blogList"`
		Num      int              `json:"num"`
	}
	ret.Num = len(blogList)
	ret.BlogList = blogList
	json.NewEncoder(w).Encode(ret)
}
