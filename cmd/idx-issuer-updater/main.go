package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/anandasatriaadi/go-idx-scraper/internal/browser"
	"github.com/anandasatriaadi/go-idx-scraper/internal/config"
	"github.com/anandasatriaadi/go-idx-scraper/internal/helper"
	"github.com/anandasatriaadi/go-idx-scraper/internal/types"
	"github.com/chromedp/chromedp"
	"go.uber.org/zap"
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
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	if len(os.Args) < 2 {
		logger.Error("Usage: program <config-file>")
		os.Exit(1)
	}

	// Read configuration file provided in args
	cfg, err := config.Load(os.Args[1])
	if err != nil {
		logger.Error("Failed to read config", zap.Error(err))
		os.Exit(1)
	}

	ctx, cancel := browser.SetupChromeDp(cfg)
	defer cancel()

	// Load current list of stocks and make it into a set (in this case a map)
	currIssuerList, err := helper.LoadCurrent(cfg.Paths.IssuerList, logger)
	if err != nil {
		logger.Error("Failed to load current issuer list", zap.Error(err))
		os.Exit(1)
	}
	currIssuerSet := stringSliceToSet(currIssuerList)

	jsonData, err := fetchStocks(ctx)
	if err != nil {
		logger.Error("Failed to fetch stocks", zap.Error(err))
		os.Exit(1)
	}
	for _, stock := range jsonData {
		if !currIssuerSet[stock.Code] {
			currIssuerList = append(currIssuerList, stock.Code)
			currIssuerSet[stock.Code] = true
		}
	}

	// Save stocks list appended with new stocks
	if err := saveIssuer(cfg.Paths.IssuerList, currIssuerList); err != nil {
		logger.Error("Failed to save issuer list", zap.Error(err))
		os.Exit(1)
	}
	logger.Info("Issuer list updated", zap.Int("total", len(currIssuerList)))
}

func fetchStocks(ctx context.Context) ([]types.StockData, error) {
	url := fmt.Sprintf("https://www.idx.co.id/primary/TradingSummary/GetStockSummary?length=9999&start=0&date=%s", time.Now().AddDate(0, 0, -1).Format("20060102"))

	var data string
	err := chromedp.Run(ctx, getPageData(url, &data))
	if err != nil {
		return nil, fmt.Errorf("running chromedp: %w", err)
	}
	var resp types.StockListResponse
	if err := json.Unmarshal([]byte(data), &resp); err != nil {
		return nil, fmt.Errorf("unmarshaling response: %w", err)
	}
	return resp.Data, nil
}

// saveIssuer saves stocks.
func saveIssuer(filePath string, stocks []string) error {
	data, err := json.MarshalIndent(stocks, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling stocks: %w", err)
	}
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}
	return nil
}

func getPageData(urlstr string, res *string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.Evaluate(`document.body.innerText`, res),
	}
}
