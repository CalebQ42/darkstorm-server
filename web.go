package main

import (
	"flag"
	"log"
	"net/http"
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
	http.Handle("/", http.FileServer(http.Dir(path)))
	err := http.ListenAndServeTLS(":443", keyPath+"/cert.pem", keyPath+"/key.pem", nil)
	log.Println("Error while serving website:", err)
	quitChan <- "web err"
}
