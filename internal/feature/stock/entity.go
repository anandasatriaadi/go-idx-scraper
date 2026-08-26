package stock

import (
	"context"
	"time"
)

// StockData represents stock data.
type StockData struct {
	Code string `json:"StockCode"`
}

// StockListResponse represents API response.
type StockListResponse struct {
	Data []StockData `json:"data"`
}

// PriceCandle represents daily OHLCV price data for a stock or currency.
type PriceCandle struct {
	ID        string    `bson:"_id,omitempty" json:"id,omitempty"`
	Ticker    string    `bson:"ticker" json:"ticker"`
	Date      time.Time `bson:"date" json:"date"`
	Open      float64   `bson:"open" json:"open"`
	High      float64   `bson:"high" json:"high"`
	Low       float64   `bson:"low" json:"low"`
	Close     float64   `bson:"close" json:"close"`
	AdjClose  float64   `bson:"adj_close" json:"adj_close"`
	Volume    int64     `bson:"volume" json:"volume"`
	CreatedAt time.Time `bson:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedAt time.Time `bson:"updated_at,omitempty" json:"updated_at,omitempty"`
}

type Repository interface {
	// Add methods as needed, currently a domain data structure for parsing
}

// PriceRepository defines price persistence port.
type PriceRepository interface {
	UpsertCandles(ctx context.Context, ticker string, candles []PriceCandle) error
	GetPrices(ctx context.Context, ticker string, limit int) ([]*PriceCandle, error)
}
