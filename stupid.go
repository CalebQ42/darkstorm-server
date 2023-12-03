package main

import (
	"context"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/CalebQ42/cdr-backend"
	"github.com/CalebQ42/darkstorm-server/internal/darkstormtech"
	"github.com/CalebQ42/stupid-backend/v2"
	"github.com/CalebQ42/stupid-backend/v2/db"
	swassistantbackend "github.com/CalebQ42/swassistant-backend"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setupStupid(keyPath, mongoStr string) error {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoStr))
	if err != nil {
		log.Println("Issues connecting to mongo:", err)
		return err
	}
	stupid := stupid.NewStupidBackend(db.NewMongoTable(client.Database("stupid").Collection("keys")), map[string]any{
		"swassistant":   swassistantbackend.NewSWBackend(client),
		"cdr":           cdr.NewBackend(client),
		"darkstormtech": darkstormtech.NewDarkstormTech(client, filepath.Join(flag.Arg(0), "files")),
	}, "https://darkstorm.tech")
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
