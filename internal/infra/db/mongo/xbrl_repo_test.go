package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/xbrl"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MockXBRLRepo struct {
	Statement *xbrl.Statement
	List      []*xbrl.Statement
	Err       error
}

func (m *MockXBRLRepo) Upsert(ctx context.Context, s *xbrl.Statement) error {
	m.Statement = s
	return m.Err
}

func (m *MockXBRLRepo) FindByTickerAndPeriod(ctx context.Context, ticker string, year int, period string) (*xbrl.Statement, error) {
	return m.Statement, m.Err
}

func (m *MockXBRLRepo) FindHistoricalByTicker(ctx context.Context, ticker string, limit int) ([]*xbrl.Statement, error) {
	return m.List, m.Err
}

func (m *MockXBRLRepo) FindAllWithFilter(ctx context.Context, filter any, opts ...options.Lister[options.FindOptions]) ([]*xbrl.Statement, error) {
	return m.List, m.Err
}

func TestXBRLRepository_InterfaceCompliance(t *testing.T) {
	var _ xbrl.Repository = (*XBRLRepository)(nil)
	var _ xbrl.Repository = (*MockXBRLRepo)(nil)
}

func TestXBRLStatement_MockRepo(t *testing.T) {
	mock := &MockXBRLRepo{}
	stmt := &xbrl.Statement{
		ID:            bson.NewObjectID(),
		Ticker:        "BBRI",
		Year:          2025,
		Period:        "Q3",
		PeriodEndDate: time.Now(),
	}

	err := mock.Upsert(context.Background(), stmt)
	if err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}

	if mock.Statement.Ticker != "BBRI" {
		t.Errorf("Expected Ticker BBRI, got %s", mock.Statement.Ticker)
	}
}
