package finreport

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
)

type Service struct {
	repo   Repository
	logger *zap.Logger
}

func NewService(repo Repository, logger *zap.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

func (s *Service) Create(ctx context.Context, f *FinancialReport) error {
	return s.repo.Create(ctx, f)
}

func (s *Service) FindAll(ctx context.Context, filter bson.M, opts ...options.Lister[options.FindOptions]) ([]*FinancialReport, error) {
	return s.repo.FindAll(ctx, filter, opts...)
}

func (s *Service) FindOne(ctx context.Context, filter any) (*FinancialReport, error) {
	return s.repo.FindOne(ctx, filter)
}

func (s *Service) UpdateOne(ctx context.Context, filter, update any) error {
	return s.repo.UpdateOne(ctx, filter, update)
}

var filenameRegex = regexp.MustCompile(`^FinancialStatement-(\d{4})-(I{1,3}|IV|Tahunan)-([A-Z]+)\.xlsx$`)

func ParseFinancialStatementFilename(filename string) (year int, periodString, issuerCode string, err error) {
	matches := filenameRegex.FindStringSubmatch(filename)
	if len(matches) != 4 {
		return 0, "", "", fmt.Errorf("invalid filename format: %s", filename)
	}
	year, err = strconv.Atoi(matches[1])
	if err != nil {
		return 0, "", "", fmt.Errorf("invalid year in filename: %s", filename)
	}
	return year, matches[2], matches[3], nil
}

func (s *Service) ConstructReportURL(year int, periodString, issuerCode string) string {
	var modePeriod string
	if periodString == "Tahunan" {
		modePeriod = "Audit"
	} else {
		modePeriod = "TW" + romanToNumeral(periodString)
	}
	filename := fmt.Sprintf("FinancialStatement-%d-%s-%s.xlsx", year, periodString, issuerCode)
	url := fmt.Sprintf("https://www.idx.co.id/Portals/0/StaticData/ListedCompanies/Corporate_Actions/New_Info_JSX/Jenis_Informasi/01_Laporan_Keuangan/02_Soft_Copy_Laporan_Keuangan//Laporan%%20Keuangan%%20Tahun%%20%d/%s/%s/%s",
		year, modePeriod, issuerCode, filename)
	return url
}

func romanToNumeral(roman string) string {
	switch roman {
	case "I":
		return "1"
	case "II":
		return "2"
	case "III":
		return "3"
	case "IV":
		return "4"
	default:
		return roman
	}
}

func (s *Service) FindByIssuerYearPeriod(ctx context.Context, issuerCode string, year int, periodString string) (*FinancialReport, error) {
	filter := bson.M{
		"issuer_code":   issuerCode,
		"year":          year,
		"period_string": periodString,
	}
	return s.repo.FindOne(ctx, filter)
}

func (s *Service) MarkAsNeedsDownload(ctx context.Context, id, announcementID string) error {
	update := bson.M{
		"$set": bson.M{
			"is_latest":       false,
			"announcement_id": announcementID,
			"updated_at":      time.Now(),
		},
	}
	return s.repo.UpdateOne(ctx, bson.M{"_id": id}, update)
}

func (s *Service) MarkAsDownloaded(ctx context.Context, id bson.ObjectID, reportURL string) error {
	update := bson.M{
		"$set": bson.M{
			"is_latest":     true,
			"report_url":    reportURL,
			"downloaded_at": time.Now().Unix(),
			"updated_at":    time.Now(),
		},
	}
	return s.repo.UpdateOne(ctx, bson.M{"_id": id}, update)
}

func (s *Service) FindAllNotLatest(ctx context.Context) ([]*FinancialReport, error) {
	filter := bson.M{"is_latest": false}
	opts := options.Find().SetSort(bson.D{{Key: "downloaded_at", Value: 1}})
	return s.repo.FindAll(ctx, filter, opts)
}

type DownloadedReport struct {
	Report    *FinancialReport
	IsUpdated bool
}

func (s *Service) GetReportsForEmail(ctx context.Context) (newReports, updatedReports []*FinancialReport, err error) {
	filter := bson.M{"is_latest": true}
	reports, err := s.repo.FindAll(ctx, filter)
	if err != nil {
		return nil, nil, err
	}

	for _, r := range reports {
		if r.AnnouncementID == "" {
			newReports = append(newReports, r)
		} else {
			updatedReports = append(updatedReports, r)
		}
	}

	sort.Slice(newReports, func(i, j int) bool {
		return newReports[i].IssuerCode < newReports[j].IssuerCode
	})
	sort.Slice(updatedReports, func(i, j int) bool {
		return updatedReports[i].IssuerCode < updatedReports[j].IssuerCode
	})

	return newReports, updatedReports, nil
}
