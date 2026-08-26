package finreport

import (
	"context"
	"time"
)

type FinancialReport struct {
	ID             string    `bson:"_id,omitempty" json:"id,omitempty"`
	IssuerCode     string    `bson:"issuer_code" json:"issuer_code"`
	ReportURL      string    `bson:"report_url" json:"report_url"`
	Year           int       `bson:"year" json:"year"`
	PeriodString   string    `bson:"period_string" json:"period_string"`
	AnnouncementID string    `bson:"announcement_id" json:"announcement_id"`
	DownloadedAt   time.Time `bson:"downloaded_at" json:"downloaded_at"`
	IsLatest       bool      `bson:"is_latest" json:"is_latest"`
	CreatedAt      time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time `bson:"updated_at" json:"updated_at"`
}

type Repository interface {
	Create(ctx context.Context, report *FinancialReport) error
	FindByIssuerAndPeriod(ctx context.Context, issuerCode string, year int, periodString string) (*FinancialReport, error)
	UpdateIsLatest(ctx context.Context, issuerCode string, year int, periodString string, isLatest bool) error
	ListByIssuer(ctx context.Context, issuerCode string, limit int) ([]*FinancialReport, error)
	FindAllNotLatest(ctx context.Context) ([]*FinancialReport, error)
	MarkDownloaded(ctx context.Context, id string, reportURL string) error
	MarkNeedsDownload(ctx context.Context, id string, announcementID string) error
}
