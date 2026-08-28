package parser

import (
	"strings"

	domain "github.com/anandasatriaadi/go-idx-scraper/internal/feature/xbrl"
)

// AssignDEIMetadata maps Document and Entity Information (/dei) taxonomy tags to the Statement entity
func AssignDEIMetadata(s *domain.Statement, tag, val string) {
	switch tag {
	case "EntityCode":
		s.Ticker = val
	case "EntityName":
		s.CompanyName = val
	case "Sector":
		s.Metadata.Sector = val
	case "Subsector":
		s.Metadata.Subsector = val
	case "Industry":
		s.Metadata.Industry = val
	case "Subindustry":
		s.Metadata.Subindustry = val
	case "DescriptionOfPresentationCurrency":
		if strings.Contains(val, "USD") || strings.Contains(val, "Dollar") {
			s.Metadata.Currency = "USD"
		} else {
			s.Metadata.Currency = "IDR"
		}
	case "LevelOfRoundingUsedInFinancialStatements":
		s.Metadata.RoundingLevel = val
		if strings.Contains(val, "Thousand") || strings.Contains(val, "Ribuan") {
			s.Metadata.RoundingMultiplier = 1000
		} else if strings.Contains(val, "Million") || strings.Contains(val, "Jutaan") {
			s.Metadata.RoundingMultiplier = 1000000
		} else if strings.Contains(val, "Billion") || strings.Contains(val, "Miliaran") {
			s.Metadata.RoundingMultiplier = 1000000000
		}
	case "ConversionRateAtReportingDateIfPresentationCurrencyIsOtherThanRupiah":
		rate, _ := ParseNumericValue(val)
		if rate > 0 && rate < 1000 {
			// Handles Indonesian notation where 16.680 was written with dot as thousand separator
			rate = rate * 1000
		}
		s.Metadata.ConversionRate = rate
	case "CurrentPeriodEndDate", "PeriodEndDate", "BalanceSheetDate", "CurrentPeriodStartDate":
		if t, err := ParseFlexibleDate(val); err == nil {
			if s.Year == 0 || tag == "CurrentPeriodEndDate" {
				s.PeriodEndDate = t
				s.Year = t.Year()
			}
		}
	case "PeriodOfFinancialStatementsSubmissions":
		s.PeriodType = val
		if strings.Contains(val, "Kuartal III") || strings.Contains(val, "Third") {
			s.Period = "Q3"
		} else if strings.Contains(val, "Kuartal II") || strings.Contains(val, "Second") {
			s.Period = "Q2"
		} else if strings.Contains(val, "Kuartal IV") || strings.Contains(val, "Fourth") {
			s.Period = "FY"
		} else if strings.Contains(val, "Kuartal I") || strings.Contains(val, "First") {
			s.Period = "Q1"
		} else {
			s.Period = "FY"
		}
	case "TypeOfReportOnFinancialStatements":
		s.Metadata.AuditStatus = val
	case "TypeOfAuditorsOpinion":
		s.Metadata.AuditorOpinion = val
	}
}
