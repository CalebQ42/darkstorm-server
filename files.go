package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

const fileElement = "<div class='files-link'><a href='https://darkstorm.tech%v'>%v</a><div style='float:right;'>%v</div></div>"

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
			slices.SortFunc(dirs, func(a, b os.DirEntry) int {
				return strings.Compare(a.Name(), b.Name())
			})
			if err != nil {
				pageContent = "<p>Server error!</p>"
				w.WriteHeader(http.StatusInternalServerError)
				log.Println("error serving files:", err)
			}
			for _, f := range dirs {
				if f.IsDir() {
					continue
				}
				inf, _ := f.Info()
				pageContent += fmt.Sprintf(fileElement, filepath.Join(partPath, f.Name()), f.Name(), inf.ModTime().Format(time.DateOnly))
			}
		} else {
			http.ServeFile(w, r, path)
			return
		}
	}
	sendContent(w, r, pageContent, "Files", "")
}
