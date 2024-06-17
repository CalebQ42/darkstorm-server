package main

import (
	"context"
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"path/filepath"

	"github.com/CalebQ42/darkstorm-server/internal/blog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	mongoClient *mongo.Client
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
	mux := http.NewServeMux()
	mongoClient = setupMongo(*mongoURL)
	setupBackend(mux)
	setupWebsite(mux, *webRoot)
	http.ListenAndServeTLS(":443", filepath.Join(flag.Arg(0), "cert.pem"), filepath.Join(flag.Arg(0), "key.pem"), mux)
}

func setupMongo(uri string) *mongo.Client {
	mongoCert, err := tls.LoadX509KeyPair(filepath.Join(flag.Arg(0), "mongo.pem"), filepath.Join(flag.Arg(0)+"key.pem"))
	if err != nil {
		log.Fatal("error loading mongo keys:", err)
	}
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri).SetTLSConfig(&tls.Config{
		Certificates: []tls.Certificate{mongoCert},
	}))
	if err != nil {
		log.Fatal("error connecting to mongo:", err)
	}
	return client
}

func setupWebsite(mux *http.ServeMux, root string) {}

func setupBackend(mux *http.ServeMux) {
}
