package finreport

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type FinancialReport struct {
	ID             bson.ObjectID `bson:"_id,omitempty" json:"id"`
	CreatedAt      time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time     `bson:"updated_at" json:"updated_at"`
	IssuerCode     string        `bson:"issuer_code" json:"issuer_code"`
	ReportURL      string        `bson:"report_url" json:"report_url"`
	Year           int           `bson:"year" json:"year"`
	PeriodString   string        `bson:"period_string" json:"period_string"`
	AnnouncementID string        `bson:"announcement_id" json:"announcement_id"`
	DownloadedAt   time.Time     `bson:"downloaded_at" json:"downloaded_at"`
	IsLatest       bool          `bson:"is_latest" json:"is_latest"`
}

type Repository interface {
	Create(ctx context.Context, report *FinancialReport) error
	FindAll(ctx context.Context, filter any, opts ...options.Lister[options.FindOptions]) ([]*FinancialReport, error)
	FindOne(ctx context.Context, filter any) (*FinancialReport, error)
	UpdateOne(ctx context.Context, filter, update any) error
}
