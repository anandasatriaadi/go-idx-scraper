package main

import (
	"testing"
)

func TestDefaultCollections(t *testing.T) {
	expected := []string{
		"xbrl_statements",
		"stock_prices",
		"financial_reports",
		"news",
		"daily_briefings",
		"announcements",
		"last_runs",
	}

	if len(defaultCollections) != len(expected) {
		t.Fatalf("expected %d default collections, got %d", len(expected), len(defaultCollections))
	}

	foundMap := make(map[string]bool)
	for _, col := range defaultCollections {
		foundMap[col] = true
	}

	for _, exp := range expected {
		if !foundMap[exp] {
			t.Errorf("missing expected collection in defaults: %s", exp)
		}
	}
}

func TestResolveConfigPath(t *testing.T) {
	custom := "config/custom.yml"
	if got := resolveConfigPath(custom); got != custom {
		t.Errorf("expected %s, got %s", custom, got)
	}
}
