package announcement

import (
	"context"
	"testing"

	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/finreport"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
)

type MockRepository struct {
	Announcements []*Announcement
	One           *Announcement
	ExistsRes     bool
	Err           error
}

func (m *MockRepository) Create(ctx context.Context, a *Announcement) error { return m.Err }
func (m *MockRepository) FindAll(ctx context.Context, filter any, opts ...options.Lister[options.FindOptions]) ([]*Announcement, error) {
	return m.Announcements, m.Err
}
func (m *MockRepository) FindByID(ctx context.Context, id string) (*Announcement, error) {
	return m.One, m.Err
}
func (m *MockRepository) Exists(ctx context.Context, id string) (bool, error) {
	return m.ExistsRes, m.Err
}

type MockFinreportRepository struct{}

func (m *MockFinreportRepository) Create(ctx context.Context, r *finreport.FinancialReport) error {
	return nil
}
func (m *MockFinreportRepository) FindByIssuerAndPeriod(ctx context.Context, issuerCode string, year int, periodString string) (*finreport.FinancialReport, error) {
	return nil, nil
}
func (m *MockFinreportRepository) UpdateIsLatest(ctx context.Context, issuerCode string, year int, periodString string, isLatest bool) error {
	return nil
}
func (m *MockFinreportRepository) ListByIssuer(ctx context.Context, issuerCode string, limit int) ([]*finreport.FinancialReport, error) {
	return nil, nil
}
func (m *MockFinreportRepository) FindAllNotLatest(ctx context.Context) ([]*finreport.FinancialReport, error) {
	return nil, nil
}
func (m *MockFinreportRepository) MarkDownloaded(ctx context.Context, id string, reportURL string) error {
	return nil
}
func (m *MockFinreportRepository) MarkNeedsDownload(ctx context.Context, id string, announcementID string) error {
	return nil
}

func TestService_Create(t *testing.T) {
	mock := &MockRepository{}
	mockFinreport := &MockFinreportRepository{}
	svc := NewService(mock, mockFinreport, zap.NewNop())
	err := svc.Create(context.Background(), &Announcement{ID: "1"})
	if err != nil {
		t.Errorf("Create failed: %v", err)
	}
}

func TestService_FindAll(t *testing.T) {
	mock := &MockRepository{Announcements: []*Announcement{{ID: "1"}}}
	mockFinreport := &MockFinreportRepository{}
	svc := NewService(mock, mockFinreport, zap.NewNop())
	res, err := svc.FindAll(context.Background(), nil)
	if err != nil {
		t.Errorf("FindAll failed: %v", err)
	}
	if len(res) != 1 {
		t.Errorf("Expected 1 announcement, got %d", len(res))
	}
}
