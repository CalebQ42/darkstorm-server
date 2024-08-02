package swassistant

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/CalebQ42/darkstorm-server/internal/backend"
	"github.com/lithammer/shortuuid/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UploadedProf struct {
	Profile    map[string]any `json:"profile" bson:"profile"`
	ID         string         `json:"id" bson:"_id"`
	Type       string         `json:"type" bson:"type"`
	Expiration int64          `json:"expiration" bson:"expiration"`
}

func (s *SWBackend) UploadProfile(w http.ResponseWriter, r *http.Request) {
	hdr, err := s.back.VerifyHeader(w, r, "profile", false)
	if err != nil {
		return
	}
	if hdr.Key.AppID != "swassistant" {
		backend.ReturnError(w, http.StatusUnauthorized, "unauthorized", "Application not authorized")
		return
	}
	profType := r.URL.Query().Get("type")
	if profType == "" || (profType != "character" && profType != "vehicle" && profType != "minion") {
		backend.ReturnError(w, http.StatusBadRequest, "bad request", "Application sent a bad request")
		return
	}
	if r.Body == nil {
		backend.ReturnError(w, http.StatusBadRequest, "bad request", "Application sent a bad request")
		return
	}
	data, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil || len(data) == 0 {
		backend.ReturnError(w, http.StatusBadRequest, "bad request", "Application sent a bad request")
		return
	} else if len(data) > 5242880 { // 5MB
		backend.ReturnError(w, http.StatusRequestEntityTooLarge, "too large", "Profile is too large")
		return
	}
	prof := make(map[string]any)
	err = json.Unmarshal(data, &prof)
	if err != nil {
		backend.ReturnError(w, http.StatusBadRequest, "bad request", "Application sent a bad request")
		return
	}
	delete(prof, "uid")
	toUpload := UploadedProf{
		ID:         shortuuid.New(),
		Expiration: time.Now().Add(time.Hour * 12).Round(time.Hour).Unix(),
		Type:       profType,
		Profile:    prof,
	}
	_, err = s.db.Collection("profiles").InsertOne(context.Background(), toUpload)
	if err != nil {
		backend.ReturnError(w, http.StatusInternalServerError, "internal", "Server error")
		log.Println("error inserting profile:", err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{"id": toUpload.ID, "expiration": toUpload.Expiration})
}

func (s *SWBackend) GetProfile(w http.ResponseWriter, r *http.Request) {
	res := s.db.Collection("profiles").FindOne(context.Background(), bson.M{"_id": r.PathValue("profileID")})
	if res.Err() == mongo.ErrNoDocuments {
		backend.ReturnError(w, 404, "not found", "Profile not found")
		return
	} else if res.Err() != nil {
		backend.ReturnError(w, http.StatusInternalServerError, "internal", "Server error")
		log.Println("error getting profile:", res.Err())
		return
	}
	var prof UploadedProf
	err := res.Decode(&prof)
	if err != nil {
		backend.ReturnError(w, http.StatusInternalServerError, "internal", "Server error")
		log.Println("error decoding profile:", err)
		return
	}
	prof.Profile["type"] = prof.Type
	json.NewEncoder(w).Encode(prof.Profile)
}
