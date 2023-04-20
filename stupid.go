package main

import (
	"context"
	"fmt"
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
	swApp := NewDefaultApp("swassistant", client)
	cdrApp := NewDefaultApp("cdr", client)
	webApp := NewDarkstormtech(client)
	stupid := stupid.NewStupidBackend(db.NewMongoTable(client.Database("stupid").Collection("keys")), func(app string) stupid.App {
		switch app {
		case "swassistant":
			return swApp
		case "cdr":
			return cdrApp
		case "darkstormtech":
			return webApp
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

type defaultApp struct {
	d *mongo.Database
}

func NewDefaultApp(name string, c *mongo.Client) *defaultApp {
	return &defaultApp{
		d: c.Database(name),
	}
}

func (d *defaultApp) Logs() db.Table {
	return db.NewMongoTable(d.d.Collection("logs"))
}

func (d *defaultApp) Crashes() db.CrashTable {
	return db.NewMongoTable(d.d.Collection("crashes"))
}

func (d *defaultApp) Extension(*stupid.Request) bool {
	return false
}

type darkstormtech struct {
	*defaultApp
}

func NewDarkstormtech(c *mongo.Client) *darkstormtech {
	return &darkstormtech{
		defaultApp: NewDefaultApp("darkstormtech", c),
	}
}

func (d *darkstormtech) Extension(r *stupid.Request) bool {
	if len(r.Path) != 2 {
		return false
	}
	if r.Path[0] != "page" {
		return false
	}
	res := d.d.Collection("pages").FindOne(context.TODO(), bson.M{"page": r.Path[1]})
	if res.Err() == mongo.ErrNoDocuments {
		r.Resp.WriteHeader(http.StatusNotFound)
		return true
	} else if res.Err() != nil {
		log.Println("Error while finding darkstorm.tech page:", res.Err())
		r.Resp.WriteHeader(http.StatusInternalServerError)
		return true
	}
	darkstormPage := struct {
		Content string
	}{}
	err := res.Decode(&darkstormPage)
	if err != nil {
		log.Println("Error while decoding darkstorm.tech page:", err)
		r.Resp.WriteHeader(http.StatusInternalServerError)
		return true
	}
	fmt.Println(darkstormPage)
	_, err = r.Resp.Write([]byte(darkstormPage.Content))
	if err != nil {
		log.Println("Error while sending darkstorm.tech page:", err)
		r.Resp.WriteHeader(http.StatusInternalServerError)
		return true
	}
	return true
}
