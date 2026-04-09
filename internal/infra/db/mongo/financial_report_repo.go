package mongo

import (
	"context"
	"log"
	"time"

	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/finreport"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type FinancialReportRepository struct {
	collection *mongo.Collection
}

func NewFinancialReportRepository(db *mongo.Database) finreport.Repository {
	repo := &FinancialReportRepository{
		collection: db.Collection("financial_reports"),
	}
	if err := repo.ensureIndexes(context.Background()); err != nil {
		log.Printf("warn: failed to ensure financial report indexes: %v", err)
	}
	return repo
}

func (r *FinancialReportRepository) ensureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "issuer_code", Value: 1},
				{Key: "year", Value: 1},
				{Key: "period_string", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "is_latest", Value: 1}},
		},
	}

	if _, err := r.collection.Indexes().CreateMany(ctx, indexes); err != nil {
		return err
	}
	return nil
}

func (r *FinancialReportRepository) Create(ctx context.Context, model *finreport.FinancialReport) error {
	now := time.Now()
	if model.CreatedAt.IsZero() {
		model.CreatedAt = now
	}
	model.UpdatedAt = now
	_, err := r.collection.InsertOne(ctx, model)
	return err
}

func (r *FinancialReportRepository) FindAll(ctx context.Context, filter any, opts ...options.Lister[options.FindOptions]) ([]*finreport.FinancialReport, error) {
	cursor, err := r.collection.Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*finreport.FinancialReport
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (r *FinancialReportRepository) FindOne(ctx context.Context, filter any) (*finreport.FinancialReport, error) {
	var result finreport.FinancialReport
	err := r.collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &result, nil
}

func (r *FinancialReportRepository) UpdateOne(ctx context.Context, filter, update any) error {
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}
