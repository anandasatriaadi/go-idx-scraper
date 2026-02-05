package mongo

import (
	"context"

	"github.com/anandasatriaadi/go-idx-scraper/internal/domain/finreport"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type FinancialReportRepository struct {
	collection *mongo.Collection
}

func NewFinancialReportRepository(db *mongo.Database) finreport.Repository {
	return &FinancialReportRepository{
		collection: db.Collection("financial_reports"),
	}
}

func (r *FinancialReportRepository) Create(ctx context.Context, model *finreport.FinancialReport) error {
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
