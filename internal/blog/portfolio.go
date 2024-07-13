package blog

import (
	"go.mongodb.org/mongo-driver/mongo"
)

type project struct {
	Title       string `bson:"_id"`
	Repository  string `bson:"respository"`
	Description string `bson:"description"`
	Languages   []struct {
		Language string `bson:"language"`
		Dates    string `bson:"dates"`
	} `bson:"language"`
}

func portfolio(client *mongo.Client) {
	//TODO
}
