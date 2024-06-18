package main

import (
	"context"
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"path/filepath"

	"github.com/CalebQ42/darkstorm-server/internal/backend"
	"github.com/CalebQ42/darkstorm-server/internal/backend/db"
	"github.com/CalebQ42/darkstorm-server/internal/blog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	mongoClient *mongo.Client
	back        *backend.Backend
	blogApp     *blog.BlogApp
)

func main() {
	mongoURL := flag.String("mongo", "", "Enables MongoDB usage for Darkstorm backend.")
	webRoot := flag.String("web-root", "", "Sets root directory of web server.")
	flag.Parse()
	if flag.NArg() != 1 {
		log.Fatal("You must specify key directory. ex: darkstorm-server /etc/web-keys")
	}
	if *mongoURL == "" || *webRoot == "" {
		log.Fatal("SPECIFY MONGO AND WEB-ROOT OR I WILL DIE (Death noises).")
	}
	go func() {
		http.ListenAndServe(":80", http.RedirectHandler("https://darkstorm.tech", http.StatusPermanentRedirect))
	}()
	mux := http.NewServeMux()
	setupMongo(*mongoURL)
	setupBackend(mux)
	setupWebsite(mux, *webRoot)
	serv := &http.Server{
		Addr:    ":443",
		Handler: mux,
	}
	err := serv.ListenAndServeTLS(filepath.Join(flag.Arg(0), "cert.pem"), filepath.Join(flag.Arg(0), "key.pem"))
	log.Println("webserver closed:", err)
}

func setupMongo(uri string) {
	mongoCert, err := tls.LoadX509KeyPair(filepath.Join(flag.Arg(0), "mongo.pem"), filepath.Join(flag.Arg(0)+"key.pem"))
	if err != nil {
		log.Fatal("error loading mongo keys:", err)
	}
	mongoClient, err = mongo.Connect(context.Background(), options.Client().ApplyURI(uri).SetTLSConfig(&tls.Config{
		Certificates: []tls.Certificate{mongoCert},
	}))
	if err != nil {
		log.Fatal("error connecting to mongo:", err)
	}
}

func setupBackend(mux *http.ServeMux) {
	blogApp = blog.NewBlogApp(back, mongoClient.Database("blog"), mux)
	//TODO: SWAssistant and CDR backends
	var err error
	back, err = backend.NewBackend(db.NewMongoTable[backend.ApiKey](mongoClient.Database("darkstorm").Collection("keys")), blogApp)
	if err != nil {
		log.Fatal("error setting up backend:", err)
	}
}

func setupWebsite(mux *http.ServeMux, root string) {}
