package main

import (
	"testing"

	"github.com/anandasatriaadi/go-idx-scraper/internal/config"
	"go.uber.org/zap"
)

func TestPreparePeriodParams(t *testing.T) {
	logger := zap.NewNop()

	// Test case 1: Normal input
	cfg1 := &config.Config{
		Download: config.DownloadConfig{
			MonthPeriod: "3",
			Mode:        "TW",
		},
	}
	period1, modePeriod1 := preparePeriodParams(cfg1, logger)
	if period1 != "III" {
		t.Errorf("Expected period 'III', got '%s'", period1)
	}
	if modePeriod1 != "TW3" {
		t.Errorf("Expected modePeriod 'TW3', got '%s'", modePeriod1)
	}

	// Test case 2: Invalid month (should default to 0)
	cfg2 := &config.Config{
		Download: config.DownloadConfig{
			MonthPeriod: "invalid",
			Mode:        "TW",
		},
	}
	period2, modePeriod2 := preparePeriodParams(cfg2, logger)
	if period2 != "" { // 0 'I's
		t.Errorf("Expected period '', got '%s'", period2)
	}
	if modePeriod2 != "TW0" {
		t.Errorf("Expected modePeriod 'TW0', got '%s'", modePeriod2)
	}
}

func TestCheckPageTitleForErrors(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		title     string
		expectErr bool
		errString string
	}{
		{"Financial Report 2024", false, ""},
		{"404 Not Found", true, "not found"},
		{"Document Not Found", true, "not found"},
		{"503 Service Unavailable", true, "server error"},
		{"Just a moment...", true, "bot detector"},
		{"Attention Required!", true, "bot detector"},
	}

	for _, tt := range tests {
		err := checkPageTitleForErrors(tt.title, "StockA", logger)
		if tt.expectErr {
			if err == nil {
				t.Errorf("Expected error for title '%s', got nil", tt.title)
			} else if err.Error() != tt.errString {
				t.Errorf("Expected error '%s' for title '%s', got '%s'", tt.errString, tt.title, err.Error())
			}
		} else {
			if err != nil {
				t.Errorf("Expected no error for title '%s', got '%s'", tt.title, err.Error())
			}
		}
	}
}
