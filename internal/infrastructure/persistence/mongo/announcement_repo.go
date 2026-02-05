package mongo

import (
	"context"

	"github.com/anandasatriaadi/go-idx-scraper/internal/domain/announcement"
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
	_, err := r.collection.InsertOne(ctx, model)
	return err
}

func (r *AnnouncementRepository) FindAll(ctx context.Context, filter interface{}, opts ...options.Lister[options.FindOptions]) ([]*announcement.Announcement, error) {
	cursor, err := r.collection.Find(ctx, filter, opts...)
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

func (r *AnnouncementRepository) FindByID(ctx context.Context, id string) (*announcement.Announcement, error) {
	var model announcement.Announcement
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&model)
	if err != nil {
		return nil, err
	}
	return &model, nil
}

func (r *AnnouncementRepository) Exists(ctx context.Context, id string) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"_id": id}, options.Count().SetLimit(1))
	return count > 0, err
}
