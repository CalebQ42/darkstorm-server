package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
)

func webserver() {
	flag.Parse()
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
	http.Handle("/SWAssistant/", swaHandler{})
	http.Handle("/", http.FileServer(http.Dir(path)))
	err := http.ListenAndServeTLS(":443", keyPath+"/cert.pem", keyPath+"/key.pem", nil)
	log.Println("Error while serving website:", err)
	quitChan <- "web err"
}

type swaHandler struct {
}

func (s swaHandler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	if _, err := os.Open(path.Join(flag.Arg(0) + req.URL.EscapedPath())); strings.Contains(req.URL.EscapedPath(), "#") || err == nil {
		http.FileServer(http.Dir(flag.Arg(0))).ServeHTTP(writer, req)
	} else {
		http.Redirect(writer, req, "https://darkstorm.tech/SWAssistant/#"+strings.TrimPrefix(req.URL.EscapedPath(), "/SWAssistant"), http.StatusFound)
		// log.Println("https://darkstorm.tech/SWAssistant/#" + strings.TrimPrefix(req.URL.EscapedPath(), "/SWAssistant"))
	}
}
