package blog

type BlogList struct {
	ID         string `json:"id" bson:"_id"`
	Title      string `json:"title" bson:"title"`
	Draft      bool   `json:"draft" bson:"draft"`
	CreateTime int64  `json:"createTime" bson:"createTime"`
}
