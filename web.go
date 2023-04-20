package main

import (
	"crypto/tls"
	"flag"
	"fmt"
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

func (f *fileOrIndexHandler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	reqPath := strings.Split(strings.TrimPrefix(path.Clean(req.URL.Path), "/"), "/")
	if len(reqPath) == 0 {
		reqPath = []string{"index.html"}
	}
	fils, err := os.ReadDir(f.baseFolder)
	if err != nil {
		if os.IsNotExist(err) {
			writer.WriteHeader(http.StatusNotFound)
		} else {
			log.Println("Error while ReadDir:", err)
			writer.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	filename := path.Clean(f.baseFolder)
	var found bool
outer:
	for pathI := 0; pathI < len(reqPath); pathI++ {
		found = false
		for filI := range fils {
			if strings.EqualFold(strings.ToLower(fils[filI].Name()), reqPath[pathI]) {
				found = true
				filename = path.Join(filename, fils[filI].Name())
				if pathI == len(reqPath)-1 {
					if fils[filI].IsDir() {
						reqPath = append(reqPath, "index.html")
					}
					break
				} else if !fils[filI].IsDir() {
					break outer
				} else {
					fils, err = os.ReadDir(filename)
					if err != nil {
						log.Println("Error while ReadDir:", err)
						writer.WriteHeader(http.StatusInternalServerError)
						return
					}
				}
				break
			}
		}
		if !found {
			break
		}
	}
	if !found {
		for _, a := range f.appFolders {
			if strings.EqualFold(reqPath[0], a) {
				http.ServeFile(writer, req, path.Join(f.baseFolder, a, "index.html"))
				return
			}
		}
	}
	fil, err := os.Open(filename)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	st, _ := fil.Stat()
	if st.IsDir() {
		var subFil *os.File
		subFil, err = os.Open(path.Join(filename, "index.html"))
		if os.IsNotExist(err) {
			fmt.Println("file server for", filename)
			http.FileServer(http.Dir(filename)).ServeHTTP(writer, req)
			return
		}
		fil = subFil
	}
	http.ServeFile(writer, req, fil.Name())
}
