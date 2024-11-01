package main

import (
	"log"
	"net/http"
)

func portfolioRequest(w http.ResponseWriter, r *http.Request) {
	selectedTech := r.URL.Query().Get("tech")
	proj, err := blogApp.Projects(r.Context(), selectedTech)
	if err != nil {
		log.Println("error getting portfolio projects:", err)
		w.WriteHeader(http.StatusInternalServerError)
		sendContent(w, r, "Error getting portfolio", "", "")
		return
	}
	sendContent(w, r, proj.FullHTMX(r.Context(), blogApp, selectedTech), "Portfolio", "")
}
