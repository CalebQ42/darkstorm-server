package main

import (
	"context"
	"crypto/tls"
	"flag"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/CalebQ42/stupid-backend"
	"github.com/CalebQ42/stupid-backend/pkg/db"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func webserver(mongoStr string) {
	path := flag.Arg(0)
	keyPath := flag.Arg(1)
	if path == "" {
		log.Println("No argument given for website file path. website signing off...")
		quitChan <- "web arg"
		return
	} else if keyPath == "" {
		log.Println("No argument given for key files. website signing off...")
		quitChan <- "web arg"
		return
	}
	if mongoStr != "" {
		client, err := mongo.NewClient(options.Client().ApplyURI(mongoStr))
		if err != nil {
			log.Println("Issues connecting to mongo:", err)
			quitChan <- "web arg"
			return
		}
		err = client.Connect(context.TODO())
		if err != nil {
			log.Println("Issues connecting to mongo:", err)
			quitChan <- "web arg"
			return
		}
		stupid := stupid.NewStupidBackend(db.NewMongoTable(client.Database("stupid").Collection("keys")))
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
		stupid.SetApps(map[string]db.App{
			"swassistant": {
				Logs:    db.NewMongoTable(client.Database("swassistant").Collection("log")),
				Crashes: db.NewMongoTable(client.Database("swassistant").Collection("crash")),
			},
			"cdr": {
				Logs:    db.NewMongoTable(client.Database("cdr").Collection("log")),
				Crashes: db.NewMongoTable(client.Database("cdr").Collection("crash")),
			},
		})
		http.Handle("api.darkstorm.tech/", stupid)
	}
	http.Handle("/", http.FileServer(http.Dir(path)))
	http.Handle("/SWAssistant/", swaHandler{})
	http.Handle("/CDR/", cdrHandler{})
	http.Handle("rpg.darkstorm.tech/", sup{})
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	err := http.ListenAndServeTLS(":443", keyPath+"/fullchain.pem", keyPath+"/key.pem", nil)
	log.Println("Error while serving website:", err)
	quitChan <- "web err"
}

type sup struct{}

func (s sup) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	url, err := url.Parse("https://localhost:30000")
	if err != nil {
		log.Println(err)
		http.FileServer(http.Dir(flag.Arg(0))).ServeHTTP(writer, req)
	}
	rvProx := httputil.NewSingleHostReverseProxy(url)
	rvProx.ServeHTTP(writer, req)
}

type swaHandler struct{}

func (swaHandler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	if _, err := os.Open(path.Join(flag.Arg(0) + req.URL.EscapedPath())); strings.Contains(req.URL.EscapedPath(), "#") || err == nil {
		http.FileServer(http.Dir(flag.Arg(0))).ServeHTTP(writer, req)
	} else {
		http.Redirect(writer, req, "https://darkstorm.tech/SWAssistant/#"+strings.TrimPrefix(req.URL.EscapedPath(), "/SWAssistant"), http.StatusFound)
		// log.Println("https://darkstorm.tech/SWAssistant/#" + strings.TrimPrefix(req.URL.EscapedPath(), "/SWAssistant"))
	}
}

type cdrHandler struct{}

func (cdrHandler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	if _, err := os.Open(path.Join(flag.Arg(0) + req.URL.EscapedPath())); strings.Contains(req.URL.EscapedPath(), "#") || err == nil {
		http.FileServer(http.Dir(flag.Arg(0))).ServeHTTP(writer, req)
	} else {
		http.Redirect(writer, req, "https://darkstorm.tech/CDR/#"+strings.TrimPrefix(req.URL.EscapedPath(), "/CDR"), http.StatusFound)
		// log.Println("https://darkstorm.tech/SWAssistant/#" + strings.TrimPrefix(req.URL.EscapedPath(), "/SWAssistant"))
	}
}
