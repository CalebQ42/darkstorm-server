package blog

type Author struct {
	ID     string `json:"id" bson:"_id"`
	Name   string `json:"name" bson:"name"`
	About  string `json:"about" bson:"about"`
	PicURL string `json:"picurl" bson:"picurl"`
}
