package blog

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/CalebQ42/darkstorm-server/internal/backend"
	"go.mongodb.org/mongo-driver/bson"
)

type PortfolioProject struct {
	Title        string   `json:"_id" bson:"_id"`
	Repository   string   `json:"repository" bson:"repository"`
	Description  string   `json:"description" bson:"description"`
	Technologies []string `json:"technologies" bson:"technologies"`
	Languages    []struct {
		Language string `json:"language" bson:"language"`
		Dates    string `json:"dates" bson:"dates"`
	} `json:"language" bson:"language"`
}

func (b *BlogApp) Projects(languageFilter string) ([]PortfolioProject, error) {
	filter := bson.M{}
	if languageFilter != "" {
		filter["language.language"] = languageFilter
	}
	res, err := b.portfolioCol.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	var out []PortfolioProject
	err = res.All(context.Background(), &out)
	return out, err
}

func (b *BlogApp) reqPortfolio(w http.ResponseWriter, r *http.Request) {
	folio, err := b.Projects(r.URL.Query().Get("lang"))
	if err != nil {
		backend.ReturnError(w, http.StatusInternalServerError, "internal", "Server Error")
		return
	}
	json.NewEncoder(w).Encode(folio)
}
