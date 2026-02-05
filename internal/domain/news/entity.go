package news

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type News struct {
	ID       bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Title    string        `bson:"title" json:"title"`
	Date     time.Time     `bson:"date" json:"date"`
	Summary  string        `bson:"summary" json:"summary"`
	Content  string        `bson:"content" json:"content"`
	Priority int           `bson:"priority" json:"priority"`
	Link     string        `bson:"link" json:"link"`
}

type Repository interface {
	Create(ctx context.Context, news *News) error
	FindAll(ctx context.Context, filter any, opts ...options.Lister[options.FindOptions]) ([]*News, error)
	FindByID(ctx context.Context, id bson.ObjectID) (*News, error)
	UpdateByID(ctx context.Context, id bson.ObjectID, update any) error
}
