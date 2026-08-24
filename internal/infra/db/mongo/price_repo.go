package mongo

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/stock"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// PriceRepository implements stock.PriceRepository for MongoDB.
type PriceRepository struct {
	collection *mongo.Collection
}

var _ stock.PriceRepository = (*PriceRepository)(nil)

// NewPriceRepository initializes PriceRepository and ensures indexes.
func NewPriceRepository(db *mongo.Database) stock.PriceRepository {
	repo := &PriceRepository{
		collection: db.Collection("stock_prices"),
	}
	if err := repo.ensureIndexes(context.Background()); err != nil {
		log.Printf("warn: failed to ensure stock_prices indexes: %v", err)
	}
	return repo
}

func (r *PriceRepository) ensureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "ticker", Value: 1},
				{Key: "date", Value: -1},
			},
			Options: options.Index().SetUnique(true),
		},
	}
	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

// UpsertCandles bulk upserts daily price candles for the given ticker.
func (r *PriceRepository) UpsertCandles(ctx context.Context, ticker string, candles []stock.PriceCandle) error {
	if len(candles) == 0 {
		return nil
	}

	cleanTicker := strings.ToUpper(strings.TrimSpace(ticker))
	if strings.HasSuffix(cleanTicker, ".JK") {
		cleanTicker = strings.TrimSuffix(cleanTicker, ".JK")
	}

	now := time.Now().UTC()
	models := make([]mongo.WriteModel, 0, len(candles))

	for _, c := range candles {
		candDate := c.Date.UTC()
		filter := bson.M{
			"ticker": cleanTicker,
			"date":   candDate,
		}

		update := bson.M{
			"$set": bson.M{
				"ticker":     cleanTicker,
				"date":       candDate,
				"open":       c.Open,
				"high":       c.High,
				"low":        c.Low,
				"close":      c.Close,
				"adj_close":  c.AdjClose,
				"volume":     c.Volume,
				"updated_at": now,
			},
			"$setOnInsert": bson.M{
				"created_at": now,
			},
		}

		model := mongo.NewUpdateOneModel().
			SetFilter(filter).
			SetUpdate(update).
			SetUpsert(true)

		models = append(models, model)
	}

	opts := options.BulkWrite().SetOrdered(false)
	_, err := r.collection.BulkWrite(ctx, models, opts)
	if err != nil {
		return fmt.Errorf("bulk upserting candles for %s: %w", cleanTicker, err)
	}
	return nil
}

// GetPrices retrieves historical price candles for a ticker sorted by date descending.
func (r *PriceRepository) GetPrices(ctx context.Context, ticker string, limit int) ([]*stock.PriceCandle, error) {
	cleanTicker := strings.ToUpper(strings.TrimSpace(ticker))
	if strings.HasSuffix(cleanTicker, ".JK") {
		cleanTicker = strings.TrimSuffix(cleanTicker, ".JK")
	}

	filter := bson.M{"ticker": cleanTicker}
	opts := options.Find().SetSort(bson.D{{Key: "date", Value: -1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("finding prices for %s: %w", cleanTicker, err)
	}
	defer cursor.Close(ctx)

	var results []*stock.PriceCandle
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("decoding prices for %s: %w", cleanTicker, err)
	}
	return results, nil
}
