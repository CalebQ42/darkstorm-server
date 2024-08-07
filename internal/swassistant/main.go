package swassistant

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/CalebQ42/darkstorm-server/internal/backend"
	"github.com/CalebQ42/darkstorm-server/internal/backend/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type SWBackend struct {
	back *backend.Backend
	db   *mongo.Database
}

func NewSWBackend(db *mongo.Database) *SWBackend {
	go func() {
		for range time.Tick(time.Hour) {
			log.Println("SWAssistant: Deleting expired profiles")
			res, err := db.Collection("profiles").DeleteMany(context.Background(), bson.M{"expiration": bson.M{"$lt": time.Now().Unix()}})
			if err == mongo.ErrNoDocuments {
				continue
			}
			log.Println("SWAssistant: Deleted", res.DeletedCount, "profiles")
		}
	}()
	return &SWBackend{
		db: db,
	}
}

func (s *SWBackend) AppID() string {
	return "swassistant"
}

func (s *SWBackend) CountTable() backend.CountTable {
	return db.NewMongoTable[backend.CountLog](s.db.Collection("logs"))
}

func (s *SWBackend) CrashTable() backend.CrashTable {
	return db.NewMongoCrashTable(s.db.Collection("crashes"), s.db.Collection("crashArchive"))
}

func (s *SWBackend) AddBackend(b *backend.Backend) {
	s.back = b
}

func (s *SWBackend) AddCrash(cr backend.IndividualCrash) bool {
	res := s.db.Collection("versions").FindOne(context.Background(), bson.M{"version": cr.Version})
	return res.Err() != mongo.ErrNoDocuments
}

func (s *SWBackend) Extension(mux *http.ServeMux) {
	mux.HandleFunc("GET /swa/room", s.ListRooms)
	mux.HandleFunc("POST /swa/room", s.NewRoom)
	mux.HandleFunc("GET /swa/room/{roomID}", s.GetRoom)

	mux.HandleFunc("POST /swa/profile", s.UploadProfile)
	mux.HandleFunc("GET /swa/profile/{profileID}", s.GetProfile)

	//Legacy (TODO: remove this after a month or two after the applciation gets updated)
	mux.HandleFunc("GET /room/list", s.ListRooms)
	mux.HandleFunc("POST /room/new", s.NewRoom)
	mux.HandleFunc("GET /room/{roomID}", s.GetRoom)

	mux.HandleFunc("POST /profile/upload", s.UploadProfile)
	mux.HandleFunc("GET /profile/{profileID}", s.GetProfile)
}
