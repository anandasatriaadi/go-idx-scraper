package xbrl

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestStatementEntity_Fields(t *testing.T) {
	id := bson.NewObjectID()
	now := time.Now()

	s := &Statement{
		ID:            id,
		Ticker:        "AADI",
		CompanyName:   "PT Adaro Andalan Indonesia Tbk",
		Year:          2026,
		Period:        "Q1",
		PeriodEndDate: now,
		Metadata: StatementMetadata{
			Sector:             "A. Energy",
			Industry:           "A12. Coal",
			Currency:           "USD",
			RoundingMultiplier: 1000,
			AuditStatus:        "Tidak Diaudit / Unaudit",
		},
		Core: CoreFinancials{
			TotalAssets:       5780540000,
			TotalLiabilities:  1999310000,
			TotalEquity:       3781230000,
			Revenue:           1044192000,
			NetIncome:         153768000,
			OperatingCashFlow: 285400000,
			CapEx:             62300000,
			FreeCashFlow:      223100000,
		},
		ComputedRatios: ComputedRatios{
			ROIC:            0.185,
			ROE:             0.198,
			PiotroskiFScore: 8,
			AltmanZScore:    3.45,
		},
		Valuation: ValuationMetrics{
			NormalizedEPS:    790.15,
			NormalizedBVPS:   4624.80,
			GrahamNumber:     9060.50,
			CurrentPrice:     4150.0,
			MarginOfSafetyPct: 54.2,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if s.Ticker != "AADI" {
		t.Errorf("Expected Ticker AADI, got %s", s.Ticker)
	}
	if s.Core.TotalAssets != 5780540000 {
		t.Errorf("Expected TotalAssets 5780540000, got %f", s.Core.TotalAssets)
	}
	if s.ComputedRatios.PiotroskiFScore != 8 {
		t.Errorf("Expected Piotroski F-Score 8, got %d", s.ComputedRatios.PiotroskiFScore)
	}
	if s.Valuation.MarginOfSafetyPct <= 0 {
		t.Errorf("Expected positive MarginOfSafetyPct")
	}
}
