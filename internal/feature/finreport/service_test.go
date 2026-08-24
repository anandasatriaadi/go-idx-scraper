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
func (m *MockRepository) FindOne(ctx context.Context, filter any) (*FinancialReport, error) {
	return nil, m.Err
}
func (m *MockRepository) UpdateOne(ctx context.Context, filter, update any) error {
	return m.Err
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

func TestConstructXBRLReportURL(t *testing.T) {
	svc := NewService(nil, zap.NewNop())

	url1 := svc.ConstructXBRLReportURL(2026, "I", "AADI", "")
	expected1 := "https://www.idx.co.id/Portals/0/StaticData/ListedCompanies/Corporate_Actions/New_Info_JSX/Jenis_Informasi/01_Laporan_Keuangan/02_Soft_Copy_Laporan_Keuangan//Laporan%20Keuangan%20Tahun%202026/TW1/AADI/instance.zip"
	if url1 != expected1 {
		t.Errorf("Expected url:\n%s\nGot:\n%s", expected1, url1)
	}

	url2 := svc.ConstructXBRLReportURL(2025, "Tahunan", "BBRI", "inlineXBRL.zip")
	expected2 := "https://www.idx.co.id/Portals/0/StaticData/ListedCompanies/Corporate_Actions/New_Info_JSX/Jenis_Informasi/01_Laporan_Keuangan/02_Soft_Copy_Laporan_Keuangan//Laporan%20Keuangan%20Tahun%202025/Audit/BBRI/inlineXBRL.zip"
	if url2 != expected2 {
		t.Errorf("Expected url:\n%s\nGot:\n%s", expected2, url2)
	}
}
