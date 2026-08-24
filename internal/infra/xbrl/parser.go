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

// ParseInstanceZip opens an instance.zip file and parses the instance.xbrl inside
func ParseInstanceZip(zipPath string) (*domain.Statement, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("opening zip file: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		if strings.HasSuffix(strings.ToLower(f.Name), ".xbrl") || strings.HasSuffix(strings.ToLower(f.Name), ".xml") {
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
	return nil, fmt.Errorf("no .xbrl or .xml instance file found in zip archive: %s", zipPath)
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

			// Extract DEI Metadata
			if strings.Contains(space, "/dei") {
				var textVal string
				if err := decoder.DecodeElement(&textVal, &se); err == nil {
					assignDEIMetadata(stmt, local, strings.TrimSpace(textVal))
				}
				continue
			}

			// Extract Core Financial Facts
			if strings.Contains(space, "/cor") {
				var contextRef, unitRef, decimalsStr string
				var isNil bool

				for _, attr := range se.Attr {
					switch attr.Name.Local {
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
	case "CurrentPeriodEndDate":
		t, err := time.Parse("2006-01-02", val)
		if err == nil {
			s.PeriodEndDate = t
			s.Year = t.Year()
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
	case "SalesAndRevenue", "Revenues":
		c.Revenue = val
	case "CostOfSalesAndRevenue":
		c.CostOfRevenue = val
	case "GrossProfit":
		c.GrossProfit = val
	case "OperatingIncomeExpense":
		c.OperatingIncome = val
	case "FinanceCosts":
		c.FinanceCosts = val
	case "ProfitLoss":
		c.NetIncome = val
	case "ProfitLossAttributableToOwnersOfParentEntity":
		c.NetIncomeParent = val
	case "NetCashFlowsFromUsedInOperatingActivities":
		c.OperatingCashFlow = val
	case "PaymentsForPropertyPlantEquipment":
		c.CapEx = val
	case "WeightedAverageShares", "NumberOfIssuedAndFullyPaidShares":
		c.SharesOutstanding = val
	}
}

func finalizeCoreFinancials(s *domain.Statement) {
	c := &s.Core
	if c.TotalDebt == 0 {
		c.TotalDebt = c.ShortTermDebt + c.LongTermDebt
	}
	if c.FreeCashFlow == 0 && c.OperatingCashFlow != 0 {
		c.FreeCashFlow = c.OperatingCashFlow - c.CapEx
	}
}
