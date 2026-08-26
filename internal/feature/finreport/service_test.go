package finreport_test

import (
	"context"
	"testing"
	"time"

	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/finreport"
	"go.uber.org/zap"
)

type mockFinReportRepo struct {
	reports map[string]*finreport.FinancialReport
}

func newMockFinReportRepo() *mockFinReportRepo {
	return &mockFinReportRepo{
		reports: make(map[string]*finreport.FinancialReport),
	}
}

func (m *mockFinReportRepo) Create(ctx context.Context, report *finreport.FinancialReport) error {
	key := report.IssuerCode + "-" + report.PeriodString
	m.reports[key] = report
	return nil
}

func (m *mockFinReportRepo) FindByIssuerAndPeriod(ctx context.Context, issuerCode string, year int, periodString string) (*finreport.FinancialReport, error) {
	key := issuerCode + "-" + periodString
	return m.reports[key], nil
}

func (m *mockFinReportRepo) UpdateIsLatest(ctx context.Context, issuerCode string, year int, periodString string, isLatest bool) error {
	key := issuerCode + "-" + periodString
	if r, ok := m.reports[key]; ok {
		r.IsLatest = isLatest
	}
	return nil
}

func (m *mockFinReportRepo) ListByIssuer(ctx context.Context, issuerCode string, limit int) ([]*finreport.FinancialReport, error) {
	var list []*finreport.FinancialReport
	for _, r := range m.reports {
		if r.IssuerCode == issuerCode {
			list = append(list, r)
		}
	}
	return list, nil
}

func (m *mockFinReportRepo) FindAllNotLatest(ctx context.Context) ([]*finreport.FinancialReport, error) {
	var list []*finreport.FinancialReport
	for _, r := range m.reports {
		if !r.IsLatest {
			list = append(list, r)
		}
	}
	return list, nil
}

func (m *mockFinReportRepo) MarkDownloaded(ctx context.Context, id string, reportURL string) error {
	for _, r := range m.reports {
		if r.ID == id {
			r.IsLatest = true
			r.ReportURL = reportURL
			r.DownloadedAt = time.Now()
		}
	}
	return nil
}

func (m *mockFinReportRepo) MarkNeedsDownload(ctx context.Context, id string, announcementID string) error {
	for _, r := range m.reports {
		if r.ID == id {
			r.IsLatest = false
			r.AnnouncementID = announcementID
		}
	}
	return nil
}

func TestService_Create(t *testing.T) {
	mock := newMockFinReportRepo()
	svc := finreport.NewService(mock, zap.NewNop())
	err := svc.Create(context.Background(), &finreport.FinancialReport{
		IssuerCode:   "AAPL",
		PeriodString: "Tahunan",
		Year:         2024,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := svc.FindByIssuerYearPeriod(context.Background(), "AAPL", 2024, "Tahunan")
	if err != nil {
		t.Fatalf("FindByIssuerYearPeriod failed: %v", err)
	}
	if found == nil || found.IssuerCode != "AAPL" {
		t.Fatalf("Expected report for AAPL, got %+v", found)
	}
}

func TestParseFinancialStatementFilename(t *testing.T) {
	filename := "FinancialStatement-2024-Tahunan-BBRI.xlsx"
	year, period, issuer, err := finreport.ParseFinancialStatementFilename(filename)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if year != 2024 || period != "Tahunan" || issuer != "BBRI" {
		t.Fatalf("unexpected parse result: year=%d, period=%s, issuer=%s", year, period, issuer)
	}
}

func TestConstructXBRLReportURL(t *testing.T) {
	svc := finreport.NewService(nil, zap.NewNop())

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
		ps, mp := finreport.NormalizePeriod(tt.input)
		if ps != tt.expectedPeriodStr || mp != tt.expectedModePeriod {
			t.Errorf("NormalizePeriod(%q) = (%q, %q), expected (%q, %q)",
				tt.input, ps, mp, tt.expectedPeriodStr, tt.expectedModePeriod)
		}
	}
}

func TestIsPeriodReleasedOnIDX(t *testing.T) {
	august2026 := time.Date(2026, 8, 25, 10, 0, 0, 0, time.UTC)

	// Past years: all periods released
	if !finreport.IsPeriodReleasedOnIDX(2025, "Audit", august2026) {
		t.Errorf("Expected 2025 Audit to be released in 2026")
	}
	if !finreport.IsPeriodReleasedOnIDX(2025, "TW3", august2026) {
		t.Errorf("Expected 2025 TW3 to be released in 2026")
	}

	// Current year 2026 in August:
	// Q1 (TW1) -> true
	if !finreport.IsPeriodReleasedOnIDX(2026, "TW1", august2026) {
		t.Errorf("Expected 2026 TW1 to be released in August")
	}
	// Q2 (TW2) -> true
	if !finreport.IsPeriodReleasedOnIDX(2026, "TW2", august2026) {
		t.Errorf("Expected 2026 TW2 to be released in August")
	}
	// Q3 (TW3 - ends Sept 30) -> false in August
	if finreport.IsPeriodReleasedOnIDX(2026, "TW3", august2026) {
		t.Errorf("Expected 2026 TW3 to NOT be released in August 2026")
	}
	// FY / Audit -> false in August of same year
	if finreport.IsPeriodReleasedOnIDX(2026, "Tahunan", august2026) {
		t.Errorf("Expected 2026 Tahunan to NOT be released in August 2026")
	}
	if finreport.IsPeriodReleasedOnIDX(2026, "Audit", august2026) {
		t.Errorf("Expected 2026 Audit to NOT be released in August 2026")
	}
}
