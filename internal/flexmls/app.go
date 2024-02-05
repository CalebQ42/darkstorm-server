package flexmls

import (
	"github.com/CalebQ42/stupid-backend/v2/defaultapp"
	"go.mongodb.org/mongo-driver/mongo"
)

type FlexMLS struct {
	*defaultapp.App
}

func NewBackend(client *mongo.Client) *FlexMLS {
	return &FlexMLS{defaultapp.NewDefaultApp(client.Database("flexmls"))}
}
