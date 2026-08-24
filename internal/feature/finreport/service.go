package finreport

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
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

func NormalizePeriod(period string) (periodString string, modePeriod string) {
	p := strings.TrimSpace(strings.ToUpper(period))
	switch p {
	case "I", "1", "TW1", "Q1":
		return "I", "TW1"
	case "II", "2", "TW2", "Q2":
		return "II", "TW2"
	case "III", "3", "TW3", "Q3":
		return "III", "TW3"
	case "IV", "4", "TW4", "Q4":
		return "IV", "TW4"
	case "TAHUNAN", "AUDIT", "FY":
		return "Tahunan", "Audit"
	default:
		return period, "TW" + romanToNumeral(period)
	}
}

func (s *Service) ConstructReportURL(year int, periodString, issuerCode string) string {
	ps, modePeriod := NormalizePeriod(periodString)
	filename := fmt.Sprintf("FinancialStatement-%d-%s-%s.xlsx", year, ps, issuerCode)
	url := fmt.Sprintf("https://www.idx.co.id/Portals/0/StaticData/ListedCompanies/Corporate_Actions/New_Info_JSX/Jenis_Informasi/01_Laporan_Keuangan/02_Soft_Copy_Laporan_Keuangan//Laporan%%20Keuangan%%20Tahun%%20%d/%s/%s/%s",
		year, modePeriod, issuerCode, filename)
	return url
}

func (s *Service) ConstructXBRLReportURL(year int, periodString, issuerCode string, fileType string) string {
	_, modePeriod := NormalizePeriod(periodString)
	if fileType == "" {
		fileType = "instance.zip"
	}
	url := fmt.Sprintf("https://www.idx.co.id/Portals/0/StaticData/ListedCompanies/Corporate_Actions/New_Info_JSX/Jenis_Informasi/01_Laporan_Keuangan/02_Soft_Copy_Laporan_Keuangan//Laporan%%20Keuangan%%20Tahun%%20%d/%s/%s/%s",
		year, modePeriod, issuerCode, fileType)
	return url
}

// IsPeriodReleasedOnIDX checks if the reporting period's filing window has opened on IDX based on the current date
func IsPeriodReleasedOnIDX(year int, period string, now time.Time) bool {
	currentYear := now.Year()
	currentMonth := now.Month()

	// Past fiscal years are always eligible
	if year < currentYear {
		return true
	}

	// Future fiscal years are never eligible
	if year > currentYear {
		return false
	}

	// For the CURRENT fiscal year:
	ps, modePeriod := NormalizePeriod(period)
	pUpper := strings.ToUpper(ps)
	mUpper := strings.ToUpper(modePeriod)

	// FY / Audit / Tahunan for current year Y is only released in year Y+1 (March/April onwards)
	if pUpper == "TAHUNAN" || pUpper == "AUDIT" || pUpper == "FY" || mUpper == "AUDIT" || mUpper == "TAHUNAN" {
		return false
	}

	// Q1 / TW1 (period ends March 31): Filing window opens in April (Month 4)
	if pUpper == "I" || mUpper == "TW1" || pUpper == "Q1" {
		return currentMonth >= time.April
	}

	// Q2 / TW2 (period ends June 30): Filing window opens in July (Month 7)
	if pUpper == "II" || mUpper == "TW2" || pUpper == "Q2" {
		return currentMonth >= time.July
	}

	// Q3 / TW3 (period ends September 30): Filing window opens in October (Month 10)
	if pUpper == "III" || mUpper == "TW3" || pUpper == "Q3" {
		return currentMonth >= time.October
	}

	// Q4 / TW4
	if pUpper == "IV" || mUpper == "TW4" || pUpper == "Q4" {
		return false
	}

	return true
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
