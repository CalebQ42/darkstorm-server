package blog

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
