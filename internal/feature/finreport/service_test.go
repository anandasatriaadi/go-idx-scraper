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

	url3 := svc.ConstructXBRLReportURL(2024, "TW2", "TLKM", "instance.zip")
	expected3 := "https://www.idx.co.id/Portals/0/StaticData/ListedCompanies/Corporate_Actions/New_Info_JSX/Jenis_Informasi/01_Laporan_Keuangan/02_Soft_Copy_Laporan_Keuangan//Laporan%20Keuangan%20Tahun%202024/TW2/TLKM/instance.zip"
	if url3 != expected3 {
		t.Errorf("Expected url:\n%s\nGot:\n%s", expected3, url3)
	}

	url4 := svc.ConstructXBRLReportURL(2023, "Audit", "BBCA", "")
	expected4 := "https://www.idx.co.id/Portals/0/StaticData/ListedCompanies/Corporate_Actions/New_Info_JSX/Jenis_Informasi/01_Laporan_Keuangan/02_Soft_Copy_Laporan_Keuangan//Laporan%20Keuangan%20Tahun%202023/Audit/BBCA/instance.zip"
	if url4 != expected4 {
		t.Errorf("Expected url:\n%s\nGot:\n%s", expected4, url4)
	}
}

func TestNormalizePeriod(t *testing.T) {
	tests := []struct {
		input              string
		expectedPeriodStr  string
		expectedModePeriod string
	}{
		{"1", "I", "TW1"},
		{"I", "I", "TW1"},
		{"TW1", "I", "TW1"},
		{"Q1", "I", "TW1"},
		{"2", "II", "TW2"},
		{"II", "II", "TW2"},
		{"TW2", "II", "TW2"},
		{"Q2", "II", "TW2"},
		{"3", "III", "TW3"},
		{"III", "III", "TW3"},
		{"TW3", "III", "TW3"},
		{"Q3", "III", "TW3"},
		{"4", "IV", "TW4"},
		{"IV", "IV", "TW4"},
		{"TW4", "IV", "TW4"},
		{"Q4", "IV", "TW4"},
		{"Audit", "Tahunan", "Audit"},
		{"Tahunan", "Tahunan", "Audit"},
		{"FY", "Tahunan", "Audit"},
		{"tahunan", "Tahunan", "Audit"},
		{"audit", "Tahunan", "Audit"},
	}

	for _, tt := range tests {
		ps, mp := NormalizePeriod(tt.input)
		if ps != tt.expectedPeriodStr || mp != tt.expectedModePeriod {
			t.Errorf("NormalizePeriod(%q) = (%q, %q), expected (%q, %q)",
				tt.input, ps, mp, tt.expectedPeriodStr, tt.expectedModePeriod)
		}
	}
}
