package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/CalebQ42/stupid-backend"
	"github.com/CalebQ42/stupid-backend/pkg/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setupStupid(keyPath, mongoStr string) error {
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoStr))
	if err != nil {
		log.Println("Issues connecting to mongo:", err)
		return err
	}
	err = client.Connect(context.TODO())
	if err != nil {
		log.Println("Issues connecting to mongo:", err)
		return err
	}
	swApp := &stupid.App{
		Logs:    db.NewMongoTable(client.Database("swassistant").Collection("log")),
		Crashes: db.NewMongoTable(client.Database("swassistant").Collection("crash")),
	}
	cdrApp := &stupid.App{
		Logs:    db.NewMongoTable(client.Database("cdr").Collection("log")),
		Crashes: db.NewMongoTable(client.Database("cdr").Collection("crash")),
	}
	websiteApp := &stupid.App{
		Logs:    db.NewMongoTable(client.Database("darkstormtech").Collection("log")),
		Crashes: db.NewMongoTable(client.Database("darkstormtech").Collection("crash")),
		Extension: func(r *stupid.Request) bool {
			return websiteRequest(*client.Database("darkstormtech"), r)
		},
	}
	stupid := stupid.NewStupidBackend(db.NewMongoTable(client.Database("stupid").Collection("keys")), func(app string) *stupid.App {
		switch app {
		case "swassistant":
			return swApp
		case "cdr":
			return cdrApp
		case "darkstormtech":
			return websiteApp
		}
		return nil
	})
	users := true
	var pub, priv []byte
	stupidPubFil, err := os.Open(keyPath + "/stupid-pub.key")
	if err != nil {
		log.Println("Disabling API users:", err)
		users = false
	} else {
		pub, err = io.ReadAll(stupidPubFil)
		if err != nil {
			log.Println("Disabling API users:", err)
			users = false
		}
	}
	stupidPrivFil, err := os.Open(keyPath + "/stupid-pub.key")
	if err != nil {
		log.Println("Disabling API users:", err)
		users = false
	} else {
		priv, err = io.ReadAll(stupidPrivFil)
		if err != nil {
			log.Println("Disabling API users:", err)
			users = false
		}
	}
	if users {
		stupid.EnableUserAuth(db.NewMongoTable(client.Database("stupid").Collection("keys")), pub, priv)
	}
	http.Handle("api.darkstorm.tech/", stupid)
	return nil
}

type page struct {
	Page     string
	Contents string
}

func websiteRequest(d mongo.Database, r *stupid.Request) bool {
	if len(r.Path) > 0 {
		r.Resp.WriteHeader(http.StatusBadRequest)
		return true
	}
	if pag, ok := r.Query["page"]; ok {
		res := d.Collection("pages").FindOne(context.TODO(), bson.M{"page": pag})
		if res.Err() == mongo.ErrNoDocuments {
			r.Resp.WriteHeader(http.StatusNotFound)
			return true
		}
		var p page
		err := res.Decode(&p)
		if err != nil {
			r.Resp.WriteHeader(http.StatusInternalServerError)
			log.Print("page decode:", err)
			return true
		}
		_, err = r.Resp.Write([]byte(p.Contents))
		if err != nil {
			r.Resp.WriteHeader(http.StatusInternalServerError)
			log.Print("content send:", err)
			return true
		}
		return true
	}
	r.Resp.WriteHeader(http.StatusBadRequest)
	return true
}
