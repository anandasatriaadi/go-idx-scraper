package main

import (
	"reflect"
	"testing"
	"time"
)

func TestParseYears(t *testing.T) {
	currentYear := time.Now().Year()

	tests := []struct {
		name        string
		input       string
		defaultYear string
		expected    []int
		expectErr   bool
	}{
		{
			name:        "Count of 5 years",
			input:       "5",
			defaultYear: "",
			expected: []int{
				currentYear - 4,
				currentYear - 3,
				currentYear - 2,
				currentYear - 1,
				currentYear,
			},
			expectErr: false,
		},
		{
			name:        "Year range 2021-2025",
			input:       "2021-2025",
			defaultYear: "",
			expected:    []int{2021, 2022, 2023, 2024, 2025},
			expectErr:   false,
		},
		{
			name:        "Comma separated years",
			input:       "2022, 2023, 2024",
			defaultYear: "",
			expected:    []int{2022, 2023, 2024},
			expectErr:   false,
		},
		{
			name:        "Single specific year",
			input:       "2024",
			defaultYear: "",
			expected:    []int{2024},
			expectErr:   false,
		},
		{
			name:        "Empty input with default year",
			input:       "",
			defaultYear: "2023",
			expected:    []int{2023},
			expectErr:   false,
		},
		{
			name:        "Invalid range format",
			input:       "2021-2025-2026",
			defaultYear: "",
			expected:    nil,
			expectErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseYears(tt.input, tt.defaultYear)
			if tt.expectErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if !reflect.DeepEqual(result, tt.expected) {
					t.Errorf("expected %v, got %v", tt.expected, result)
				}
			}
		})
	}
}

func TestParsePeriods(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Default periods",
			input:    "",
			expected: []string{"TW1", "TW2", "TW3", "Audit"},
		},
		{
			name:     "Custom periods",
			input:    "TW1, TW2, Audit",
			expected: []string{"TW1", "TW2", "Audit"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parsePeriods(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestPeriodRank(t *testing.T) {
	if periodRank("TW1") >= periodRank("TW2") {
		t.Errorf("TW1 rank should be less than TW2")
	}
	if periodRank("TW2") >= periodRank("TW3") {
		t.Errorf("TW2 rank should be less than TW3")
	}
	if periodRank("TW3") >= periodRank("Audit") {
		t.Errorf("TW3 rank should be less than Audit")
	}
	if periodRank("Audit") != periodRank("Tahunan") {
		t.Errorf("Audit and Tahunan should have equal rank")
	}
}

func TestFormattingHelpers(t *testing.T) {
	if got := formatIDR(1.5e12); got != "1.50 T" {
		t.Errorf("expected '1.50 T', got '%s'", got)
	}
	if got := formatIDR(2.35e9); got != "2.35 B" {
		t.Errorf("expected '2.35 B', got '%s'", got)
	}
	if got := formatIDR(4.5e6); got != "4.50 M" {
		t.Errorf("expected '4.50 M', got '%s'", got)
	}
	if got := formatIDR(0); got != "-" {
		t.Errorf("expected '-', got '%s'", got)
	}

	if got := formatSignedPct(25.5); got != "+25.50%" {
		t.Errorf("expected '+25.50%%', got '%s'", got)
	}
	if got := formatSignedPct(-15.2); got != "-15.20%" {
		t.Errorf("expected '-15.20%%', got '%s'", got)
	}
}
