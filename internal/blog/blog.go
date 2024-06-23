package blog

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/CalebQ42/darkstorm-server/internal/backend"
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
	CreateTime int    `json:"createTime" bson:"createTime"`
	UpdateTime int    `json:"updateTime" bson:"updateTime"`
}

func (b *Blog) ConvertBlog() {
	//TODO: parse BBCode/Markdown from blog
	//b.Blog = bbCodeConvert(b.Blog)
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

func (b *BlogApp) GetBlog(ID string) (*Blog, error) {
	res := b.blogCol.FindOne(context.Background(), bson.M{"_id": ID})
	if res.Err() != nil {
		if res.Err() == mongo.ErrNoDocuments {
			return nil, backend.ErrNotFound
		}
		return nil, res.Err()
	}
	var blog Blog
	err := res.Decode(blog)
	if err != nil {
		return nil, err
	}
	blog.ConvertBlog()
	return &blog, nil
}

func (b *BlogApp) Blog(w http.ResponseWriter, r *http.Request) {
	blogID := r.PathValue("blogID")
	if blogID == "" {
		backend.ReturnError(w, http.StatusBadRequest, "badRequest", "Must provide a blogID")
		return
	}
	blog, err := b.GetBlog(blogID)
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

func (b *BlogApp) CreateBlog(w http.ResponseWriter, r *http.Request) {
	//TODO
}

func (b *BlogApp) UpdateBlog(w http.ResponseWriter, r *http.Request) {
	//TODO
}

func (b *BlogApp) GetLatestBlogs(page int64) ([]Blog, error) {
	res, err := b.blogCol.Find(context.Background(), bson.M{}, options.Find().
		SetSort(bson.M{"createTime": 1}).
		SetLimit(5).
		SetSkip(page*5))
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, backend.ErrNotFound
		}
		return nil, err
	}
	var out []Blog
	err = res.All(context.Background(), &out)
	if err != nil {
		return nil, err
	}
	for i := range out {
		out[i].ConvertBlog()
	}
	return out, nil
}

func (b *BlogApp) LatestBlogs(w http.ResponseWriter, r *http.Request) {
	var page int
	var err error
	pagQuery := r.URL.Query().Get("page")
	if pagQuery != "" {
		page, err = strconv.Atoi(pagQuery)
		if err != nil {
			page = 0
		}
	}
	blogs, err := b.GetLatestBlogs(int64(page))
	if err != nil && err != backend.ErrNotFound {
		log.Println("error getting latest blogs:", err)
		backend.ReturnError(w, http.StatusInternalServerError, "internal", "internal error")
		return
	}
	var ret struct {
		Blogs []Blog `json:"blogs"`
		Num   int    `json:"num"`
	}
	ret.Num = len(blogs)
	ret.Blogs = blogs
	json.NewEncoder(w).Encode(ret)
}

type BlogListResult struct {
	ID         string `json:"id" bson:"_id"`
	CreateTime int    `json:"createTime" bson:"createTime"`
}

func (b *BlogApp) GetBlogList(page int64) ([]BlogListResult, error) {
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

func (b *BlogApp) BlogList(w http.ResponseWriter, r *http.Request) {
	var page int
	var err error
	pagQuery := r.URL.Query().Get("page")
	if pagQuery != "" {
		page, err = strconv.Atoi(pagQuery)
		if err != nil {
			page = 0
		}
	}
	blogList, err := b.GetBlogList(int64(page))
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
