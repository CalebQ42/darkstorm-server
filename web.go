package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
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
	go func() {
		app := fiber.New()
		app.Static("/", path)
		app.Static("/files", filepath.Join(path, "files"), fiber.Static{
			Browse: true,
		})
		app.Static("/SWAssistant", filepath.Join(path, "SWAssistant"))
		err := app.ListenTLS(":443", filepath.Join(keyPath, "cert.pem"), filepath.Join(keyPath, "key.pem"))
		log.Println("Error while serving website:", err)
		quitChan <- "web err"
	}()
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
