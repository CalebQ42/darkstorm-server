package blog

import (
	"context"
	"encoding/json"
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

type Blog struct {
	ID         string `json:"id" bson:"_id"`
	Author     string `json:"author" bson:"author"`
	Favicon    string `json:"favicon" bson:"favicon"`
	Title      string `json:"title" bson:"title"`
	Blog       string `json:"blog" bson:"blog"`
	StaticPage bool   `json:"staticPage" bson:"staticPage"`
	Draft      bool   `json:"draft" bson:"draft"`
	CreateTime int64  `json:"createTime" bson:"createTime"`
	UpdateTime int64  `json:"updateTime" bson:"updateTime"`
}

func (b *BlogApp) ConvertBlog(blog *Blog) {
	//TODO: parse BBCode/Markdown from blog
	if !blog.StaticPage {
		blog.Blog = b.conv.HTMLConvert(blog.Blog)
	}
}

func (b *BlogApp) GetAuthor(blog *Blog) (*Author, error) {
	res := b.authCol.FindOne(context.Background(), bson.M{"_id": blog.Author})
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

func (b *BlogApp) Blog(ID string) (*Blog, error) {
	res := b.blogCol.FindOne(context.Background(), bson.M{"_id": ID, "draft": false})
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

func (b *BlogApp) reqBlog(w http.ResponseWriter, r *http.Request) {
	blogID := r.PathValue("blogID")
	if blogID == "" {
		backend.ReturnError(w, http.StatusBadRequest, "badRequest", "Must provide a blogID")
		return
	}
	blog, err := b.Blog(blogID)
	if err != nil {
		if err == backend.ErrNotFound {
			backend.ReturnError(w, http.StatusNotFound, "notFound", "Not blog found with the given ID")
			return
		}
		log.Println("error getting blog:", err)
		backend.ReturnError(w, http.StatusInternalServerError, "internal", "Server error")
		return
	}
	json.NewEncoder(w).Encode(blog)
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
	_, err = b.blogCol.InsertOne(context.Background(), newBlog)
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
	res, err := b.blogCol.UpdateByID(context.Background(), r.PathValue("blogID"), reqUpd)
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

func (b *BlogApp) LatestBlogs(page int64) ([]*Blog, error) {
	res, err := b.blogCol.Find(context.Background(), bson.M{"staticPage": false, "draft": false}, options.Find().
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
	err = res.All(context.Background(), &out)
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
	blogs, err := b.LatestBlogs(int64(page))
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
	CreateTime int    `json:"createTime" bson:"createTime"`
}

func (b *BlogApp) BlogList(page int64) ([]BlogListResult, error) {
	res, err := b.blogCol.Find(context.Background(), bson.M{}, options.Find().
		SetProjection(bson.M{"_id": 1, "createTime": 1}).
		SetSort(bson.M{"createTime": 1}).
		SetLimit(50).
		SetSkip(page*50))
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, backend.ErrNotFound
		}
		return nil, err
	}
	var out []BlogListResult
	err = res.All(context.Background(), &out)
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
	blogList, err := b.BlogList(int64(page))
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
