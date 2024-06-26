package db

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/CalebQ42/darkstorm-server/internal/backend"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoCrashTable struct {
	*MongoTable[backend.CrashReport]
	archiveCol *mongo.Collection
}

func NewMongoCrashTable(crashCol *mongo.Collection, archiveCol *mongo.Collection) *MongoCrashTable {
	return &MongoCrashTable{
		MongoTable: NewMongoTable[backend.CrashReport](crashCol),
		archiveCol: archiveCol,
	}
}

func (m *MongoCrashTable) Archive(toArchive backend.ArchivedCrash) error {
	if toArchive.Platform == "" {
		toArchive.Platform = "all"
	}
	_, err := m.archiveCol.InsertOne(context.Background(), toArchive)
	return err
}

func (m *MongoCrashTable) IsArchived(ind backend.IndividualCrash) bool {
	res := m.archiveCol.FindOne(context.Background(),
		bson.M{"error": ind.Error, "stack": ind.Stack, "platform": bson.M{"$in": []string{ind.Platform, "all"}}},
	)
	return res.Err() == nil
}

func (m *MongoCrashTable) InsertCrash(ind backend.IndividualCrash) error {
	first, _, _ := strings.Cut(ind.Stack, "\n")
	findRes := m.col.FindOne(context.Background(),
		bson.M{"error": ind.Error, "firstLine": first, //filter main report
			"individual.stack": ind.Stack, "individual.platform": ind.Platform}, //filter individual
		// bson.M{"$inc": bson.M{"individual.count": 1}}, //increment count
	)
	var out map[string]any
	findRes.Decode(&out)
	fmt.Println(out)
	return errors.New("STUFF")
	// if err != nil && err != mongo.ErrNoDocuments {
	// 	return err
	// }
	// if err == mongo.ErrNoDocuments || res.MatchedCount == 0 {
	// 	ind.Count = 1
	// 	res, err = m.col.UpdateMany(context.Background(),
	// 		bson.M{"error": ind.Error, "firstLine": first}, //filter
	// 		bson.M{"$push": bson.M{"individual": ind}},     //Add new individual report
	// 	)
	// 	if err != nil && err != mongo.ErrNoDocuments {
	// 		return err
	// 	}
	// 	if err == mongo.ErrNoDocuments || res.MatchedCount == 0 {
	// 		var id uuid.UUID
	// 		id, err = uuid.NewV7()
	// 		if err != nil {
	// 			return err
	// 		}
	// 		_, err = m.col.InsertOne(context.Background(),
	// 			backend.CrashReport{
	// 				ID:         id.String(),
	// 				Error:      ind.Error,
	// 				FirstLine:  first,
	// 				Individual: []backend.IndividualCrash{ind},
	// 			},
	// 		)
	// 	}
	// }
	// return err
}
