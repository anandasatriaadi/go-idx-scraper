package mongo

import (
	"context"
	"log"
	"time"

	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/xbrl"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type XBRLRepository struct {
	collection *mongo.Collection
}

func NewXBRLRepository(db *mongo.Database) xbrl.Repository {
	repo := &XBRLRepository{
		collection: db.Collection("xbrl_statements"),
	}
	if err := repo.ensureIndexes(context.Background()); err != nil {
		log.Printf("warn: failed to ensure xbrl indexes: %v", err)
	}
	return repo
}

func (r *XBRLRepository) ensureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "ticker", Value: 1},
				{Key: "year", Value: -1},
				{Key: "period", Value: -1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "valuation.margin_of_safety_pct", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "computed_ratios.roic", Value: -1},
			},
		},
	}
	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

func (r *XBRLRepository) Upsert(ctx context.Context, s *xbrl.Statement) error {
	now := time.Now()
	if s.CreatedAt.IsZero() {
		s.CreatedAt = now
	}
	s.UpdatedAt = now

	filter := bson.M{
		"ticker": s.Ticker,
		"year":   s.Year,
		"period": s.Period,
	}

	opts := options.UpdateOne().SetUpsert(true)
	_, err := r.collection.UpdateOne(ctx, filter, bson.M{"$set": s}, opts)
	return err
}

func (r *XBRLRepository) FindByTickerAndPeriod(ctx context.Context, ticker string, year int, period string) (*xbrl.Statement, error) {
	filter := bson.M{
		"ticker": ticker,
		"year":   year,
		"period": period,
	}
	var stmt xbrl.Statement
	err := r.collection.FindOne(ctx, filter).Decode(&stmt)
	if err != nil {
		return nil, err
	}
	return &stmt, nil
}

func (r *XBRLRepository) FindHistoricalByTicker(ctx context.Context, ticker string, limit int) ([]*xbrl.Statement, error) {
	filter := bson.M{"ticker": ticker}
	opts := options.Find().SetSort(bson.D{{Key: "year", Value: -1}, {Key: "period", Value: -1}}).SetLimit(int64(limit))
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*xbrl.Statement
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (r *XBRLRepository) FindAllWithFilter(ctx context.Context, filter any, opts ...options.Lister[options.FindOptions]) ([]*xbrl.Statement, error) {
	cursor, err := r.collection.Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*xbrl.Statement
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}
