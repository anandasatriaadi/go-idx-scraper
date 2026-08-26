package mongo

import (
	"context"
	"time"

	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/news"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type BriefingRepository struct {
	collection *mongo.Collection
}

var _ news.BriefingRepository = (*BriefingRepository)(nil)

func NewBriefingRepository(db *mongo.Database) news.BriefingRepository {
	return &BriefingRepository{
		collection: db.Collection("daily_briefings"),
	}
}

func (r *BriefingRepository) Create(ctx context.Context, model *news.Briefing) error {
	now := time.Now()
	if model.CreatedAt.IsZero() {
		model.CreatedAt = now
	}
	model.UpdatedAt = now
	_, err := r.collection.InsertOne(ctx, model)
	return err
}

func (r *BriefingRepository) FindByDate(ctx context.Context, date time.Time) (*news.Briefing, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.AddDate(0, 0, 1)

	filter := bson.M{
		"date": bson.M{
			"$gte": startOfDay,
			"$lt":  endOfDay,
		},
	}
	var result news.Briefing
	err := r.collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *BriefingRepository) FindLatest(ctx context.Context) (*news.Briefing, error) {
	opts := options.FindOne().SetSort(bson.D{{Key: "date", Value: -1}})
	var result news.Briefing
	err := r.collection.FindOne(ctx, bson.M{}, opts).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *BriefingRepository) FindRecent(ctx context.Context, limit int) ([]*news.Briefing, error) {
	opts := options.Find().SetSort(bson.D{{Key: "date", Value: -1}}).SetLimit(int64(limit))
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*news.Briefing
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}
