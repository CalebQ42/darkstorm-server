package blog

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
