package cdr

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

type CDRBackend struct {
	back *backend.Backend
	db   *mongo.Database
}

func NewBackend(db *mongo.Database) *CDRBackend {
	go func() {
		for range time.Tick(time.Hour) {
			log.Println("CDR: Deleting expired dice")
			res, err := db.Collection("profiles").DeleteMany(context.Background(), bson.M{"expiration": bson.M{"$lt": time.Now().Unix()}})
			if err == mongo.ErrNoDocuments {
				continue
			}
			log.Println("CDR: Deleted", res.DeletedCount, "dice")
		}
	}()
	return &CDRBackend{
		db: db,
	}
}

func (b CDRBackend) AppID() string {
	return "cdr"
}

func (b CDRBackend) CountTable() backend.CountTable {
	return db.NewMongoTable[backend.CountLog](b.db.Collection("logs"))
}

func (b CDRBackend) CrashTable() backend.CrashTable {
	return db.NewMongoCrashTable(b.db.Collection("crashes"), b.db.Collection("crashArchive"))
}

func (b *CDRBackend) AddBackend(back *backend.Backend) {
	b.back = back
}

func (s CDRBackend) AddCrash(cr backend.IndividualCrash) bool {
	res := s.db.Collection("versions").FindOne(context.Background(), bson.M{"version": cr.Version})
	return res.Err() != mongo.ErrNoDocuments
}

func (b CDRBackend) Extension(mux *http.ServeMux) {
	mux.HandleFunc("POST /cdr/die", b.UploadDie)
	mux.HandleFunc("GET /cdr/die/{dieID}", b.GetDie)

	//Legacy (TODO: remove this after a month or two after the applciation gets updated)
	mux.HandleFunc("POST /upload", b.UploadDie)
	mux.HandleFunc("GET /die/{dieID}", b.GetDie)
}
