package types

// StockData represents stock data.
type StockData struct {
	Code string `json:"StockCode"`
}

// StockListResponse represents API response.
type StockListResponse struct {
	Data []StockData `json:"data"`
}
