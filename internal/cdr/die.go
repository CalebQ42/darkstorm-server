package cdr

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/CalebQ42/darkstorm-server/internal/backend"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UploadedDie struct {
	Die        map[string]any `json:"die" bson:"die"`
	ID         string         `json:"id" bson:"_id"`
	Expiration int64          `json:"expiration" bson:"expiration"`
}

func (b CDRBackend) UploadDie(w http.ResponseWriter, r *http.Request) {
	hdr, err := b.back.VerifyHeader(w, r, "dice", false)
	if err != nil {
		return
	}
	if hdr.Key.AppID != "cdr" {
		backend.ReturnError(w, http.StatusUnauthorized, "unauthorized", "Application not authorized")
		return
	}
	if r.Body == nil {
		backend.ReturnError(w, http.StatusBadRequest, "bad request", "Application sent a bad request")
		return
	}
	bod, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		backend.ReturnError(w, http.StatusBadRequest, "bad request", "Application sent a bad request")
		return
	}
	if len(bod) > 1048576 { //1MB
		backend.ReturnError(w, http.StatusRequestEntityTooLarge, "too large", "Die is too large to upload")
		return
	}
	var toUpload = UploadedDie{
		Die:        make(map[string]any),
		ID:         uuid.New().String(),
		Expiration: time.Now().Add(12 * time.Hour).Round(time.Hour).Unix(),
	}
	err = json.Unmarshal(bod, &toUpload.Die)
	if err != nil {
		backend.ReturnError(w, http.StatusBadRequest, "bad request", "Application sent a bad request")
		return
	}
	if toUpload.Die["uuid"] != nil {
		delete(toUpload.Die, "uuid")
	}
	_, err = b.db.Collection("dice").InsertOne(context.Background(), toUpload)
	if err != nil {
		backend.ReturnError(w, http.StatusInternalServerError, "internal", "Server error")
		log.Println("error inserting die:", err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{"id": toUpload.ID, "expiration": toUpload.Expiration})
}

func (b CDRBackend) GetDie(w http.ResponseWriter, r *http.Request) {
	res := b.db.Collection("dice").FindOne(context.Background(), bson.M{"_id": r.PathValue("dieID")})
	if res.Err() == mongo.ErrNoDocuments {
		backend.ReturnError(w, 404, "not found", "Die with the given id is not found")
		return
	} else if res.Err() != nil {
		backend.ReturnError(w, http.StatusInternalServerError, "internal", "Server error")
		log.Println("error getting CDR die:", res.Err())
		return
	}
	var dieGet UploadedDie
	err := res.Decode(&dieGet)
	if err != nil {
		backend.ReturnError(w, http.StatusInternalServerError, "internal", "Server error")
		log.Println("error decoding die:", err)
		return
	}
	json.NewEncoder(w).Encode(dieGet.Die)
}
