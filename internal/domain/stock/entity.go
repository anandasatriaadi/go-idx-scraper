package stock

// StockData represents stock data.
type StockData struct {
	Code string `json:"StockCode"`
}

// StockListResponse represents API response.
type StockListResponse struct {
	Data []StockData `json:"data"`
}

type Repository interface {
	// Add methods as needed, currently likely just a data structure for parsing
}
