package blog

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

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
	var newAuth Author
	err = json.NewDecoder(r.Body).Decode(&newAuth)
	r.Body.Close()
	if err != nil {
		backend.ReturnError(w, http.StatusBadRequest, "badRequest", "Invalid request")
		return
	}
	for i := 1; ; i++ {
		newID := strings.ReplaceAll(newAuth.Name, " ", "-")
		if i != 1 {
			newID += strconv.Itoa(i)
		}
		collisionCheck := b.authCol.FindOne(context.Background(), bson.M{"name": newAuth.Name})
		if collisionCheck.Err() == mongo.ErrNoDocuments {
			newAuth.ID = newID
			break
		} else if collisionCheck.Err() != nil {
			log.Println("error checking for new author ID collisions:", err)
			backend.ReturnError(w, http.StatusInternalServerError, "internal", "Server Error")
			return
		}
	}
	_, err = b.authCol.InsertOne(context.Background(), newAuth)
	if err != nil {
		log.Println("error inserting new author:", err)
		backend.ReturnError(w, http.StatusInternalServerError, "internal", "Server Error")
		return
	}
	w.WriteHeader(http.StatusCreated)
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
	var rawUpd map[string]string
	err = json.NewDecoder(r.Body).Decode(&rawUpd)
	r.Body.Close()
	if err != nil {
		backend.ReturnError(w, http.StatusBadRequest, "badRequest", "Invalid request")
		return
	}
	actlUpd := make(map[string]string)
	if rawUpd["name"] != "" {
		actlUpd["name"] = rawUpd["name"]
	}
	if rawUpd["about"] != "" {
		actlUpd["about"] = rawUpd["about"]
	}
	if rawUpd["picurl"] != "" {
		actlUpd["picurl"] = rawUpd["picurl"]
	}
	res, err := b.authCol.UpdateByID(context.Background(), r.PathValue("authorID"), actlUpd)
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
