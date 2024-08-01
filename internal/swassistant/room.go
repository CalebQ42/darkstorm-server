package swassistant

import (
	"context"
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
	res, err := s.db.Collection("rooms").Find(context.TODO(), bson.M{"users": hdr.User.Username}, options.Find().SetProjection(bson.M{"_id": 1, "name": 1, "owner": 1}))
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
		err = res.All(context.TODO(), &out)
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
	if req.Method != http.MethodPost || req.Query["name"] == nil || len(req.Query["name"]) != 1 || req.Query["name"][0] == "" {
		req.Resp.WriteHeader(http.StatusBadRequest)
		return true
	} else if req.User == nil {
		req.Resp.WriteHeader(http.StatusUnauthorized)
		return true
	}
	//TODO: check room name for unsavory words
	newRoom := Room{
		ID:       uuid.NewString(),
		Name:     req.Query["name"][0],
		Owner:    req.User.Username,
		Users:    []string{},
		Profiles: []string{},
	}
	_, err := s.db.Collection("rooms").InsertOne(context.TODO(), newRoom)
	if err != nil {
		log.Println("SWAssistant: Error creating room:", err)
		req.Resp.WriteHeader(http.StatusInternalServerError)
		return true
	}
	out, err := json.Marshal(map[string]string{"id": newRoom.ID, "name": newRoom.Name})
	if err != nil {
		log.Println("SWAssistant: Error encoding new room:", err)
		req.Resp.WriteHeader(http.StatusInternalServerError)
		return true
	}
	req.Resp.WriteHeader(http.StatusCreated)
	_, err = req.Resp.Write(out)
	if err != nil {
		log.Println("SWAssistant: Error writing new room:", err)
		req.Resp.WriteHeader(http.StatusInternalServerError)
	}
	return true
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
	if req.Method != http.MethodGet {
		req.Resp.WriteHeader(http.StatusBadRequest)
		return true
	} else if req.User == nil {
		req.Resp.WriteHeader(http.StatusUnauthorized)
		return true
	}
	res := s.db.Collection("rooms").FindOne(context.TODO(), bson.M{"_id": req.Path[1]})
	if res.Err() == mongo.ErrNoDocuments {
		req.Resp.WriteHeader(http.StatusNotFound)
		return true
	} else if res.Err() != nil {
		log.Println("SWAssistant: Error getting room:", res.Err())
		req.Resp.WriteHeader(http.StatusInternalServerError)
		return true
	}
	r := Room{}
	err := res.Decode(&r)
	if err != nil {
		log.Println("SWAssistant: Error decoding room:", err)
		req.Resp.WriteHeader(http.StatusInternalServerError)
		return true
	}
	out, err := json.Marshal(r)
	if err != nil {
		log.Println("SWAssistant: Error encoding room:", err)
		req.Resp.WriteHeader(http.StatusInternalServerError)
		return true
	}
	_, err = req.Resp.Write(out)
	if err != nil {
		log.Println("SWAssistant: Error writing room:", err)
		req.Resp.WriteHeader(http.StatusInternalServerError)
	}
	return true
}
