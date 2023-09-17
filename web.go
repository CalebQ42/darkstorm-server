package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"strings"
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
	var err error
	if mongoStr != "" {
		err = setupStupid(keyPath, mongoStr)
		if err != nil {
			quitChan <- "web err"
			return
		}
	}
	url, err := url.Parse("https://localhost:30000")
	if err != nil {
		log.Println("Can't parse foundry url:", err)
		quitChan <- "web err"
		return
	}
	// http.Handle("/", http.FileServer(http.Dir(path)))
	mainHandle := &fileOrIndexHandler{
		baseFolder: path,
		appFolders: []string{
			"SWAssistant",
			"CDR",
		},
	}
	http.Handle("/", mainHandle)
	// http.Handle("/SWAssistant/", swaHandler{})
	// http.Handle("/CDR/", cdrHandler{})
	http.Handle("rpg.darkstorm.tech/", httputil.NewSingleHostReverseProxy(url))
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	err = http.ListenAndServeTLS(":443", keyPath+"/fullchain.pem", keyPath+"/key.pem", nil)
	log.Println("Error while serving website:", err)
	quitChan <- "web err"
}

type fileOrIndexHandler struct {
	baseFolder string
	appFolders []string
}

func (f *fileOrIndexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	reqPath := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	if reqPath == "" || reqPath == "index.html" {
		http.ServeFile(w, r, path.Join(f.baseFolder, "index.html"))
		return
	}
	reqPath = path.Join(f.baseFolder, reqPath)
	if fil, err := os.Open(reqPath); err == nil {
		inf, _ := fil.Stat()
		if !inf.IsDir() {
			http.ServeFile(w, r, reqPath)
			return
		} else if _, err = os.Open(path.Join(reqPath, "index.html")); err == nil {
			http.ServeFile(w, r, path.Join(reqPath, "index.html"))
			return
		}
	}
	for _, a := range f.appFolders {
		if strings.HasPrefix(reqPath, path.Join(f.baseFolder, a)) {
			http.ServeFile(w, r, path.Join(f.baseFolder, a, "index.html"))
			return
		}
	}
	http.ServeFile(w, r, path.Join(f.baseFolder, "index.html"))
}
