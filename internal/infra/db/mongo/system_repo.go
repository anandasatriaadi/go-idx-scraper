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

var _ system.Repository = (*SystemRepository)(nil)

func NewSystemRepository(db *mongo.Database) system.Repository {
	return &SystemRepository{
		collection: db.Collection("last_runs"),
	}
}

func (r *SystemRepository) GetLastRun(ctx context.Context, scriptName string) (*system.LastRun, error) {
	var model system.LastRun
	err := r.collection.FindOne(ctx, bson.M{"scriptName": scriptName}).Decode(&model)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &model, nil
}

func (r *SystemRepository) SaveLastRun(ctx context.Context, lastRun *system.LastRun) error {
	now := time.Now()
	if lastRun.CreatedAt.IsZero() {
		lastRun.CreatedAt = now
	}
	lastRun.UpdatedAt = now

	filter := bson.M{"scriptName": lastRun.ScriptName}
	update := bson.M{
		"$set": bson.M{
			"scriptName": lastRun.ScriptName,
			"lastRunAt":  lastRun.LastRunAt,
			"metadata":   lastRun.Metadata,
			"updatedAt":  now,
		},
		"$setOnInsert": bson.M{
			"createdAt": lastRun.CreatedAt,
		},
	}
	opts := options.UpdateOne().SetUpsert(true)
	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

