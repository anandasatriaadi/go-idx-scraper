package mongo

import (
	"context"
	"time"

	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/news"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// NewsRepository defines strict repository methods
type NewsRepository struct {
	collection *mongo.Collection
}

var _ news.Repository = (*NewsRepository)(nil)

// NewNewsRepository creates a new repository instance
func NewNewsRepository(db *mongo.Database) news.Repository {
	return &NewsRepository{
		collection: db.Collection("news"),
	}
}

// Create inserts a single News
func (r *NewsRepository) Create(ctx context.Context, model *news.News) error {
	now := time.Now()
	if model.CreatedAt.IsZero() {
		model.CreatedAt = now
	}
	model.UpdatedAt = now
	if model.Status == "" {
		model.Status = news.StatusPending
	}
	_, err := r.collection.InsertOne(ctx, model)
	return err
}

func idFilter(id string) bson.M {
	if objID, err := bson.ObjectIDFromHex(id); err == nil {
		return bson.M{"$or": []bson.M{{"_id": objID}, {"_id": id}}}
	}
	return bson.M{"_id": id}
}

// FindByID retrieves a document by its _id
func (r *NewsRepository) FindByID(ctx context.Context, id string) (*news.News, error) {
	var model news.News
	err := r.collection.FindOne(ctx, idFilter(id)).Decode(&model)
	if err != nil {
		return nil, err
	}
	return &model, nil
}

// UpdateSummary updates structured summary fields for a news item
func (r *NewsRepository) UpdateSummary(ctx context.Context, id string, summary *news.NewsSummaryUpdate) error {
	setFields := bson.M{
		"updated_at": time.Now(),
	}
	if summary.Title != "" {
		setFields["title"] = summary.Title
	}
	if summary.Summary != "" {
		setFields["summary"] = summary.Summary
	}
	if summary.Priority != 0 {
		setFields["priority"] = summary.Priority
	}
	setFields["value_score"] = summary.ValueScore
	if summary.ImpactDirection != "" {
		setFields["impact_direction"] = summary.ImpactDirection
	}
	if summary.InvestmentTakeaway != "" {
		setFields["investment_takeaway"] = summary.InvestmentTakeaway
	}
	if summary.Tickers != nil {
		setFields["tickers"] = summary.Tickers
	}
	if summary.Sector != "" {
		setFields["sector"] = summary.Sector
	}
	if summary.Subsector != "" {
		setFields["subsector"] = summary.Subsector
	}
	if summary.Industry != "" {
		setFields["industry"] = summary.Industry
	}
	setFields["is_industry_wide"] = summary.IsIndustryWide
	if summary.Status != "" {
		setFields["status"] = summary.Status
	}

	_, err := r.collection.UpdateOne(ctx, idFilter(id), bson.M{"$set": setFields})
	return err
}

// ExistsByLink checks if an article with the given link already exists in MongoDB
func (r *NewsRepository) ExistsByLink(ctx context.Context, link string) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"link": link}, options.Count().SetLimit(1))
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// FindPendingSummary retrieves news articles that need summarization
func (r *NewsRepository) FindPendingSummary(ctx context.Context, limit int) ([]*news.News, error) {
	filter := bson.M{
		"$and": []bson.M{
			{"status": bson.M{"$ne": news.StatusSummarized}},
			{
				"$or": []bson.M{
					{"status": news.StatusPending},
					{"status": news.StatusFailed},
					{"summary": ""},
					{"summary": bson.M{"$exists": false}},
				},
			},
		},
	}
	opts := options.Find().SetSort(bson.D{{Key: "date", Value: -1}}).SetLimit(int64(limit))
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*news.News
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

// FindRecent retrieves the latest news articles up to limit
func (r *NewsRepository) FindRecent(ctx context.Context, limit int) ([]*news.News, error) {
	opts := options.Find().SetSort(bson.D{{Key: "date", Value: -1}}).SetLimit(int64(limit))
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*news.News
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

// FindBetweenDates retrieves news articles published between start and end dates
func (r *NewsRepository) FindBetweenDates(ctx context.Context, start, end time.Time) ([]*news.News, error) {
	filter := bson.M{
		"created_at": bson.M{
			"$gte": start,
			"$lte": end,
		},
	}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*news.News
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}
