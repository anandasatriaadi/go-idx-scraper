package mongo

import (
	"context"
	"time"

	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/announcement"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type AnnouncementRepository struct {
	collection *mongo.Collection
}

func NewAnnouncementRepository(db *mongo.Database) announcement.Repository {
	return &AnnouncementRepository{
		collection: db.Collection("announcements"),
	}
}

func (r *AnnouncementRepository) Create(ctx context.Context, model *announcement.Announcement) error {
	now := time.Now()
	if model.CreatedAt.IsZero() {
		model.CreatedAt = now
	}
	model.UpdatedAt = now
	_, err := r.collection.InsertOne(ctx, model)
	return err
}

func (r *AnnouncementRepository) FindByID(ctx context.Context, id string) (*announcement.Announcement, error) {
	var model announcement.Announcement
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&model)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &model, nil
}

func (r *AnnouncementRepository) Exists(ctx context.Context, id string) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"_id": id}, options.Count().SetLimit(1))
	return count > 0, err
}

func (r *AnnouncementRepository) FindRecent(ctx context.Context, limit int) ([]*announcement.Announcement, error) {
	opts := options.Find().SetSort(bson.M{"created_date": -1})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*announcement.Announcement
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (r *AnnouncementRepository) GetLatestCreatedDate(ctx context.Context) (*time.Time, error) {
	opts := options.Find().SetSort(bson.M{"created_date": -1}).SetLimit(1)
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*announcement.Announcement
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	if len(results) == 0 || results[0].CreatedDate == nil {
		return nil, nil
	}
	return results[0].CreatedDate, nil
}

func (r *AnnouncementRepository) FindExistingIDs(ctx context.Context, limit int) (map[string]bool, error) {
	opts := options.Find().SetProjection(bson.D{{Key: "_id", Value: 1}}).SetSort(bson.M{"created_date": -1})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	type idDoc struct {
		ID string `bson:"_id"`
	}
	var docs []idDoc
	if err = cursor.All(ctx, &docs); err != nil {
		return nil, err
	}

	exists := make(map[string]bool, len(docs))
	for _, doc := range docs {
		exists[doc.ID] = true
	}
	return exists, nil
}
