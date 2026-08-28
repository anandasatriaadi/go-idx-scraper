package parser

import (
	domain "github.com/anandasatriaadi/go-idx-scraper/internal/feature/xbrl"
)

// AssignSharesMetric maps shares outstanding line item taxonomy tags to CoreFinancials
func AssignSharesMetric(c *domain.CoreFinancials, tag string, val float64) {
	switch tag {
	case "WeightedAverageShares", "NumberOfIssuedAndFullyPaidShares":
		c.SharesOutstanding = val
	}
}

// FinalizeSharesOutstanding cascades through raw FactMap tags if SharesOutstanding is not yet populated
func FinalizeSharesOutstanding(s *domain.Statement) {
	c := &s.Core
	if c.SharesOutstanding != 0 {
		return
	}

	tags := []string{
		"NumberOfIssuedAndFullyPaidShares",
		"WeightedAverageShares",
		"NumberOfSharesOutstanding",
		"IssuedAndFullyPaidShares",
		"CommonStocksNumberOfShares",
		"EntitySharesOutstanding",
		"CapitalStockNumberOfShares",
		"TotalShares",
		"SharesOutstanding",
	}

	for _, tag := range tags {
		if m, ok := s.Facts[tag]; ok {
			for _, fv := range m {
				if fv.Value > 1000 {
					c.SharesOutstanding = fv.Value
					return
				}
			}
		}
	}
}
