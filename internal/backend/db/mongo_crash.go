package db

import (
	"context"
	"strings"

	"github.com/CalebQ42/darkstorm-server/internal/backend"
	"github.com/google/uuid"
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

func (m *MongoCrashTable) Archive(ctx context.Context, toArchive backend.ArchivedCrash) error {
	if toArchive.Platform == "" {
		toArchive.Platform = "all"
	}
	_, err := m.archiveCol.InsertOne(ctx, toArchive)
	return err
}

func (m *MongoCrashTable) IsArchived(ctx context.Context, ind backend.IndividualCrash) bool {
	res := m.archiveCol.FindOne(ctx,
		bson.M{"error": ind.Error, "stack": ind.Stack, "platform": bson.M{"$in": []string{ind.Platform, "all"}}},
	)
	return res.Err() == nil
}

func (m *MongoCrashTable) InsertCrash(ctx context.Context, ind backend.IndividualCrash) error {
	first, _, _ := strings.Cut(ind.Stack, "\n")
	res, err := m.col.UpdateOne(ctx,
		bson.M{"error": ind.Error, "firstLine": first, //filter main report
			"individual": bson.M{"$elemMatch": bson.M{"stack": ind.Stack, "platform": ind.Platform}}}, //filter individual
		bson.M{"$inc": bson.M{"individual.$.count": 1}}, //increment count
	)
	if err != nil && err != mongo.ErrNoDocuments {
		return err
	}
	if err == mongo.ErrNoDocuments || res.MatchedCount == 0 {
		ind.Count = 1
		res, err = m.col.UpdateMany(ctx,
			bson.M{"error": ind.Error, "firstLine": first}, //filter
			bson.M{"$push": bson.M{"individual": ind}},     //Add new individual report
		)
		if err != nil && err != mongo.ErrNoDocuments {
			return err
		}
		if err == mongo.ErrNoDocuments || res.MatchedCount == 0 {
			var id uuid.UUID
			id, err = uuid.NewV7()
			if err != nil {
				return err
			}
			ind.Count = 1
			_, err = m.col.InsertOne(ctx,
				backend.CrashReport{
					ID:         id.String(),
					Error:      ind.Error,
					FirstLine:  first,
					Individual: []backend.IndividualCrash{ind},
				},
			)
		}
	}
	return err
}
