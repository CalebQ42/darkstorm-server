package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

const (
	contentReplace = "<!--Content-->"
	faviconReplace = "<!--Favicon-->"
	titleReplace   = "<!--Title-->"
)

func sendIndexWithContent(w http.ResponseWriter, content string, title string, favicon string) {
	indexFile, err := os.Open(filepath.Join(*webRoot, "index.html"))
	if err != nil {
		log.Println("error when opening main index.html:", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	dat, err := io.ReadAll(indexFile)
	if err != nil {
		log.Println("error reading main index.html:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	dat = bytes.ReplaceAll(dat, []byte(contentReplace), []byte(content))
	if title == "" {
		title = "Darkstorm.tech"
	}
	dat = bytes.ReplaceAll(dat, []byte(titleReplace), []byte(title))
	if favicon == "" {
		favicon = "https://darkstorm.tech/favicon.png"
	}
	dat = bytes.ReplaceAll(dat, []byte(faviconReplace), []byte(favicon))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(dat)
}
