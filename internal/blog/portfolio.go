package blog

import (
	"context"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
)

type PortfolioProject struct {
	Title       string `bson:"_id"`
	Repository  string `bson:"repository"`
	Description string `bson:"description"`
	Languages   []struct {
		Language string `bson:"language"`
		Dates    string `bson:"dates"`
	} `bson:"language"`
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
	//TODO
}
