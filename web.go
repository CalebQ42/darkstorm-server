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
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	http.Handle("/SWAssistant/", swaHandler{})
	http.Handle("/", http.FileServer(http.Dir(path)))
	http.Handle("rpg.darkstorm.tech/", sup{})
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
	}
	cert, err := tls.LoadX509KeyPair(keyPath+"/cert.pem", keyPath+"/key.pem")
	if err != nil {
		log.Println("Error while serving website:", err)
		quitChan <- "web err"
		return
	}
	tlsConf.Certificates = append(tlsConf.Certificates, cert)
	serve := http.Server{
		Addr:      ":443",
		TLSConfig: tlsConf,
	}
	err = serve.ListenAndServeTLS("", "")
	// err := http.ListenAndServeTLS(":443", keyPath+"/cert.pem", keyPath+"/key.pem", nil)
	log.Println("Error while serving website:", err)
	quitChan <- "web err"
}

type sup struct{}

func (s sup) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	url, err := url.Parse("https://localhost:30000")
	if err != nil {
		fmt.Println(err)
		http.FileServer(http.Dir(flag.Arg(0))).ServeHTTP(writer, req)
	}
	rvProx := httputil.NewSingleHostReverseProxy(url)
	rvProx.ServeHTTP(writer, req)
}

type swaHandler struct{}

func (s swaHandler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	if _, err := os.Open(path.Join(flag.Arg(0) + req.URL.EscapedPath())); strings.Contains(req.URL.EscapedPath(), "#") || err == nil {
		http.FileServer(http.Dir(flag.Arg(0))).ServeHTTP(writer, req)
	} else {
		http.Redirect(writer, req, "https://darkstorm.tech/SWAssistant/#"+strings.TrimPrefix(req.URL.EscapedPath(), "/SWAssistant"), http.StatusFound)
		// log.Println("https://darkstorm.tech/SWAssistant/#" + strings.TrimPrefix(req.URL.EscapedPath(), "/SWAssistant"))
	}
}
