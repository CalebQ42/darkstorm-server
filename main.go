package main

import (
	"context"
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

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
	webRoot     *string
)

func main() {
	addr := ":4223"
	mongoURL := flag.String("mongo", "", "Enables MongoDB usage for Darkstorm backend.")
	webRoot = flag.String("web-root", "", "Sets root directory of web server.")
	flag.Parse()
	if flag.NArg() != 1 {
		log.Fatal("You must specify key directory. ex: darkstorm-server /etc/web-keys")
	}
	if *mongoURL == "" || *webRoot == "" {
		log.Fatal("SPECIFY MONGO AND WEB-ROOT OR I WILL DIE, OH NO, THEY'RE COMING FOR ME.... **DEATH NOISES**")
	}
	go func() {
		http.ListenAndServe(":80", http.RedirectHandler("https://darkstorm.tech", http.StatusPermanentRedirect))
	}()
	mux := http.NewServeMux()
	setupMongo(*mongoURL)
	setupBackend(mux)
	setupWebsite(mux)
	serv := &http.Server{
		Addr:    addr,
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
	testApp := backend.NewSimpleApp("testing",
		db.NewMongoTable[backend.CountLog](mongoClient.Database("testing").Collection("count")),
		db.NewMongoCrashTable(
			mongoClient.Database("testing").Collection("crash"),
			mongoClient.Database("testing").Collection("archive"),
		))
	blogApp = blog.NewBlogApp(back, mongoClient.Database("blog"))
	//TODO: SWAssistant and CDR backends
	var err error
	back, err = backend.NewBackend(db.NewMongoTable[backend.ApiKey](
		mongoClient.Database("darkstorm").Collection("keys")),
		testApp,
		blogApp,
	)
	if err != nil {
		log.Fatal("error setting up backend:", err)
	}
	mux.Handle("api.darkstorm.tech/", back)
}

func setupWebsite(mux *http.ServeMux) {
	mux.HandleFunc("GET /files", filesRequest)
	mux.HandleFunc("GET /portfolio", portfolioRequest)
	mux.HandleFunc("/", mainHandle)
}

func mainHandle(w http.ResponseWriter, r *http.Request) {
	path := path.Clean(r.URL.Path)
	if path == "/" || path == "" {
		latestBlogsHandle(w, r)
		return
	}
	stat, err := os.Stat(filepath.Join(*webRoot, path))
	if err == nil && !stat.IsDir() {
		http.ServeFile(w, r, filepath.Join(*webRoot, path))
		return
	}
	spl := strings.Split(path, "/")
	stat, err = os.Stat(filepath.Join(*webRoot, spl[0]))
	if err == nil && stat.IsDir() {
		http.ServeFile(w, r, filepath.Join(*webRoot, spl[0], "index.html"))
		return
	}
	blogHandle(w, path)
}
