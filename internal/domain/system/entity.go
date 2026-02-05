package system

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type LastRun struct {
	ID         bson.ObjectID `bson:"_id,omitempty" json:"id"`
	ScriptName string        `bson:"scriptName" json:"scriptName"`
	LastRunAt  time.Time     `bson:"lastRunAt" json:"lastRunAt"`
	Metadata   bson.M        `bson:"metadata,omitempty" json:"metadata,omitempty"`
}

type Repository interface {
	FindOne(ctx context.Context, filter interface{}, opts ...options.Lister[options.FindOneOptions]) (*LastRun, error)
	UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...options.Lister[options.UpdateOneOptions]) error
}
