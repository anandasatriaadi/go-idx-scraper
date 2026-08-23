package news

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type News struct {
	CreatedAt          time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt          time.Time     `bson:"updated_at" json:"updated_at"`
	Date               time.Time     `bson:"date" json:"date"`
	Title              string        `bson:"title" json:"title"`
	Summary            string        `bson:"summary" json:"summary"`
	Content            string        `bson:"content" json:"content"`
	Link               string        `bson:"link" json:"link"`
	Priority           int           `bson:"priority" json:"priority"`
	ValueScore         int           `bson:"value_score" json:"value_score"`
	ImpactDirection    string        `bson:"impact_direction" json:"impact_direction"`
	InvestmentTakeaway string        `bson:"investment_takeaway" json:"investment_takeaway"`
	ID                 bson.ObjectID `bson:"_id,omitempty" json:"id"`
}

// Port: Repository defines data persistence interface (Output Port)
type Repository interface {
	Create(ctx context.Context, news *News) error
	FindAll(ctx context.Context, filter any, opts ...options.Lister[options.FindOptions]) ([]*News, error)
	FindByID(ctx context.Context, id bson.ObjectID) (*News, error)
	UpdateByID(ctx context.Context, id bson.ObjectID, update any) error
}

// Port: Scraper defines news scraping interface (Input Port/External Service)
type Scraper interface {
	Scrape(ctx context.Context, startDate, endDate time.Time, onNewsFound func(*News) error) error
}
