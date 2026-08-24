package xbrl

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	domain "github.com/anandasatriaadi/go-idx-scraper/internal/feature/xbrl"
)

// ParseInstanceZip opens an instance.zip or inlineXBRL.zip file and parses the contents
func ParseInstanceZip(zipPath string) (*domain.Statement, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("opening zip file: %w", err)
	}
	defer r.Close()

	// 1. Try finding standalone .xbrl or .xml instance file
	for _, f := range r.File {
		nameLower := strings.ToLower(f.Name)
		if (strings.HasSuffix(nameLower, ".xbrl") || strings.HasSuffix(nameLower, ".xml")) && !strings.Contains(nameLower, "taxonomy") {
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("reading file in zip: %w", err)
			}
			defer rc.Close()
			stmt, err := ParseInstanceXML(rc)
			if err != nil {
				return nil, err
			}
			stmt.Metadata.SourceFile = filepath.Base(zipPath)
			return stmt, nil
		}
	}

	// 2. Try parsing inline XBRL HTML files inside zip (e.g. 1000000.html, 1210000.html)
	stmt := &domain.Statement{
		Metadata: domain.StatementMetadata{
			RoundingMultiplier: 1.0,
			SourceFile:         filepath.Base(zipPath),
		},
		Facts: make(domain.FactMap),
	}
	hasParsedHTML := false
	for _, f := range r.File {
		nameLower := strings.ToLower(f.Name)
		if strings.HasSuffix(nameLower, ".html") || strings.HasSuffix(nameLower, ".htm") {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			s, err := ParseInstanceXML(rc)
			rc.Close()
			if err == nil {
				hasParsedHTML = true
				mergeStatements(stmt, s)
			}
		}
	}

	if hasParsedHTML && (stmt.Ticker != "" || stmt.CompanyName != "") {
		finalizeCoreFinancials(stmt)
		return stmt, nil
	}

	return nil, fmt.Errorf("no valid .xbrl, .xml, or inline XBRL .html found in zip archive: %s", zipPath)
}

func mergeStatements(target, source *domain.Statement) {
	if target.Ticker == "" {
		target.Ticker = source.Ticker
	}
	if target.CompanyName == "" {
		target.CompanyName = source.CompanyName
	}
	if target.Year == 0 {
		target.Year = source.Year
	}
	if target.Period == "" {
		target.Period = source.Period
	}
	if target.PeriodType == "" {
		target.PeriodType = source.PeriodType
	}
	if target.Metadata.Sector == "" {
		target.Metadata.Sector = source.Metadata.Sector
	}
	if target.Metadata.Industry == "" {
		target.Metadata.Industry = source.Metadata.Industry
	}
	if target.Metadata.Currency == "" {
		target.Metadata.Currency = source.Metadata.Currency
	}
	if target.Metadata.ConversionRate == 0 {
		target.Metadata.ConversionRate = source.Metadata.ConversionRate
	}

	// Merge core
	if target.Core.TotalAssets == 0 {
		target.Core.TotalAssets = source.Core.TotalAssets
	}
	if target.Core.CashAndEquivalents == 0 {
		target.Core.CashAndEquivalents = source.Core.CashAndEquivalents
	}
	if target.Core.TotalLiabilities == 0 {
		target.Core.TotalLiabilities = source.Core.TotalLiabilities
	}
	if target.Core.TotalEquity == 0 {
		target.Core.TotalEquity = source.Core.TotalEquity
	}
	if target.Core.Revenue == 0 {
		target.Core.Revenue = source.Core.Revenue
	}
	if target.Core.GrossProfit == 0 {
		target.Core.GrossProfit = source.Core.GrossProfit
	}
	if target.Core.OperatingIncome == 0 {
		target.Core.OperatingIncome = source.Core.OperatingIncome
	}
	if target.Core.NetIncome == 0 {
		target.Core.NetIncome = source.Core.NetIncome
	}

	// Merge facts
	for k, v := range source.Facts {
		if target.Facts[k] == nil {
			target.Facts[k] = make(map[string]domain.FactValue)
		}
		for ctxRef, fv := range v {
			target.Facts[k][ctxRef] = fv
		}
	}
}

// ParseInstanceXML parses an XBRL instance stream into a domain Statement
func ParseInstanceXML(r io.Reader) (*domain.Statement, error) {
	decoder := xml.NewDecoder(r)

	stmt := &domain.Statement{
		Metadata: domain.StatementMetadata{
			RoundingMultiplier: 1.0,
		},
		Facts: make(domain.FactMap),
	}

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("xml decoding token: %w", err)
		}

		switch se := token.(type) {
		case xml.StartElement:
			local := se.Name.Local
			space := se.Name.Space

			// Handle inline XBRL (ix:nonNumeric and ix:nonFraction with name="idx-dei:..." or name="idx-cor:...")
			var nameAttr, contextRef, unitRef, decimalsStr string
			var isNil bool

			for _, attr := range se.Attr {
				switch attr.Name.Local {
				case "name":
					nameAttr = attr.Value
				case "contextRef":
					contextRef = attr.Value
				case "unitRef":
					unitRef = attr.Value
				case "decimals":
					decimalsStr = attr.Value
				case "nil":
					isNil = (attr.Value == "true")
				}
			}

			if nameAttr != "" {
				parts := strings.Split(nameAttr, ":")
				if len(parts) == 2 {
					if parts[0] == "idx-dei" {
						var textVal string
						if err := decoder.DecodeElement(&textVal, &se); err == nil {
							assignDEIMetadata(stmt, parts[1], strings.TrimSpace(textVal))
						}
						continue
					} else if parts[0] == "idx-cor" {
						local = parts[1]
						space = "/cor"
					}
				}
			}

			// Extract DEI Metadata from standalone tags
			if strings.Contains(space, "/dei") {
				var textVal string
				if err := decoder.DecodeElement(&textVal, &se); err == nil {
					assignDEIMetadata(stmt, local, strings.TrimSpace(textVal))
				}
				continue
			}

			// Extract Core Financial Facts
			if strings.Contains(space, "/cor") {
				var rawVal string
				if err := decoder.DecodeElement(&rawVal, &se); err == nil {
					rawVal = strings.TrimSpace(rawVal)
					if rawVal != "" && !isNil {
						numVal, err := parseNumericValue(rawVal)
						if err == nil {
							dec, _ := strconv.Atoi(decimalsStr)
							if stmt.Facts[local] == nil {
								stmt.Facts[local] = make(map[string]domain.FactValue)
							}
							stmt.Facts[local][contextRef] = domain.FactValue{
								Value:    numVal,
								Unit:     unitRef,
								Decimals: dec,
							}

							// Populate core financials for primary contexts
							if contextRef == "CurrentYearInstant" || contextRef == "CurrentYearDuration" {
								assignCoreMetric(&stmt.Core, local, numVal)
							}
						}
					}
				}
			}
		}
	}

	// Post-processing: Calculate derived metrics
	finalizeCoreFinancials(stmt)

	return stmt, nil
}

func parseNumericValue(s string) (float64, error) {
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, " ", "")
	return strconv.ParseFloat(s, 64)
}

func assignDEIMetadata(s *domain.Statement, tag, val string) {
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
		rate, _ := parseNumericValue(val)
		s.Metadata.ConversionRate = rate
	case "CurrentPeriodEndDate", "PeriodEndDate", "BalanceSheetDate", "CurrentPeriodStartDate":
		if t, err := parseFlexibleDate(val); err == nil {
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

func assignCoreMetric(c *domain.CoreFinancials, tag string, val float64) {
	switch tag {
	case "Assets":
		c.TotalAssets = val
	case "CashAndCashEquivalents":
		c.CashAndEquivalents = val
	case "CurrentAssets":
		c.CurrentAssets = val
	case "Liabilities":
		c.TotalLiabilities = val
	case "CurrentLiabilities":
		c.CurrentLiabilities = val
	case "ShortTermBankLoans":
		c.ShortTermDebt = val
	case "LongTermBankLoans":
		c.LongTermDebt = val
	case "Equity", "TotalEquity":
		c.TotalEquity = val
	case "RetainedEarningsUnappropriated":
		c.RetainedEarnings = val
	case "SalesAndRevenue", "Revenues", "InterestIncome", "InterestAndFinanceIncome":
		c.Revenue = val
	case "CostOfSalesAndRevenue", "InterestExpense":
		c.CostOfRevenue = val
	case "GrossProfit", "NetInterestIncome":
		c.GrossProfit = val
	case "OperatingIncomeExpense", "OperatingIncome":
		c.OperatingIncome = val
	case "FinanceCosts", "InterestAndFinanceCosts":
		c.FinanceCosts = val
	case "ProfitLoss":
		c.NetIncome = val
	case "ProfitLossAttributableToOwnersOfParentEntity":
		c.NetIncomeParent = val
	case "NetCashFlowsReceivedFromUsedInOperatingActivities", "NetCashFlowsFromUsedInOperatingActivities", "CashGeneratedFromUsedInOperations":
		if c.OperatingCashFlow == 0 {
			c.OperatingCashFlow = val
		}
	case "NetCashFlowsReceivedFromUsedInInvestingActivities", "NetCashFlowsFromUsedInInvestingActivities":
		c.InvestingCashFlow = val
	case "NetCashFlowsReceivedFromUsedInFinancingActivities", "NetCashFlowsFromUsedInFinancingActivities":
		c.FinancingCashFlow = val
	case "PaymentsForPropertyPlantEquipment", "PaymentsForAcquisitionOfPropertyPlantAndEquipment", "AdditionInPropertyPlantAndEquipment":
		c.CapEx = val
	case "DistributionsOfCashDividends", "DividendsPaidFromFinancingActivities":
		c.DividendsPaid = val
	case "WeightedAverageShares", "NumberOfIssuedAndFullyPaidShares":
		c.SharesOutstanding = val
	}
}

func finalizeCoreFinancials(s *domain.Statement) {
	c := &s.Core
	if c.TotalDebt == 0 {
		c.TotalDebt = c.ShortTermDebt + c.LongTermDebt
	}
	if c.WorkingCapital == 0 && c.CurrentAssets != 0 {
		c.WorkingCapital = c.CurrentAssets - c.CurrentLiabilities
	}
	if c.FreeCashFlow == 0 && c.OperatingCashFlow != 0 {
		c.FreeCashFlow = c.OperatingCashFlow - c.CapEx
	}
	if c.EBITDA == 0 && c.OperatingIncome != 0 {
		c.EBITDA = c.OperatingIncome + (c.CapEx * 0.7) // Estimate D&A if not explicitly separated
	}
}

func parseFlexibleDate(val string) (time.Time, error) {
	val = strings.TrimSpace(val)
	formats := []string{
		"2006-01-02",
		"January 02, 2006",
		"January 2, 2006",
		"02 January 2006",
		"2 January 2006",
		"02-01-2006",
		"02/01/2006",
		"September 30, 2006",
		"March 31, 2006",
		"June 30, 2006",
		"December 31, 2006",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, val); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unknown date format: %s", val)
}
