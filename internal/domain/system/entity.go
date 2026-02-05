package system

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type LastRun struct {
	LastRunAt  time.Time     `bson:"lastRunAt" json:"lastRunAt"`
	Metadata   bson.M        `bson:"metadata,omitempty" json:"metadata,omitempty"`
	ScriptName string        `bson:"scriptName" json:"scriptName"`
	ID         bson.ObjectID `bson:"_id,omitempty" json:"id"`
}

type Repository interface {
	FindOne(ctx context.Context, filter any, opts ...options.Lister[options.FindOneOptions]) (*LastRun, error)
	UpdateOne(ctx context.Context, filter any, update any, opts ...options.Lister[options.UpdateOneOptions]) error
}
