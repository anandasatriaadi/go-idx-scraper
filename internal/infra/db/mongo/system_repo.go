package mongo

import (
	"context"
	"time"

	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/system"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type SystemRepository struct {
	collection *mongo.Collection
}

func NewSystemRepository(db *mongo.Database) system.Repository {
	return &SystemRepository{
		collection: db.Collection("last_runs"),
	}
}

func (r *SystemRepository) FindOne(ctx context.Context, filter any, opts ...options.Lister[options.FindOneOptions]) (*system.LastRun, error) {
	var model system.LastRun
	err := r.collection.FindOne(ctx, filter, opts...).Decode(&model)
	if err != nil {
		return nil, err
	}
	return &model, nil
}

func (r *SystemRepository) UpdateOne(ctx context.Context, filter any, update any, opts ...options.Lister[options.UpdateOneOptions]) error {
	var updateDoc bson.M
	switch v := update.(type) {
	case bson.M:
		updateDoc = v
	case map[string]interface{}:
		updateDoc = bson.M(v)
	default:
		_, err := r.collection.UpdateOne(ctx, filter, update, opts...)
		return err
	}

	if updateDoc["$set"] == nil {
		updateDoc["$set"] = bson.M{}
	}
	setMap := updateDoc["$set"].(bson.M)
	setMap["updatedAt"] = time.Now()

	if updateDoc["$setOnInsert"] == nil {
		updateDoc["$setOnInsert"] = bson.M{}
	}
	soiMap := updateDoc["$setOnInsert"].(bson.M)
	soiMap["createdAt"] = time.Now()

	_, err := r.collection.UpdateOne(ctx, filter, updateDoc, opts...)
	return err
}
