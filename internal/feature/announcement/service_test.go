package announcement_test

import (
	"context"
	"testing"
	"time"

	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/announcement"
	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/common"
	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/finreport"
	"go.uber.org/zap"
)

type mockAnnouncementRepo struct {
	items       map[string]*announcement.Announcement
	latestDate  *time.Time
	existingIDs map[string]bool
	createErr   error
}

func newMockAnnouncementRepo() *mockAnnouncementRepo {
	return &mockAnnouncementRepo{
		items:       make(map[string]*announcement.Announcement),
		existingIDs: make(map[string]bool),
	}
}

func (m *mockAnnouncementRepo) Create(ctx context.Context, a *announcement.Announcement) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.items[a.ID] = a
	m.existingIDs[a.ID] = true
	return nil
}

func (m *mockAnnouncementRepo) FindByID(ctx context.Context, id string) (*announcement.Announcement, error) {
	return m.items[id], nil
}

func (m *mockAnnouncementRepo) Exists(ctx context.Context, id string) (bool, error) {
	return m.existingIDs[id], nil
}

func (m *mockAnnouncementRepo) FindRecent(ctx context.Context, limit int) ([]*announcement.Announcement, error) {
	var list []*announcement.Announcement
	for _, item := range m.items {
		list = append(list, item)
		if limit > 0 && len(list) >= limit {
			break
		}
	}
	return list, nil
}

func (m *mockAnnouncementRepo) GetLatestCreatedDate(ctx context.Context) (*time.Time, error) {
	return m.latestDate, nil
}

func (m *mockAnnouncementRepo) FindExistingIDs(ctx context.Context, limit int) (map[string]bool, error) {
	result := make(map[string]bool)
	for id, exists := range m.existingIDs {
		result[id] = exists
	}
	return result, nil
}

type mockFinreportRepo struct {
	reports        map[string]*finreport.FinancialReport
	markedNeeds    map[string]string
	createReportFn func(r *finreport.FinancialReport) error
}

func newMockFinreportRepo() *mockFinreportRepo {
	return &mockFinreportRepo{
		reports:     make(map[string]*finreport.FinancialReport),
		markedNeeds: make(map[string]string),
	}
}

func (m *mockFinreportRepo) Create(ctx context.Context, r *finreport.FinancialReport) error {
	if m.createReportFn != nil {
		return m.createReportFn(r)
	}
	key := r.IssuerCode + "-" + r.PeriodString
	m.reports[key] = r
	return nil
}

func (m *mockFinreportRepo) FindByIssuerAndPeriod(ctx context.Context, issuerCode string, year int, periodString string) (*finreport.FinancialReport, error) {
	key := issuerCode + "-" + periodString
	return m.reports[key], nil
}

func (m *mockFinreportRepo) UpdateIsLatest(ctx context.Context, issuerCode string, year int, periodString string, isLatest bool) error {
	key := issuerCode + "-" + periodString
	if r, ok := m.reports[key]; ok {
		r.IsLatest = isLatest
	}
	return nil
}

func (m *mockFinreportRepo) ListByIssuer(ctx context.Context, issuerCode string, limit int) ([]*finreport.FinancialReport, error) {
	var list []*finreport.FinancialReport
	for _, r := range m.reports {
		if r.IssuerCode == issuerCode {
			list = append(list, r)
		}
	}
	return list, nil
}

func (m *mockFinreportRepo) FindAllNotLatest(ctx context.Context) ([]*finreport.FinancialReport, error) {
	var list []*finreport.FinancialReport
	for _, r := range m.reports {
		if !r.IsLatest {
			list = append(list, r)
		}
	}
	return list, nil
}

func (m *mockFinreportRepo) MarkDownloaded(ctx context.Context, id string, reportURL string) error {
	return nil
}

func (m *mockFinreportRepo) MarkNeedsDownload(ctx context.Context, id string, announcementID string) error {
	m.markedNeeds[id] = announcementID
	return nil
}

type mockIDXProvider struct {
	announcements []*announcement.Announcement
	err           error
}

func (m *mockIDXProvider) Fetch(ctx context.Context, dateFrom, dateTo string) ([]*announcement.Announcement, error) {
	return m.announcements, m.err
}

func TestService_BasicOperations(t *testing.T) {
	ctx := context.Background()
	repo := newMockAnnouncementRepo()
	finRepo := newMockFinreportRepo()
	svc := announcement.NewService(repo, finRepo, zap.NewNop())

	ann := &announcement.Announcement{
		ID: "ann-1",
	}

	if err := svc.Create(ctx, ann); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := svc.FindByID(ctx, "ann-1")
	if err != nil || found == nil || found.ID != "ann-1" {
		t.Fatalf("FindByID failed, expected ann-1, got %v", found)
	}

	exists, err := svc.Exists(ctx, "ann-1")
	if err != nil || !exists {
		t.Fatalf("Exists expected true, got %v", exists)
	}

	list, err := svc.FindRecent(ctx, 10)
	if err != nil || len(list) != 1 {
		t.Fatalf("FindRecent expected 1 item, got %d", len(list))
	}

	ids, err := svc.FindExistingIDs(ctx, 10)
	if err != nil || !ids["ann-1"] {
		t.Fatalf("FindExistingIDs failed, expected ann-1 in map")
	}
}

func TestService_CreateMany(t *testing.T) {
	ctx := context.Background()
	repo := newMockAnnouncementRepo()
	finRepo := newMockFinreportRepo()
	svc := announcement.NewService(repo, finRepo, zap.NewNop())

	anns := []*announcement.Announcement{
		{ID: "ann-1"},
		{ID: "ann-2"},
	}

	if err := svc.CreateMany(ctx, anns); err != nil {
		t.Fatalf("CreateMany failed: %v", err)
	}

	if len(repo.items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(repo.items))
	}
}

func TestService_FilterDisclosuresForEmail(t *testing.T) {
	repo := newMockAnnouncementRepo()
	finRepo := newMockFinreportRepo()
	svc := announcement.NewService(repo, finRepo, zap.NewNop())

	title1 := "Penyampaian Laporan Keuangan Tahunan"
	title2 := "Laporan Bulanan Registrasi Pemegang Efek"
	title3 := "Penjelasan atas Volatilitas Transaksi"

	anns := []*announcement.Announcement{
		{ID: "1", JudulPengumuman: &title1},
		{ID: "2", JudulPengumuman: &title2},
		{ID: "3", JudulPengumuman: &title3},
		{ID: "4", JudulPengumuman: nil},
	}

	filtered := svc.FilterDisclosuresForEmail(anns)
	if len(filtered) != 2 {
		t.Fatalf("expected 2 filtered announcements, got %d", len(filtered))
	}
	if filtered[0].ID != "1" || filtered[1].ID != "4" {
		t.Fatalf("unexpected filtered IDs: %v, %v", filtered[0].ID, filtered[1].ID)
	}
}

func TestService_SyncDisclosures(t *testing.T) {
	ctx := context.Background()
	repo := newMockAnnouncementRepo()
	finRepo := newMockFinreportRepo()
	svc := announcement.NewService(repo, finRepo, zap.NewNop())

	now := time.Now()
	past := now.Add(-1 * time.Hour)
	future := now.Add(1 * time.Hour)

	title := "Penyampaian Laporan Keuangan Tahunan 2024"
	issuer := "BBRI"
	filename := "FinancialStatement-2024-Tahunan-BBRI.xlsx"

	provider := &mockIDXProvider{
		announcements: []*announcement.Announcement{
			{
				ID:              "ann-new",
				CreatedDate:     &future,
				JudulPengumuman: &title,
				KodeEmiten:      &issuer,
				Attachments: []common.Attachment{
					{OriginalFilename: &filename},
				},
			},
			{
				ID:          "ann-old",
				CreatedDate: &past,
			},
		},
	}

	created, err := svc.SyncDisclosures(ctx, "20240101", "20240102", provider, &now)
	if err != nil {
		t.Fatalf("SyncDisclosures failed: %v", err)
	}

	if len(created) != 1 || created[0].ID != "ann-new" {
		t.Fatalf("expected ann-new created, got %d items", len(created))
	}

	// Verify finreport was created
	rep, err := finRepo.FindByIssuerAndPeriod(ctx, "BBRI", 2024, "Tahunan")
	if err != nil || rep == nil {
		t.Fatalf("expected finreport created for BBRI, got %v", rep)
	}
	if rep.AnnouncementID != "ann-new" {
		t.Fatalf("expected announcement ID ann-new, got %s", rep.AnnouncementID)
	}
}

func TestService_ProcessFinancialReportAnnouncement(t *testing.T) {
	ctx := context.Background()
	repo := newMockAnnouncementRepo()
	finRepo := newMockFinreportRepo()
	svc := announcement.NewService(repo, finRepo, zap.NewNop())

	// 1. Non-matching title
	nonFinTitle := "Keterbukaan Informasi Umum"
	issuer := "BBRI"
	filename := "FinancialStatement-2024-Tahunan-BBRI.xlsx"
	annNonFin := &announcement.Announcement{
		ID:              "ann-1",
		JudulPengumuman: &nonFinTitle,
		KodeEmiten:      &issuer,
		Attachments:     []common.Attachment{{OriginalFilename: &filename}},
	}
	if err := svc.ProcessFinancialReportAnnouncement(ctx, annNonFin); err != nil {
		t.Fatalf("unexpected error for non-fin announcement: %v", err)
	}

	// 2. Matching announcement
	finTitle := "Penyampaian Laporan Keuangan"
	annFin := &announcement.Announcement{
		ID:              "ann-2",
		JudulPengumuman: &finTitle,
		KodeEmiten:      &issuer,
		Attachments:     []common.Attachment{{OriginalFilename: &filename}},
	}
	if err := svc.ProcessFinancialReportAnnouncement(ctx, annFin); err != nil {
		t.Fatalf("unexpected error for fin announcement: %v", err)
	}
	rep, _ := finRepo.FindByIssuerAndPeriod(ctx, "BBRI", 2024, "Tahunan")
	if rep == nil || rep.AnnouncementID != "ann-2" {
		t.Fatalf("expected report created with announcement ID ann-2, got %v", rep)
	}

	// 3. Duplicate announcement ID
	if err := svc.ProcessFinancialReportAnnouncement(ctx, annFin); err != nil {
		t.Fatalf("unexpected error on duplicate: %v", err)
	}

	// 4. Update with new announcement ID for existing report
	annFinUpdate := &announcement.Announcement{
		ID:              "ann-3",
		JudulPengumuman: &finTitle,
		KodeEmiten:      &issuer,
		Attachments:     []common.Attachment{{OriginalFilename: &filename}},
	}
	rep.ID = "report-id-1"
	if err := svc.ProcessFinancialReportAnnouncement(ctx, annFinUpdate); err != nil {
		t.Fatalf("unexpected error on update: %v", err)
	}
	if finRepo.markedNeeds["report-id-1"] != "ann-3" {
		t.Fatalf("expected report-id-1 marked with ann-3, got %s", finRepo.markedNeeds["report-id-1"])
	}
}
