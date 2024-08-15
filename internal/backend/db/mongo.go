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

func (m *MongoTable[T]) Get(ID string) (data *T, err error) {
	res := m.col.FindOne(context.Background(), bson.M{"_id": ID})
	if res.Err() == mongo.ErrNoDocuments {
		return nil, backend.ErrNotFound
	} else if res.Err() != nil {
		return nil, res.Err()
	}
	var out T
	err = res.Decode(&out)
	return &out, err
}

func (m *MongoTable[T]) Find(values map[string]any) ([]T, error) {
	res, err := m.col.Find(context.Background(), values)
	if err == mongo.ErrNoDocuments {
		return nil, backend.ErrNotFound
	} else if err != nil {
		return nil, err
	}
	var out []T
	err = res.All(context.Background(), &out)
	return out, err
}

func (m *MongoTable[T]) Insert(data T) error {
	_, err := m.col.InsertOne(context.Background(), data)
	return err
}

func (m *MongoTable[T]) Remove(ID string) error {
	res := m.col.FindOneAndDelete(context.Background(), bson.M{"_id": ID})
	return res.Err()
}

func (m *MongoTable[T]) FullUpdate(ID string, data T) error {
	res := m.col.FindOneAndReplace(context.Background(), bson.M{"_id": ID}, data)
	if res.Err() == mongo.ErrNoDocuments {
		return backend.ErrNotFound
	}
	return res.Err()
}

func (m *MongoTable[T]) PartUpdate(ID string, update map[string]any) error {
	res := m.col.FindOneAndUpdate(context.Background(), bson.M{"_id": ID}, bson.M{"$set": update})
	if res.Err() == mongo.ErrNoDocuments {
		return backend.ErrNotFound
	}
	return res.Err()
}

func (m *MongoTable[CountLog]) RemoveOldLogs(date int) error {
	_, err := m.col.DeleteMany(context.Background(), bson.M{"date": bson.M{"$lt": date}})
	if err == mongo.ErrNoDocuments {
		return nil
	}
	return err
}
func (m *MongoTable[CountLog]) Count(platform string) (int, error) {
	var filter bson.M
	if platform == "" || platform == "all" {
		filter = bson.M{}
	} else {
		filter = bson.M{"platform": platform}
	}
	out, err := m.col.CountDocuments(context.Background(), filter)
	return int(out), err
}
