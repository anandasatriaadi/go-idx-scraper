package financialreport

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type FinancialReport struct {
	Year         int    `bson:"year" json:"year"`
	Quarter      int    `bson:"quarter" json:"quarter"`
	IssuerCode   string `bson:"issuer_code" json:"issuer_code"`
	ReportURL    string `bson:"report_url" json:"report_url"`
	DownloadedAt int64  `bson:"downloaded_at" json:"downloaded_at"`
}

type Repository interface {
	Create(ctx context.Context, report *FinancialReport) error
	FindAll(ctx context.Context, filter interface{}, opts ...options.Lister[options.FindOptions]) ([]*FinancialReport, error)
	// Add specific queries as needed
}
