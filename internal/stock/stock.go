package stock

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/tebeka/selenium"
)

// StockData represents stock data.
type StockData struct {
	Code string `json:"StockCode"`
}

// StockListResponse represents API response.
type StockListResponse struct {
	Data []StockData `json:"data"`
}

// LoadCurrent loads stocks from file.
func LoadCurrent(filePath string) []string {
	data, err := os.ReadFile(filePath)
	if err != nil {
		slog.Error("Loading stocks", "error", err)
		return nil
	}
	var stocks []string
	json.Unmarshal(data, &stocks)
	return stocks
}

// FetchNewStocks fetches from IDX.
func FetchNewStocks(br selenium.WebDriver) []StockData {
	url := fmt.Sprintf("https://www.idx.co.id/primary/TradingSummary/GetStockSummary?length=9999&start=0&date=%s", time.Now().Format("20060102"))
	br.Get(url)
	elem, err := br.FindElement(selenium.ByTagName, "pre")
	if err != nil {
		slog.Error("Finding element", "error", err)
		return nil
	}
	text, _ := elem.Text()
	var resp StockListResponse
	json.Unmarshal([]byte(text), &resp)
	return resp.Data
}

// SaveUpdatedStocks saves stocks.
func SaveUpdatedStocks(filePath string, stocks []string) {
	data, _ := json.MarshalIndent(stocks, "", "  ")
	os.WriteFile(filePath, data, 0644)
}
