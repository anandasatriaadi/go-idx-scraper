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

var _ finreport.Repository = (*FinancialReportRepository)(nil)

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

func (r *FinancialReportRepository) FindByIssuerAndPeriod(ctx context.Context, issuerCode string, year int, periodString string) (*finreport.FinancialReport, error) {
	filter := bson.M{
		"issuer_code":   issuerCode,
		"year":          year,
		"period_string": periodString,
	}
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

func (r *FinancialReportRepository) UpdateIsLatest(ctx context.Context, issuerCode string, year int, periodString string, isLatest bool) error {
	filter := bson.M{
		"issuer_code":   issuerCode,
		"year":          year,
		"period_string": periodString,
	}
	update := bson.M{
		"$set": bson.M{
			"is_latest":  isLatest,
			"updated_at": time.Now(),
		},
	}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *FinancialReportRepository) ListByIssuer(ctx context.Context, issuerCode string, limit int) ([]*finreport.FinancialReport, error) {
	filter := bson.M{}
	if issuerCode != "" {
		filter["issuer_code"] = issuerCode
	}
	opts := options.Find().SetSort(bson.D{{Key: "year", Value: -1}, {Key: "period_string", Value: -1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}
	cursor, err := r.collection.Find(ctx, filter, opts)
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

func (r *FinancialReportRepository) FindAllNotLatest(ctx context.Context) ([]*finreport.FinancialReport, error) {
	filter := bson.M{"is_latest": false}
	opts := options.Find().SetSort(bson.D{{Key: "downloaded_at", Value: 1}})
	cursor, err := r.collection.Find(ctx, filter, opts)
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

func (r *FinancialReportRepository) MarkDownloaded(ctx context.Context, id string, reportURL string) error {
	filter := parseIDFilter(id)
	update := bson.M{
		"$set": bson.M{
			"is_latest":     true,
			"report_url":    reportURL,
			"downloaded_at": time.Now(),
			"updated_at":    time.Now(),
		},
	}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *FinancialReportRepository) MarkNeedsDownload(ctx context.Context, id string, announcementID string) error {
	filter := parseIDFilter(id)
	update := bson.M{
		"$set": bson.M{
			"is_latest":       false,
			"announcement_id": announcementID,
			"updated_at":      time.Now(),
		},
	}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func parseIDFilter(id string) bson.M {
	if objID, err := bson.ObjectIDFromHex(id); err == nil {
		return bson.M{"_id": objID}
	}
	return bson.M{"_id": id}
}
