package blog

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/CalebQ42/darkstorm-server/internal/backend"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	portfolioTitle    = "<h2 class='portfolio-title'>%v</h2>"
	portfolioLink     = "<p class='portfolio-link'><a href='%v'>%v</a>"
	portfolioLanguage = "<p class='portfolio-language'><b>%v</b>: %v</p>"
	portfolioTech     = "<p class='portfolio-tech'><b>Tech: </b>%v</p>"
	portfolioDesc     = "<p class='portfolio-description'>%v</p>"
)

type PortfolioProject struct {
	Title        string   `json:"_id" bson:"_id"`
	Order        int      `json:"order" bson:"order"`
	Repository   string   `json:"repository" bson:"repository"`
	Description  string   `json:"description" bson:"description"`
	Technologies []string `json:"technologies" bson:"technologies"`
	Languages    []struct {
		Language string `json:"language" bson:"language"`
		Dates    string `json:"dates" bson:"dates"`
	} `json:"language" bson:"language"`
}

func (p PortfolioProject) HTMX() string {
	out := fmt.Sprintf(portfolioTitle, p.Title)
	out += fmt.Sprintf(portfolioLink, p.Repository, p.Repository)
	for _, l := range p.Languages {
		out += fmt.Sprintf(portfolioLanguage, l.Language, l.Dates)
	}
	out += fmt.Sprintf(portfolioTech, strings.Join(p.Technologies, ", "))
	out += fmt.Sprintf(portfolioDesc, p.Description)
	return out
}

type Portfolio []PortfolioProject

const (
	portfolioSelector = `
<p>Tech Filter:
	<select
	id='techSelect'
	name='tech'
	hx-get='http://localhost:2323/blog/portfolio'
	hx-target='#projects'>
		%v
	</select>
</p>`
	portfolioSelectorOption = "<option value='%v'%v>%v</option>"
)

func (p Portfolio) FullHTMX(ctx context.Context, blogApp *BlogApp, selectedTech string) string {
	aboutMe := "<h1 class='about-me-header'>About Me</h1>"
	if me, err := blogApp.AboutMe(ctx); err != nil {
		aboutMe += "Error getting info about me :("
	} else {
		aboutMe += me.HTML()
	}
	aboutMe += "<h1 class='my-projects-header' style='margin-bottom:15px'>My Projects</h1>"
	tech := make(map[string]struct{})
	for i := range p {
		for _, t := range p[i].Technologies {
			tech[t] = struct{}{}
		}
	}
	techKeys := make([]string, 0, len(tech))
	for k := range tech {
		techKeys = append(techKeys, k)
	}
	slices.Sort(techKeys)
	var out string
	if selectedTech == "" {
		out = fmt.Sprintf(portfolioSelectorOption, "", " selected=true", "All")
	} else {
		out = fmt.Sprintf(portfolioSelectorOption, "", "", "All")
	}
	for _, k := range techKeys {
		if selectedTech == strings.ToLower(k) {
			out += fmt.Sprintf(portfolioSelectorOption, k, " selected=true", k)
		} else {
			out += fmt.Sprintf(portfolioSelectorOption, k, "", k)
		}
	}
	return aboutMe + fmt.Sprintf(portfolioSelector, out) + "<div id='projects'>" + p.HTMX() + "</div>"
}

func (p Portfolio) HTMX() string {
	out := ""
	for _, proj := range p {
		out += proj.HTMX()
	}
	return out
}

func (b *BlogApp) Projects(ctx context.Context, techFilter string) (Portfolio, error) {
	filter := bson.M{}
	if techFilter != "" {
		filter = bson.M{"technologies": techFilter}
	}
	res, err := b.portfolioCol.Find(ctx, filter, options.Find().SetSort(bson.M{"order": 1}))
	if err != nil {
		return nil, err
	}
	var out []PortfolioProject
	err = res.All(ctx, &out)
	return out, err
}

func (b *BlogApp) reqPortfolio(w http.ResponseWriter, r *http.Request) {
	folio, err := b.Projects(r.Context(), r.URL.Query().Get("tech"))
	if err != nil {
		backend.ReturnError(w, http.StatusInternalServerError, "internal", "Server Error")
		return
	}
	if r.Header.Get("Hx-Request") == "true" {
		if r.URL.Query().Has("tech") {
			w.Write([]byte(folio.HTMX()))
		} else {
			w.Write([]byte(folio.FullHTMX(r.Context(), b, r.URL.Query().Get("tech"))))
		}
	} else {
		json.NewEncoder(w).Encode(folio)
	}
}
