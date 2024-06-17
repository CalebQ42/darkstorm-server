package blog

import (
	"context"
	"log"
	"net/http"

	"github.com/CalebQ42/darkstorm-server/internal/backend"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Author struct {
	ID     string `json:"id" bson:"_id"`
	About  string `json:"about" bson:"about"`
	PicURL string `json:"picurl" bson:"picurl"`
}

func (b *BlogApp) AboutCaleb() (*Author, error) {
	res := b.authCol.FindOne(context.Background(), bson.M{"_id": "caleb_gardner"})
	if res.Err() != nil {
		log.Println("error getting about me:", res.Err())
		if res.Err() == mongo.ErrNoDocuments {
			return nil, backend.ErrNotFound
		}
		return nil, res.Err()
	}
	var aboutMe Author
	err := res.Decode(&aboutMe)
	if err != nil {
		log.Println("error decoding about me:", res)
		return nil, err
	}
	return &aboutMe, nil
}

func (b *BlogApp) GetAuthorInfo(w http.ResponseWriter, r *http.Request) {

}

func (b *BlogApp) SetAuthorInfo(w http.ResponseWriter, r *http.Request) {
	hdr, err := b.back.VerifyHeader(w, r, "managment", true)
	if hdr == nil {
		if err != nil {
			log.Println("error verifying apiKey:", err)
		}
		return
	}
	if hdr.Key.AppID != "blog" {
		backend.ReturnError(w, http.StatusUnauthorized, "invalidKey", "Application not authorized")
		return
	}

}
