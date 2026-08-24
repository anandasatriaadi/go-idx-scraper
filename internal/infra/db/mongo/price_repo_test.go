package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/stock"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type MockPriceRepo struct {
	Candles map[string][]stock.PriceCandle
	Err     error
}

func NewMockPriceRepo() *MockPriceRepo {
	return &MockPriceRepo{
		Candles: make(map[string][]stock.PriceCandle),
	}
}

func (m *MockPriceRepo) UpsertCandles(ctx context.Context, ticker string, candles []stock.PriceCandle) error {
	if m.Err != nil {
		return m.Err
	}
	m.Candles[ticker] = append(m.Candles[ticker], candles...)
	return nil
}

func (m *MockPriceRepo) GetPrices(ctx context.Context, ticker string, limit int) ([]*stock.PriceCandle, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	items := m.Candles[ticker]
	results := make([]*stock.PriceCandle, 0, len(items))
	for i := range items {
		results = append(results, &items[i])
	}
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}
	return results, nil
}

func TestPriceRepository_InterfaceCompliance(t *testing.T) {
	var _ stock.PriceRepository = (*PriceRepository)(nil)
	var _ stock.PriceRepository = (*MockPriceRepo)(nil)
}

func TestPriceCandle_StructAndMockRepo(t *testing.T) {
	mock := NewMockPriceRepo()
	ctx := context.Background()

	now := time.Now().UTC()
	candles := []stock.PriceCandle{
		{
			ID:        bson.NewObjectID(),
			Ticker:    "BBRI",
			Date:      now,
			Open:      3100,
			High:      3150,
			Low:       3080,
			Close:     3120,
			AdjClose:  3120,
			Volume:    15000000,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	err := mock.UpsertCandles(ctx, "BBRI", candles)
	if err != nil {
		t.Fatalf("UpsertCandles failed: %v", err)
	}

	prices, err := mock.GetPrices(ctx, "BBRI", 10)
	if err != nil {
		t.Fatalf("GetPrices failed: %v", err)
	}

	if len(prices) != 1 {
		t.Fatalf("Expected 1 price candle, got %d", len(prices))
	}

	if prices[0].Ticker != "BBRI" || prices[0].Close != 3120 {
		t.Errorf("Price candle data mismatch: %+v", prices[0])
	}
}
