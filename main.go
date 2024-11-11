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
	"path/filepath"
	"strings"

	"github.com/CalebQ42/darkstorm-server/internal/backend"
	"github.com/CalebQ42/darkstorm-server/internal/backend/db"
	"github.com/CalebQ42/darkstorm-server/internal/blog"
	"github.com/CalebQ42/darkstorm-server/internal/cdr"
	"github.com/CalebQ42/darkstorm-server/internal/swassistant"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	mongoClient *mongo.Client
	back        *backend.Backend
	blogApp     *blog.BlogApp
	webRoot     *string
	testing     *bool
)

func main() {
	mongoURL := flag.String("mongo", "", "Enables MongoDB usage for Darkstorm backend.")
	webRoot = flag.String("web-root", "", "Sets root directory of web server.")
	addr := flag.String("addr", ":443", "Set listen address. Defaults to \":443\"")
	testing = flag.Bool("testing", false, "Start in testing mode. If you don't know what this is, don't use it.")
	flag.Parse()
	if *testing {
		*addr = ":4242"
	}
	if !*testing && flag.NArg() != 1 {
		log.Fatal("You must specify key directory. ex: darkstorm-server /etc/web-keys")
	}
	if *mongoURL == "" || *webRoot == "" {
		log.Fatal("SPECIFY MONGO AND WEB-ROOT OR I WILL DIE, OH NO, THEY'RE COMING FOR ME.... **DEATH NOISES**")
	}
	if !*testing {
		go func() {
			log.Println("error redirecting http traffice:",
				http.ListenAndServe(":80", http.RedirectHandler("https://darkstorm.tech", http.StatusPermanentRedirect)))
		}()
	}
	mux := http.NewServeMux()
	setupMongo(*mongoURL)
	setupBackend(mux)
	setupWebsite(mux)
	serv := &http.Server{
		Addr:    *addr,
		Handler: mux,
	}
	var err error
	if *testing {
		err = serv.ListenAndServe()
	} else {
		err = serv.ListenAndServeTLS(filepath.Join(flag.Arg(0), "fullchain.pem"), filepath.Join(flag.Arg(0), "key.pem"))
	}
	log.Println("webserver closed:", err)
}

func setupMongo(uri string) {
	if !*testing {
		mongoCert, err := tls.LoadX509KeyPair(filepath.Join(flag.Arg(0), "mongo.pem"), filepath.Join(flag.Arg(0), "key.pem"))
		if err != nil {
			log.Fatal("error loading mongo keys:", err)
		}
		mongoClient, err = mongo.Connect(context.Background(), options.Client().ApplyURI(uri).SetTLSConfig(&tls.Config{
			Certificates: []tls.Certificate{mongoCert},
		}))
		if err != nil {
			log.Fatal("error connecting to mongo:", err)
		}
	} else {
		var err error
		mongoClient, err = mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
		if err != nil {
			log.Fatal("error connecting to mongo:", err)
		}
	}
}

func setupBackend(mux *http.ServeMux) {
	blogApp = blog.NewBlogApp(mongoClient.Database("blog"))
	var err error
	back, err = backend.NewBackend(db.NewMongoTable[backend.ApiKey](
		mongoClient.Database("darkstorm").Collection("keys")),
		blogApp,
		swassistant.NewSWBackend(mongoClient.Database("swassistant")),
		cdr.NewBackend(mongoClient.Database("cdr")),
	)
	if !*testing {
		back.AddCorsAddress("https://darkstorm.tech")
		var pubFil, privFil *os.File
		defer pubFil.Close()
		defer privFil.Close()
		var pub, priv []byte
		pubFil, err = os.Open(filepath.Join(flag.Arg(0), "darkstorm-pub.key"))
		if err != nil {
			log.Println("error openning darkstorm user public key:", err)
			goto here
		}
		pub, err = io.ReadAll(pubFil)
		if err != nil {
			log.Println("error reading darkstorm user public key:", err)
			goto here
		}
		privFil, err = os.Open(filepath.Join(flag.Arg(0), "darkstorm-priv.key"))
		if err != nil {
			log.Println("error openning darkstorm user private key:", err)
			goto here
		}
		priv, err = io.ReadAll(privFil)
		if err != nil {
			log.Println("error reading darkstorm user private key:", err)
			goto here
		}
		back.AddUserAuth(db.NewMongoTable[backend.User](mongoClient.Database("darkstorm").Collection("users")), priv, pub)
	} else {
		back.AddCorsAddress("*")
	}
here:
	if err != nil {
		log.Fatal("error setting up backend:", err)
	}
	if !*testing {
		mux.Handle("api.darkstorm.tech/", back)
	} else {
		go func() {
			http.ListenAndServe(":2323", back)
		}()
	}
}

func setupWebsite(mux *http.ServeMux) {
	if !*testing {
		url, _ := url.Parse("https://localhost:30000")
		mux.Handle("rpg.darkstorm.tech/", httputil.NewSingleHostReverseProxy(url))
	}
	edit := NewBlogEditor(blogApp, back)
	mux.HandleFunc("GET /files/{w...}", filesRequest)
	mux.HandleFunc("GET /portfolio", portfolioRequest)
	mux.HandleFunc("GET /list", blogListHandle)
	mux.HandleFunc("GET /login", edit.LoginPage)
	mux.HandleFunc("GET /editor", edit.Editor)
	mux.HandleFunc("POST /login", edit.TrueLogin)
	mux.HandleFunc("/", mainHandle)
}

func mainHandle(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
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
	ind := filepath.Join(*webRoot, spl[0], "index.html")
	_, err = os.Stat(ind)
	if err == nil {
		http.ServeFile(w, r, ind)
		return
	}
	blogHandle(w, r, path)
}
