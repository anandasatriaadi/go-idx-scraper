package helper

import (
	"os"
	"testing"

	"go.uber.org/zap"
)

func TestLoadCurrent(t *testing.T) {
	logger := zap.NewNop()

	t.Run("File not exists", func(t *testing.T) {
		tempFile := "test_stocks.json"
		defer os.Remove(tempFile)

		stocks, err := LoadCurrent(tempFile, logger)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(stocks) != 0 {
			t.Errorf("Expected 0 stocks, got %d", len(stocks))
		}
	})

	t.Run("File exists", func(t *testing.T) {
		tempFile := "test_stocks_exists.json"
		defer os.Remove(tempFile)
		os.WriteFile(tempFile, []byte(`["BBCA", "BBRI"]`), 0644)

		stocks, err := LoadCurrent(tempFile, logger)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(stocks) != 2 {
			t.Errorf("Expected 2 stocks, got %d", len(stocks))
		}
		if stocks[0] != "BBCA" || stocks[1] != "BBRI" {
			t.Errorf("Unexpected stocks: %v", stocks)
		}
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		tempFile := "test_stocks_invalid.json"
		defer os.Remove(tempFile)
		os.WriteFile(tempFile, []byte(`invalid`), 0644)

		_, err := LoadCurrent(tempFile, logger)
		if err == nil {
			t.Error("Expected error for invalid JSON, got nil")
		}
	})
}
