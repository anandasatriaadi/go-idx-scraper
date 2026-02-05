package stock

// StockData represents stock data.
type StockData struct {
	Code string `json:"StockCode"`
}

type Repository interface {
	// Add methods as needed, currently likely just a data structure for parsing
}
