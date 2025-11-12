package main

import (
	"log"
	"os"

	"github.com/anandasatriaadi/go-idx-scraper/internal/browser"
	"github.com/anandasatriaadi/go-idx-scraper/internal/config"
	"github.com/anandasatriaadi/go-idx-scraper/internal/stock"
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

	// Create a new browser session
	service, browser := browser.SetupBrowser(*cfg)
	defer service.Stop()
	defer browser.Quit()

	// Load current list of stocks and make it into a set (in this case a map)
	currStocksList := stock.LoadCurrent(cfg.Paths.StockList)
	currStocksSet := stringSliceToSet(currStocksList)

	jsonData := stock.FetchNewStocks(browser)
	for _, stock := range jsonData {
		if !currStocksSet[stock.Code] {
			currStocksList = append(currStocksList, stock.Code)
			currStocksSet[stock.Code] = true
		}
	}

	// Save stocks list appended with new stocks
	stock.SaveUpdatedStocks(cfg.Paths.StockList, currStocksList)
}
