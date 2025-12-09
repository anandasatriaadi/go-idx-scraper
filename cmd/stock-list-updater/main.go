package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/anandasatriaadi/go-idx-scraper/internal/browser"
	"github.com/anandasatriaadi/go-idx-scraper/internal/config"
	"github.com/anandasatriaadi/go-idx-scraper/internal/helper"
	"github.com/anandasatriaadi/go-idx-scraper/internal/types"
	"github.com/chromedp/chromedp"
)

// stringSliceToSet converts a slice of strings into a set-like map.
// Each string becomes a key in the map with true as its value.
//
// Parameters:
// - strs: A slice of strings.
//
// Returns:
// - A map where each string from the slice is a key.
func stringSliceToSet(strs []string) map[string]bool {
	set := make(map[string]bool)
	for _, s := range strs {
		set[s] = true
	}
	return set
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: program <config-file>")
	}

	// Read configuration file provided in args
	cfg, err := config.Load(os.Args[1])
	if err != nil {
		log.Fatalf("Failed to read config: %v\n", err)
	}

	ctx, cancel := browser.SetupChromeDp(cfg)
	defer cancel()

	// Load current list of stocks and make it into a set (in this case a map)
	currStocksList := helper.LoadCurrent(cfg.Paths.StockList)
	currStocksSet := stringSliceToSet(currStocksList)

	jsonData := fetchStocks(&ctx)
	for _, stock := range jsonData {
		if !currStocksSet[stock.Code] {
			currStocksList = append(currStocksList, stock.Code)
			currStocksSet[stock.Code] = true
		}
	}

	// Save stocks list appended with new stocks
	saveStock(cfg.Paths.StockList, currStocksList)
}

func fetchStocks(ctx *context.Context) []types.StockData {
	url := fmt.Sprintf("https://www.idx.co.id/primary/TradingSummary/GetStockSummary?length=9999&start=0&date=%s", time.Now().AddDate(0, 0, -1).Format("20060102"))

	fmt.Println("Url", url)
	var data string
	err := chromedp.Run(*ctx, getPageData(url, &data))
	fmt.Println("resp", data)
	if err != nil {
		slog.Error("Running chromedp", "error", err)
		return nil
	}
	var resp types.StockListResponse
	json.Unmarshal([]byte(data), &resp)
	return resp.Data
}

// saveStock saves stocks.
func saveStock(filePath string, stocks []string) {
	data, _ := json.MarshalIndent(stocks, "", "  ")
	os.WriteFile(filePath, data, 0644)
}

func getPageData(urlstr string, res *string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.Evaluate(`document.body.innerText`, res),
	}
}
