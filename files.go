package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

const fileElement = "<p><a href='https://darkstorm.tech/%v/'>%v</a></p>"

func filesRequest(w http.ResponseWriter, r *http.Request) {
	partPath := filepath.Clean(r.URL.Path)
	path := filepath.Join(*webRoot, partPath)
	var pageContent string
	fil, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			pageContent = "<p>404 Not Found</p>"
			w.WriteHeader(http.StatusNotFound)
		} else {
			pageContent = "<p>Server error!</p>"
			w.WriteHeader(http.StatusInternalServerError)
			log.Println("error serving files:", err)
		}
	} else {
		stat, _ := fil.Stat()
		if stat.IsDir() {
			var dirs []os.DirEntry
			dirs, err = fil.ReadDir(-1)
			if err != nil {
				pageContent = "<p>Server error!</p>"
				w.WriteHeader(http.StatusInternalServerError)
				log.Println("error serving files:", err)
			}
			for _, f := range dirs {
				if f.IsDir() {
					continue
				}
				pageContent += fmt.Sprintf(fileElement, partPath, f.Name())
			}
		} else {
			http.ServeFile(w, r, path)
			return
		}
	}
	sendContent(w, r, pageContent, "Files", "")
}
