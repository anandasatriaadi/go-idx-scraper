package announcement

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/common"
	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/finreport"
	"go.uber.org/zap"
)

const (
	reportPrefix      = "Penyampaian Laporan Keuangan"
	attachmentPattern = "FinancialStatement-"
	twentyFourHours   = 24 * time.Hour
)

var nonAlphaRegex = regexp.MustCompile(`[^a-zA-Z]`)

var excludedEmailTitles = []string{
	"laporanbulananregistrasipemegangefek",
	"penjelasanatasvolatilitastransaksi",
	"penyampaianbuktiiklan",
}

type Service struct {
	repo      Repository
	finreport finreport.Repository
	logger    *zap.Logger
}

func NewService(repo Repository, finreportRepo finreport.Repository, logger *zap.Logger) *Service {
	return &Service{
		repo:      repo,
		finreport: finreportRepo,
		logger:    logger,
	}
}

func (s *Service) Create(ctx context.Context, a *Announcement) error {
	return s.repo.Create(ctx, a)
}

func (s *Service) CreateMany(ctx context.Context, announcements []*Announcement) error {
	for _, a := range announcements {
		if err := s.repo.Create(ctx, a); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) FindByID(ctx context.Context, id string) (*Announcement, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) Exists(ctx context.Context, id string) (bool, error) {
	return s.repo.Exists(ctx, id)
}

func (s *Service) FindRecent(ctx context.Context, limit int) ([]*Announcement, error) {
	return s.repo.FindRecent(ctx, limit)
}

func (s *Service) GetLatestCreatedDate(ctx context.Context) (*time.Time, error) {
	return s.repo.GetLatestCreatedDate(ctx)
}

func (s *Service) FindExistingIDs(ctx context.Context, limit int) (map[string]bool, error) {
	return s.repo.FindExistingIDs(ctx, limit)
}

func (s *Service) FilterDisclosuresForEmail(announcements []*Announcement) []*Announcement {
	var filtered []*Announcement
	for _, ann := range announcements {
		if ann.JudulPengumuman != nil {
			normalizedTitle := nonAlphaRegex.ReplaceAllString(*ann.JudulPengumuman, "")
			normalizedTitle = strings.TrimSpace(normalizedTitle)
			normalizedTitle = strings.ToLower(normalizedTitle)

			excluded := false
			for _, pattern := range excludedEmailTitles {
				if strings.HasPrefix(normalizedTitle, pattern) {
					excluded = true
					s.logger.Info("Announcement excluded from email", zap.String("title", *ann.JudulPengumuman))
					break
				}
			}
			if !excluded {
				filtered = append(filtered, ann)
			}
		} else {
			filtered = append(filtered, ann)
		}
	}
	return filtered
}

func (s *Service) SyncDisclosures(ctx context.Context, dateFrom, dateTo string, provider IDXDataProvider, latestDate *time.Time) ([]*Announcement, error) {
	announcements, err := provider.Fetch(ctx, dateFrom, dateTo)
	if err != nil {
		s.logger.Error("Failed to fetch announcements from IDX", zap.Error(err))
		return nil, err
	}
	s.logger.Info("Announcements fetched successfully", zap.Int("count", len(announcements)))

	exists, err := s.repo.FindExistingIDs(ctx, 500)
	if err != nil {
		s.logger.Error("Failed to check existing announcements", zap.Error(err))
		return nil, err
	}

	var filtered []*Announcement
	for _, ann := range announcements {
		if ann.CreatedDate == nil {
			continue
		}
		if (latestDate == nil || ann.CreatedDate.After(*latestDate)) && !exists[ann.ID] {
			s.logger.Info("New announcement found", zap.String("ID", ann.ID), zap.Time("CreatedDate", *ann.CreatedDate))
			filtered = append(filtered, ann)
		}
	}
	s.logger.Info("Announcements filtered", zap.Int("new", len(filtered)))

	for _, f := range filtered {
		if err := s.repo.Create(ctx, f); err != nil {
			s.logger.Error("Failed to create announcement", zap.String("ID", f.ID), zap.Error(err))
		}
		if err := s.ProcessFinancialReportAnnouncement(ctx, f); err != nil {
			s.logger.Error("Failed to process finreport announcement", zap.String("ID", f.ID), zap.Error(err))
		}
	}

	return filtered, nil
}

func (s *Service) ProcessFinancialReportAnnouncement(ctx context.Context, a *Announcement) error {
	if a.JudulPengumuman == nil || !strings.HasPrefix(*a.JudulPengumuman, reportPrefix) {
		return nil
	}

	if a.KodeEmiten == nil {
		s.logger.Warn("Announcement has no issuer code", zap.String("id", a.ID))
		return nil
	}

	var attachment *common.Attachment
	for i := range a.Attachments {
		if a.Attachments[i].OriginalFilename != nil &&
			strings.HasPrefix(*a.Attachments[i].OriginalFilename, attachmentPattern) {
			attachment = &a.Attachments[i]
			break
		}
	}
	if attachment == nil || attachment.OriginalFilename == nil {
		return nil
	}

	filename := *attachment.OriginalFilename
	year, periodString, issuerCode, err := finreport.ParseFinancialStatementFilename(filename)
	if err != nil {
		s.logger.Warn("Failed to parse filename", zap.String("filename", filename), zap.Error(err))
		return nil
	}

	if issuerCode != *a.KodeEmiten {
		s.logger.Warn("Issuer code mismatch",
			zap.String("filename_issuer", issuerCode),
			zap.String("announcement_issuer", *a.KodeEmiten))
		return nil
	}

	existing, err := s.finreport.FindByIssuerAndPeriod(ctx, issuerCode, year, periodString)
	if err != nil {
		return err
	}

	now := time.Now()
	if existing == nil {
		report := &finreport.FinancialReport{
			IssuerCode:     issuerCode,
			Year:           year,
			PeriodString:   periodString,
			AnnouncementID: a.ID,
			DownloadedAt:   time.UnixMilli(int64(0)),
			IsLatest:       false,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		return s.finreport.Create(ctx, report)
	}

	if existing.AnnouncementID == a.ID {
		return nil
	}

	if existing.AnnouncementID == "" && existing.DownloadedAt.UnixMilli() > 0 {
		if now.Sub(existing.DownloadedAt) < twentyFourHours {
			return nil
		}
	}

	return s.finreport.MarkNeedsDownload(ctx, existing.ID, a.ID)
}
