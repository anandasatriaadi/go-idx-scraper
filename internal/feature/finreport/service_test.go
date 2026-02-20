package finreport

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
)

type MockRepository struct {
	Reports []*FinancialReport
	Err     error
}

func (m *MockRepository) Create(ctx context.Context, r *FinancialReport) error { return m.Err }
func (m *MockRepository) FindAll(ctx context.Context, filter any, opts ...options.Lister[options.FindOptions]) ([]*FinancialReport, error) {
	return m.Reports, m.Err
}

func TestService_Create(t *testing.T) {
	mock := &MockRepository{}
	svc := NewService(mock, zap.NewNop())
	err := svc.Create(context.Background(), &FinancialReport{IssuerCode: "AAPL"})
	if err != nil {
		t.Errorf("Create failed: %v", err)
	}
}

func TestService_FindAll(t *testing.T) {
	mock := &MockRepository{Reports: []*FinancialReport{{IssuerCode: "AAPL"}}}
	svc := NewService(mock, zap.NewNop())
	res, err := svc.FindAll(context.Background(), nil)
	if err != nil {
		t.Errorf("FindAll failed: %v", err)
	}
	if len(res) != 1 {
		t.Errorf("Expected 1 report, got %d", len(res))
	}
}
