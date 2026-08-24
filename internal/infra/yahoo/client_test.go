package yahoo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNormalizeSymbol(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"BBRI", "BBRI.JK"},
		{"bbri", "BBRI.JK"},
		{"BBRI.JK", "BBRI.JK"},
		{"bbri.jk", "BBRI.JK"},
		{"TLKM", "TLKM.JK"},
		{"USDIDR", "USDIDR=X"},
		{"usdidr", "USDIDR=X"},
		{"USDIDR=X", "USDIDR=X"},
		{"^JKSE", "^JKSE"},
		{"AAPL", "AAPL.JK"}, // IDX context appends .JK unless symbol has dot/equals/^
		{"", ""},
	}

	for _, tt := range tests {
		got := NormalizeSymbol(tt.input)
		if got != tt.expected {
			t.Errorf("NormalizeSymbol(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}

func TestCleanTicker(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"BBRI.JK", "BBRI"},
		{"bbri.jk", "BBRI"},
		{"BBRI", "BBRI"},
		{"USDIDR=X", "USDIDR=X"},
		{"USDIDR", "USDIDR"},
	}

	for _, tt := range tests {
		got := CleanTicker(tt.input)
		if got != tt.expected {
			t.Errorf("CleanTicker(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}

func TestFetchHistoricalPrices_Success(t *testing.T) {
	sampleJSON := `{
		"chart": {
			"result": [
				{
					"meta": {
						"currency": "IDR",
						"symbol": "BBRI.JK",
						"regularMarketPrice": 3180.0
					},
					"timestamp": [1704067200, 1704153600, 1704240000],
					"indicators": {
						"quote": [
							{
								"open": [3100.0, 3120.0, null],
								"high": [3150.0, 3200.0, null],
								"low": [3080.0, 3100.0, null],
								"close": [3120.0, 3180.0, null],
								"volume": [15000000, 20000000, null]
							}
						],
						"adjclose": [
							{
								"adjclose": [3120.0, 3180.0, null]
							}
						]
					}
				}
			],
			"error": null
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Errorf("Expected User-Agent header to be set")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(sampleJSON))
	}))
	defer server.Close()

	client := NewClient(
		WithBaseURL(server.URL),
		WithHTTPClient(server.Client()),
		WithLogger(zap.NewNop()),
	)

	ctx := context.Background()
	candles, err := client.FetchHistoricalPricesWithContext(ctx, "BBRI", "5d")
	if err != nil {
		t.Fatalf("FetchHistoricalPrices failed: %v", err)
	}

	// The 3rd candle had null quotes and should be skipped
	if len(candles) != 2 {
		t.Fatalf("Expected 2 candles, got %d", len(candles))
	}

	c1 := candles[0]
	if c1.Ticker != "BBRI" {
		t.Errorf("Expected ticker BBRI, got %s", c1.Ticker)
	}
	if c1.Open != 3100.0 || c1.High != 3150.0 || c1.Low != 3080.0 || c1.Close != 3120.0 || c1.AdjClose != 3120.0 || c1.Volume != 15000000 {
		t.Errorf("Candle 1 data mismatch: %+v", c1)
	}
	expectedDate1 := time.Unix(1704067200, 0).UTC()
	if !c1.Date.Equal(expectedDate1) {
		t.Errorf("Expected date %v, got %v", expectedDate1, c1.Date)
	}

	c2 := candles[1]
	if c2.Close != 3180.0 || c2.Volume != 20000000 {
		t.Errorf("Candle 2 data mismatch: %+v", c2)
	}
}

func TestFetchHistoricalPrices_ErrorResponse(t *testing.T) {
	errorJSON := `{
		"chart": {
			"result": null,
			"error": {
				"code": "Not Found",
				"description": "No data found for symbol INVALID"
			}
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(errorJSON))
	}))
	defer server.Close()

	client := NewClient(
		WithBaseURL(server.URL),
		WithHTTPClient(server.Client()),
		WithLogger(zap.NewNop()),
	)

	_, err := client.FetchHistoricalPrices("INVALID", "1mo")
	if err == nil {
		t.Fatal("Expected error for not found ticker, got nil")
	}
}

func TestFetchUSDIDR_Success(t *testing.T) {
	sampleJSON := `{
		"chart": {
			"result": [
				{
					"meta": {
						"currency": "IDR",
						"symbol": "USDIDR=X",
						"regularMarketPrice": 16250.0
					},
					"timestamp": [1704067200],
					"indicators": {
						"quote": [
							{
								"open": [16200.0],
								"high": [16300.0],
								"low": [16150.0],
								"close": [16250.0],
								"volume": [0]
							}
						],
						"adjclose": [
							{
								"adjclose": [16250.0]
							}
						]
					}
				}
			],
			"error": null
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(sampleJSON))
	}))
	defer server.Close()

	client := NewClient(
		WithBaseURL(server.URL),
		WithHTTPClient(server.Client()),
		WithLogger(zap.NewNop()),
	)

	rate, err := client.FetchUSDIDR(context.Background())
	if err != nil {
		t.Fatalf("FetchUSDIDR failed: %v", err)
	}

	if rate != 16250.0 {
		t.Errorf("Expected rate 16250.0, got %f", rate)
	}
}
