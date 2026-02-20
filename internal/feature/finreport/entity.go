package finreport

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type FinancialReport struct {
	CreatedAt    time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time `bson:"updated_at" json:"updated_at"`
	IssuerCode   string    `bson:"issuer_code" json:"issuer_code"`
	ReportURL    string    `bson:"report_url" json:"report_url"`
	Year         int       `bson:"year" json:"year"`
	Quarter      int       `bson:"quarter" json:"quarter"`
	DownloadedAt int64     `bson:"downloaded_at" json:"downloaded_at"`
}

type Repository interface {
	Create(ctx context.Context, report *FinancialReport) error
	FindAll(ctx context.Context, filter any, opts ...options.Lister[options.FindOptions]) ([]*FinancialReport, error)
	// Add specific queries as needed
}
