package announcement

import (
	"context"
	"strings"
	"time"

	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/common"
	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/finreport"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
)

const (
	reportPrefix      = "Penyampaian Laporan Keuangan"
	attachmentPattern = "FinancialStatement-"
	twentyFourHours   = 24 * time.Hour
)

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

func (s *Service) FindAll(ctx context.Context, filter any, opts ...options.Lister[options.FindOptions]) ([]*Announcement, error) {
	return s.repo.FindAll(ctx, filter, opts...)
}

func (s *Service) FindByID(ctx context.Context, id string) (*Announcement, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) Exists(ctx context.Context, id string) (bool, error) {
	return s.repo.Exists(ctx, id)
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

	existing, err := s.finreport.FindOne(ctx, bson.M{
		"issuer_code":   issuerCode,
		"year":          year,
		"period_string": periodString,
	})
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

	return s.finreport.UpdateOne(ctx,
		bson.M{"_id": existing.ID},
		bson.M{"$set": bson.M{
			"is_latest":       false,
			"announcement_id": a.ID,
			"updated_at":      now,
		}},
	)
}
