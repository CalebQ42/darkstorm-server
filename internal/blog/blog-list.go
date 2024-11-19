package blog

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BlogList struct {
	ID         string `json:"id" bson:"_id"`
	Title      string `json:"title" bson:"title"`
	Draft      bool   `json:"draft" bson:"draft"`
	CreateTime int64  `json:"createTime" bson:"createTime"`
}

func (b *Backend) FullBlogList(ctx context.Context) ([]BlogList, error) {
	res, err := b.blogCol.Find(ctx, bson.M{}, options.Find().
		SetProjection(bson.M{"_id": 1, "createTime": 1, "title": 1, "draft": 1}).
		SetSort(bson.M{"createTime": -1}))
	if err != nil {
		return nil, err
	}
	var list []BlogList
	err = res.All(ctx, &list)
	return list, err
}

func (b *Backend) BlogList(ctx context.Context) ([]BlogList, error) {
	res, err := b.blogCol.Find(ctx, bson.M{
		"draft":      false,
		"staticPage": false,
	}, options.Find().
		SetProjection(bson.M{"_id": 1, "createTime": 1, "title": 1, "draft": 1}).
		SetSort(bson.M{"createTime": -1}))
	if err != nil {
		return nil, err
	}
	var list []BlogList
	err = res.All(ctx, &list)
	return list, err
}
