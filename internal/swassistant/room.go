package swassistant

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/CalebQ42/darkstorm-server/internal/backend"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Room struct {
	ID       string   `json:"id" bson:"_id"`
	Name     string   `json:"name" bson:"name"`
	Owner    string   `json:"owner" bson:"owner"`
	Users    []string `json:"users" bson:"users"`
	Profiles []string `json:"profiles" bson:"profiles"`
}

func (s *SWBackend) ListRooms(w http.ResponseWriter, r *http.Request) {
	hdr, err := s.back.VerifyHeader(w, r, "rooms", false)
	if err != nil {
		return
	}
	if hdr.Key.AppID != "swassistant" || hdr.User == nil {
		backend.ReturnError(w, http.StatusUnauthorized, "unauthorized", "Application not authorized")
		return
	}
	res, err := s.db.Collection("rooms").Find(r.Context(), bson.M{"users": hdr.User.Username},
		options.Find().SetProjection(bson.M{"_id": 1, "name": 1, "owner": 1}))
	if err != nil && err != mongo.ErrNoDocuments {
		log.Println("error getting room list:", err)
		backend.ReturnError(w, http.StatusInternalServerError, "internal", "Server error")
		return
	}
	out := make([]struct {
		ID    string `json:"id" bson:"_id"`
		Name  string `json:"name" bson:"name"`
		Owner string `json:"owner" bson:"owner"`
	}, 0)
	if err == nil {
		err = res.All(r.Context(), &out)
		if err != nil {
			log.Println("error decoding room list:", err)
			backend.ReturnError(w, http.StatusInternalServerError, "internal", "Server error")
			return
		}
	}
	json.NewEncoder(w).Encode(out)
}

func (s *SWBackend) NewRoom(w http.ResponseWriter, r *http.Request) {
	hdr, err := s.back.VerifyHeader(w, r, "rooms", false)
	if err != nil {
		return
	}
	if hdr.Key.AppID != "swassistant" || hdr.User == nil {
		backend.ReturnError(w, http.StatusUnauthorized, "unauthorized", "Application not authorized")
		return
	}
	name := r.URL.Query().Get("name")
	if name == "" {
		backend.ReturnError(w, http.StatusBadRequest, "bad request", "Application sent bad request")
		return
	}
	//TODO: check room name for unsavory words
	newRoom := Room{
		ID:       uuid.NewString(),
		Name:     name,
		Owner:    hdr.User.Username,
		Users:    []string{},
		Profiles: []string{},
	}
	_, err = s.db.Collection("rooms").InsertOne(r.Context(), newRoom)
	if err != nil {
		log.Println("error creating room:", err)
		backend.ReturnError(w, http.StatusInternalServerError, "internal", "Server error")
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"id": newRoom.ID, "name": newRoom.Name})
}

func (s *SWBackend) GetRoom(w http.ResponseWriter, r *http.Request) {
	hdr, err := s.back.VerifyHeader(w, r, "rooms", false)
	if err != nil {
		return
	}
	if hdr.Key.AppID != "swassistant" || hdr.User == nil {
		backend.ReturnError(w, http.StatusUnauthorized, "unauthorized", "Application not authorized")
		return
	}
	roomID := r.PathValue("roomID")
	res := s.db.Collection("rooms").FindOne(r.Context(), bson.M{"_id": roomID})
	if res.Err() == mongo.ErrNoDocuments {
		backend.ReturnError(w, http.StatusNotFound, "not found", "Room not found")
		return
	} else if res.Err() != nil {
		log.Println("error getting room:", res.Err())
		backend.ReturnError(w, http.StatusInternalServerError, "internal", "Server error")
		return
	}
	var rm Room
	err = res.Decode(&rm)
	if err != nil {
		log.Println("error decoding room:", err)
		backend.ReturnError(w, http.StatusInternalServerError, "internal", "Server error")
		return
	}
	json.NewEncoder(w).Encode(rm)
}
