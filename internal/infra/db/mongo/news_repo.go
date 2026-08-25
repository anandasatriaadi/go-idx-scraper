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

// FindByID retrieves a document by its _id
func (r *NewsRepository) FindByID(ctx context.Context, id bson.ObjectID) (*news.News, error) {
	var model news.News
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&model)
	if err != nil {
		return nil, err
	}
	return &model, nil
}

// FindAll retrieves all documents matching a filter
func (r *NewsRepository) FindAll(ctx context.Context, filter any, opts ...options.Lister[options.FindOptions]) ([]*news.News, error) {
	cursor, err := r.collection.Find(ctx, filter, opts...)
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

// ExistsByLink checks if an article with the given link already exists in MongoDB
func (r *NewsRepository) ExistsByLink(ctx context.Context, link string) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"link": link}, options.Count().SetLimit(1))
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// FindPendingSummary retrieves news articles that have not yet been summarized (status is pending or missing/empty)
func (r *NewsRepository) FindPendingSummary(ctx context.Context, limit int) ([]*news.News, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"status": news.StatusPending},
			{"status": bson.M{"$exists": false}},
			{"status": ""},
			{"summary": ""},
		},
	}
	opts := options.Find().SetSort(bson.D{{Key: "date", Value: -1}}).SetLimit(int64(limit))
	return r.FindAll(ctx, filter, opts)
}

// UpdateByID updates a document by ID
func (r *NewsRepository) UpdateByID(ctx context.Context, id bson.ObjectID, update any) error {
	var updateDoc bson.M
	switch v := update.(type) {
	case bson.M:
		updateDoc = v
	case map[string]any:
		updateDoc = bson.M(v)
	default:
		_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
		return err
	}

	if setVal, exists := updateDoc["$set"]; exists {
		switch s := setVal.(type) {
		case bson.M:
			s["updated_at"] = time.Now()
		case map[string]any:
			s["updated_at"] = time.Now()
		}
	} else {
		updateDoc["$set"] = bson.M{"updated_at": time.Now()}
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, updateDoc)
	return err
}
