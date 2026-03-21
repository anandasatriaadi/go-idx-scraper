package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	br "github.com/anandasatriaadi/go-idx-scraper/internal/browser"
	"github.com/anandasatriaadi/go-idx-scraper/internal/config"
	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/stock"
	"github.com/anandasatriaadi/go-idx-scraper/internal/helper"
	"github.com/tebeka/selenium"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		log.Printf("Application failed: %v", err)
		os.Exit(1)
	}
}

func run() error {
	var configPath string
	var noHeadless bool
	flag.StringVar(&configPath, "config", "config/config.yml", "Path to configuration file")
	flag.BoolVar(&noHeadless, "no-headless", false, "Disable headless mode for browser")
	flag.Parse()

	logger, err := helper.NewLogger("issuer")
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}
	defer logger.Sync()

	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Error("Failed to read config", zap.Error(err))
		return err
	}

	if noHeadless {
		cfg.SetHeadless(false)
	}

	browser, err := br.SetupSelenium(cfg)
	if err != nil {
		logger.Error("Failed to setup selenium", zap.Error(err))
		return err
	}
	defer browser.Close()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		logger.Info("Received shutdown signal", zap.String("signal", sig.String()))
	}()

	currIssuerList, err := helper.LoadCurrent(cfg.Paths.IssuerList, logger)
	if err != nil {
		logger.Error("Failed to load current issuer list", zap.Error(err))
		return err
	}
	currIssuerSet := stringSliceToSet(currIssuerList)

	jsonData, err := fetchStocks(browser.Driver)
	if err != nil {
		logger.Warn("Failed to fetch stocks", zap.Error(err))
		return err
	}
	for _, s := range jsonData {
		if !currIssuerSet[s.Code] {
			currIssuerList = append(currIssuerList, s.Code)
			currIssuerSet[s.Code] = true
		}
	}

	if err := saveIssuer(cfg.Paths.IssuerList, currIssuerList); err != nil {
		logger.Error("Failed to save issuer list", zap.Error(err))
		return err
	}
	logger.Info("Issuer list updated", zap.Int("total", len(currIssuerList)))
	return nil
}

func stringSliceToSet(strs []string) map[string]bool {
	set := make(map[string]bool)
	for _, s := range strs {
		set[s] = true
	}
	return set
}

func fetchStocks(driver selenium.WebDriver) ([]stock.StockData, error) {
	url := fmt.Sprintf("https://www.idx.co.id/primary/TradingSummary/GetStockSummary?length=9999&start=0&date=%s", time.Now().AddDate(0, 0, -1).Format("20060102"))

	if err := driver.Get(url); err != nil {
		return nil, fmt.Errorf("navigating to url: %w", err)
	}
	time.Sleep(1 * time.Second)

	body, err := driver.FindElement(selenium.ByTagName, "body")
	if err != nil {
		return nil, fmt.Errorf("finding body: %w", err)
	}
	data, err := body.Text()
	if err != nil {
		return nil, fmt.Errorf("getting body text: %w", err)
	}

	var resp stock.StockListResponse
	if err := json.Unmarshal([]byte(data), &resp); err != nil {
		return nil, fmt.Errorf("unmarshaling response: %w", err)
	}
	return resp.Data, nil
}

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
