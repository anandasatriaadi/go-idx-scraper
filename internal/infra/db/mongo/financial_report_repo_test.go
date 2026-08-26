package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/finreport"
)

type MockFinancialReportRepo struct {
	Reports map[string]*finreport.FinancialReport
	Err     error
}

func NewMockFinancialReportRepo() *MockFinancialReportRepo {
	return &MockFinancialReportRepo{
		Reports: make(map[string]*finreport.FinancialReport),
	}
}

func (m *MockFinancialReportRepo) makeKey(issuer string, year int, period string) string {
	return issuer + "-" + string(rune(year)) + "-" + period
}

func (m *MockFinancialReportRepo) Create(ctx context.Context, model *finreport.FinancialReport) error {
	if m.Err != nil {
		return m.Err
	}
	key := m.makeKey(model.IssuerCode, model.Year, model.PeriodString)
	m.Reports[key] = model
	return nil
}

func (m *MockFinancialReportRepo) FindByIssuerAndPeriod(ctx context.Context, issuerCode string, year int, periodString string) (*finreport.FinancialReport, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	key := m.makeKey(issuerCode, year, periodString)
	return m.Reports[key], nil
}

func (m *MockFinancialReportRepo) UpdateIsLatest(ctx context.Context, issuerCode string, year int, periodString string, isLatest bool) error {
	if m.Err != nil {
		return m.Err
	}
	key := m.makeKey(issuerCode, year, periodString)
	if r, ok := m.Reports[key]; ok {
		r.IsLatest = isLatest
	}
	return nil
}

func (m *MockFinancialReportRepo) ListByIssuer(ctx context.Context, issuerCode string, limit int) ([]*finreport.FinancialReport, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	var list []*finreport.FinancialReport
	for _, r := range m.Reports {
		if issuerCode == "" || r.IssuerCode == issuerCode {
			list = append(list, r)
		}
		if limit > 0 && len(list) >= limit {
			break
		}
	}
	return list, nil
}

func (m *MockFinancialReportRepo) FindAllNotLatest(ctx context.Context) ([]*finreport.FinancialReport, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	var list []*finreport.FinancialReport
	for _, r := range m.Reports {
		if !r.IsLatest {
			list = append(list, r)
		}
	}
	return list, nil
}

func (m *MockFinancialReportRepo) MarkDownloaded(ctx context.Context, id string, reportURL string) error {
	if m.Err != nil {
		return m.Err
	}
	for _, r := range m.Reports {
		if r.ID == id {
			r.IsLatest = true
			r.ReportURL = reportURL
			r.DownloadedAt = time.Now()
		}
	}
	return nil
}

func (m *MockFinancialReportRepo) MarkNeedsDownload(ctx context.Context, id string, announcementID string) error {
	if m.Err != nil {
		return m.Err
	}
	for _, r := range m.Reports {
		if r.ID == id {
			r.IsLatest = false
			r.AnnouncementID = announcementID
		}
	}
	return nil
}

func TestFinancialReportRepository_InterfaceCompliance(t *testing.T) {
	var _ finreport.Repository = (*FinancialReportRepository)(nil)
	var _ finreport.Repository = (*MockFinancialReportRepo)(nil)
}

func TestFinancialReport_MockRepo(t *testing.T) {
	repo := NewMockFinancialReportRepo()
	ctx := context.Background()

	now := time.Now()
	report := &finreport.FinancialReport{
		ID:           "fr-1",
		IssuerCode:   "BBRI",
		Year:         2024,
		PeriodString: "Tahunan",
		DownloadedAt: now,
		IsLatest:     true,
	}

	if err := repo.Create(ctx, report); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := repo.FindByIssuerAndPeriod(ctx, "BBRI", 2024, "Tahunan")
	if err != nil || found == nil {
		t.Fatalf("FindByIssuerAndPeriod failed: %v", err)
	}

	if err := repo.UpdateIsLatest(ctx, "BBRI", 2024, "Tahunan", false); err != nil {
		t.Fatalf("UpdateIsLatest failed: %v", err)
	}
	if found.IsLatest != false {
		t.Errorf("Expected IsLatest=false, got %v", found.IsLatest)
	}
}
