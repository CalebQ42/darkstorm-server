package main

import (
	"fmt"
	"log"
	"net/http"
)

const (
	portfolioSelector       = "<p>Language Filter: <select id='langSelect' name='langSelect'>%v</select></p>"
	portfolioSelectorOption = "<option value='%v'%v>%v</option>"
	portfolioTitle          = "<h2 class='portfolio-title'>%v</h2>"
	portfolioLink           = "<p class='portfolio-link'><a href='%v'>%v</a>"
	portfolioLanguage       = "<p class='portfolio-language'><b>%v</b>: %v</p>"
	portfolioDesc           = "<p class='portfolio-description'>%v</p>"
)

func portfolioRequest(w http.ResponseWriter, r *http.Request) {
	proj, err := blogApp.Projects(r.URL.Query().Get("lang"))
	if err != nil {
		log.Println("error getting portfolio projects:", err)
		w.WriteHeader(http.StatusInternalServerError)
		sendIndexWithContent(w, "Error getting portfolio")
		return
	}
	out := ""
	for _, p := range proj {
		out += fmt.Sprintf(portfolioTitle, p.Title)
		out += fmt.Sprintf(portfolioLink, p.Repository, p.Repository)
		for _, l := range p.Languages {
			out += fmt.Sprintf(portfolioLanguage, l.Language, l.Dates)
		}
		out += fmt.Sprintf(portfolioDesc, p.Description)
	}
	sendIndexWithContent(w, out)
}
