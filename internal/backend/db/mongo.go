package db

import (
	"context"

	"github.com/CalebQ42/darkstorm-server/internal/backend"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoTable[T backend.IDStruct] struct {
	col *mongo.Collection
}

func NewMongoTable[T backend.IDStruct](col *mongo.Collection) *MongoTable[T] {
	return &MongoTable[T]{
		col: col,
	}
}

func (m *MongoTable[T]) Get(ctx context.Context, ID string) (data *T, err error) {
	res := m.col.FindOne(ctx, bson.M{"_id": ID})
	if res.Err() == mongo.ErrNoDocuments {
		return nil, backend.ErrNotFound
	} else if res.Err() != nil {
		return nil, res.Err()
	}
	var out T
	err = res.Decode(&out)
	return &out, err
}

func (m *MongoTable[T]) Find(ctx context.Context, values map[string]any) ([]T, error) {
	res, err := m.col.Find(ctx, values)
	if err == mongo.ErrNoDocuments {
		return nil, backend.ErrNotFound
	} else if err != nil {
		return nil, err
	}
	var out []T
	err = res.All(ctx, &out)
	if len(out) == 0 {
		return nil, backend.ErrNotFound
	}
	return out, err
}

func (m *MongoTable[T]) Insert(ctx context.Context, data T) error {
	_, err := m.col.InsertOne(ctx, data)
	return err
}

func (m *MongoTable[T]) Remove(ctx context.Context, ID string) error {
	res := m.col.FindOneAndDelete(ctx, bson.M{"_id": ID})
	return res.Err()
}

func (m *MongoTable[T]) FullUpdate(ctx context.Context, ID string, data T) error {
	res := m.col.FindOneAndReplace(ctx, bson.M{"_id": ID}, data)
	if res.Err() == mongo.ErrNoDocuments {
		return backend.ErrNotFound
	}
	return res.Err()
}

func (m *MongoTable[T]) PartUpdate(ctx context.Context, ID string, update map[string]any) error {
	res := m.col.FindOneAndUpdate(ctx, bson.M{"_id": ID}, bson.M{"$set": update})
	if res.Err() == mongo.ErrNoDocuments {
		return backend.ErrNotFound
	}
	return res.Err()
}

func (m *MongoTable[CountLog]) RemoveOldLogs(ctx context.Context, date int) error {
	_, err := m.col.DeleteMany(ctx, bson.M{"date": bson.M{"$lt": date}})
	if err == mongo.ErrNoDocuments {
		return nil
	}
	return err
}
func (m *MongoTable[CountLog]) Count(ctx context.Context, platform string) (int, error) {
	var filter bson.M
	if platform == "" || platform == "all" {
		filter = bson.M{}
	} else {
		filter = bson.M{"platform": platform}
	}
	out, err := m.col.CountDocuments(ctx, filter)
	return int(out), err
}
