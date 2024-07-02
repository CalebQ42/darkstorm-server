package blog

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/CalebQ42/darkstorm-server/internal/backend"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Author struct {
	ID     string `json:"id" bson:"_id"`
	Name   string `json:"name" bson:"name"`
	About  string `json:"about" bson:"about"`
	PicURL string `json:"picurl" bson:"picurl"`
}

func (b *BlogApp) AboutMe() (*Author, error) {
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

func (b *BlogApp) reqAuthorInfo(w http.ResponseWriter, r *http.Request) {
	res := b.authCol.FindOne(context.Background(), r.PathValue("authorID"))
	if res.Err() == mongo.ErrNoDocuments {
		backend.ReturnError(w, http.StatusNotFound, "notFound", "Author with ID "+r.PathValue("authorID")+" not found")
		return
	} else if res.Err() != nil {
		log.Println("error getting author info:", res.Err())
		backend.ReturnError(w, http.StatusInternalServerError, "internal", "Server Error")
		return
	}
	var auth Author
	err := res.Decode(&auth)
	if err != nil {
		log.Println("error decoding author info:", err)
		backend.ReturnError(w, http.StatusInternalServerError, "internal", "Server Error")
		return
	}
	json.NewEncoder(w).Encode(auth)
}

func (b *BlogApp) addAuthorInfo(w http.ResponseWriter, r *http.Request) {
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
	//TODO
}

func (b *BlogApp) updateAuthorInfo(w http.ResponseWriter, r *http.Request) {
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
	//TODO
}
