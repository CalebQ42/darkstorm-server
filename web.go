package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

const replacementComment = "<!--Content-->"

func sendIndexWithContent(w http.ResponseWriter, content string) {
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
	dat = bytes.ReplaceAll(dat, []byte(replacementComment), []byte(content))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(dat)
}
