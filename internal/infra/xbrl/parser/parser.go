package parser

import (
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"

	domain "github.com/anandasatriaadi/go-idx-scraper/internal/feature/xbrl"
)

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
							AssignDEIMetadata(stmt, parts[1], strings.TrimSpace(textVal))
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
					AssignDEIMetadata(stmt, local, strings.TrimSpace(textVal))
				}
				continue
			}

			// Extract Core Financial Facts
			if strings.Contains(space, "/cor") {
				var rawVal string
				if err := decoder.DecodeElement(&rawVal, &se); err == nil {
					rawVal = strings.TrimSpace(rawVal)
					if rawVal != "" && !isNil {
						numVal, err := ParseNumericValue(rawVal)
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
								AssignCoreMetric(stmt, local, numVal)
							}
						}
					}
				}
			}
		}
	}

	// Post-processing: Calculate derived metrics
	FinalizeCoreFinancials(stmt)

	return stmt, nil
}

// ParseNumericValue parses strings with commas or spaces into float64
func ParseNumericValue(s string) (float64, error) {
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, " ", "")
	return strconv.ParseFloat(s, 64)
}

// AssignCoreMetric delegates metric assignment to balance sheet, income statement, cash flow, and shares mappers
func AssignCoreMetric(stmt *domain.Statement, tag string, val float64) {
	AssignBalanceSheetMetric(&stmt.Core, tag, val)
	AssignIncomeStatementMetric(stmt, tag, val)
	AssignCashFlowMetric(&stmt.Core, tag, val)
	AssignSharesMetric(&stmt.Core, tag, val)
}

// FinalizeCoreFinancials delegates derived financial metric calculation to focused finalizers
func FinalizeCoreFinancials(s *domain.Statement) {
	FinalizeBalanceSheet(&s.Core)
	FinalizeIncomeStatement(&s.Core)
	FinalizeCashFlow(&s.Core)
	FinalizeSharesOutstanding(s)
}
