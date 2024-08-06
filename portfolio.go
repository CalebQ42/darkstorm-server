package main

import (
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"
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
	selectedLang := r.URL.Query().Get("lang")
	proj, err := blogApp.Projects(selectedLang)
	if err != nil {
		log.Println("error getting portfolio projects:", err)
		w.WriteHeader(http.StatusInternalServerError)
		sendContent(w, r, "Error getting portfolio", "", "")
		return
	}
	aboutMe := "<h1 class='about-me-header'>About Me</h1>"
	if me, err := blogApp.AboutMe(); err != nil {
		aboutMe += "Error getting info about me :("
	} else {
		aboutMe += authorSection(me)
	}
	aboutMe += "<h1 class='my-projects-header' style='margin-bottom:15px'>My Projects</h1>"
	langs := make(map[string]struct{})
	out := ""
	for _, p := range proj {
		out += fmt.Sprintf(portfolioTitle, p.Title)
		out += fmt.Sprintf(portfolioLink, p.Repository, p.Repository)
		for _, l := range p.Languages {
			langs[l.Language] = struct{}{}
			out += fmt.Sprintf(portfolioLanguage, l.Language, l.Dates)
		}
		out += fmt.Sprintf(portfolioDesc, p.Description)
	}
	langKeys := make([]string, 0, len(langs))
	for k := range langs {
		langKeys = append(langKeys, k)
	}
	slices.Sort(langKeys)
	var tmp string
	if selectedLang == "" {
		tmp = fmt.Sprintf(portfolioSelectorOption, "", " selected=true", "All")
	} else {
		tmp = fmt.Sprintf(portfolioSelectorOption, "", "", "All")
	}
	for _, k := range langKeys {
		if selectedLang == strings.ToLower(k) {
			tmp += fmt.Sprintf(portfolioSelectorOption, strings.ToLower(k), " selected=true", k)
		} else {
			tmp += fmt.Sprintf(portfolioSelectorOption, strings.ToLower(k), "", k)
		}
	}
	out = aboutMe + fmt.Sprintf(portfolioSelector, tmp) + out
	sendContent(w, r, out, "Portfolio", "")
}
