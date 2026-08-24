package yahoo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/stock"
	"go.uber.org/zap"
)

const (
	defaultBaseURL   = "https://query1.finance.yahoo.com"
	defaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	defaultTimeout   = 15 * time.Second
)

// PriceCandle is an alias for the domain model stock.PriceCandle.
type PriceCandle = stock.PriceCandle

// Client handles interaction with Yahoo Finance API.
type Client struct {
	httpClient *http.Client
	baseURL    string
	userAgent  string
	logger     *zap.Logger
}

// Option configures Client.
type Option func(*Client)

// WithHTTPClient sets custom http.Client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithBaseURL sets custom Yahoo Finance base URL.
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = strings.TrimRight(url, "/")
	}
}

// WithUserAgent sets custom User-Agent.
func WithUserAgent(ua string) Option {
	return func(c *Client) {
		c.userAgent = ua
	}
}

// WithLogger sets custom zap.Logger.
func WithLogger(logger *zap.Logger) Option {
	return func(c *Client) {
		c.logger = logger
	}
}

// NewClient creates a new Yahoo Finance client.
func NewClient(opts ...Option) *Client {
	c := &Client{
		httpClient: &http.Client{Timeout: defaultTimeout},
		baseURL:    defaultBaseURL,
		userAgent:  defaultUserAgent,
		logger:     zap.NewNop(),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// NormalizeSymbol ensures correct Yahoo Finance symbol notation for IDX equities & FX.
func NormalizeSymbol(ticker string) string {
	t := strings.TrimSpace(ticker)
	if t == "" {
		return ""
	}
	tUpper := strings.ToUpper(t)
	if tUpper == "USDIDR" || tUpper == "USD/IDR" {
		return "USDIDR=X"
	}
	if strings.Contains(tUpper, "=") || strings.HasPrefix(tUpper, "^") {
		return tUpper
	}
	if strings.HasSuffix(tUpper, ".JK") {
		return tUpper
	}
	if strings.Contains(tUpper, ".") {
		return tUpper
	}
	return tUpper + ".JK"
}

// CleanTicker strips exchange suffix for normalized entity storage.
func CleanTicker(symbol string) string {
	s := strings.TrimSpace(symbol)
	sUpper := strings.ToUpper(s)
	if strings.HasSuffix(sUpper, ".JK") {
		return strings.TrimSuffix(sUpper, ".JK")
	}
	return sUpper
}

type yfChartResponse struct {
	Chart struct {
		Result []struct {
			Meta struct {
				Currency           string  `json:"currency"`
				Symbol             string  `json:"symbol"`
				ExchangeName       string  `json:"exchangeName"`
				InstrumentType     string  `json:"instrumentType"`
				RegularMarketPrice float64 `json:"regularMarketPrice"`
				ChartPreviousClose float64 `json:"chartPreviousClose"`
			} `json:"meta"`
			Timestamp  []int64 `json:"timestamp"`
			Indicators struct {
				Quote []struct {
					Open   []*float64 `json:"open"`
					High   []*float64 `json:"high"`
					Low    []*float64 `json:"low"`
					Close  []*float64 `json:"close"`
					Volume []*float64 `json:"volume"`
				} `json:"quote"`
				Adjclose []struct {
					Adjclose []*float64 `json:"adjclose"`
				} `json:"adjclose"`
			} `json:"indicators"`
		} `json:"result"`
		Error *struct {
			Code        string `json:"code"`
			Description string `json:"description"`
		} `json:"error"`
	} `json:"chart"`
}

// FetchHistoricalPrices fetches historical daily candles for a ticker over a given range period.
// Supported ranges: "1d", "5d", "1mo", "3mo", "6mo", "1y", "2y", "5y", "10y", "ytd", "max". Default: "5y".
func (c *Client) FetchHistoricalPrices(ticker string, rangePeriod string) ([]PriceCandle, error) {
	return c.FetchHistoricalPricesWithContext(context.Background(), ticker, rangePeriod)
}

// FetchHistoricalPricesWithContext fetches historical daily candles with a provided context.
func (c *Client) FetchHistoricalPricesWithContext(ctx context.Context, ticker string, rangePeriod string) ([]PriceCandle, error) {
	if rangePeriod == "" {
		rangePeriod = "5y"
	}
	symbol := NormalizeSymbol(ticker)
	cleanTicker := CleanTicker(ticker)

	reqURL := fmt.Sprintf("%s/v8/finance/chart/%s?range=%s&interval=1d", c.baseURL, symbol, rangePeriod)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request for %s: %w", symbol, err)
	}

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching price chart for %s: %w", symbol, err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body for %s: %w", symbol, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("yahoo finance api error (%d) for %s: %s", resp.StatusCode, symbol, string(bodyBytes))
	}

	var yfResp yfChartResponse
	if err := json.Unmarshal(bodyBytes, &yfResp); err != nil {
		return nil, fmt.Errorf("decoding yahoo chart json for %s: %w", symbol, err)
	}

	if yfResp.Chart.Error != nil {
		return nil, fmt.Errorf("yahoo finance error for %s: %s - %s", symbol, yfResp.Chart.Error.Code, yfResp.Chart.Error.Description)
	}

	if len(yfResp.Chart.Result) == 0 {
		return nil, fmt.Errorf("no chart result returned for %s", symbol)
	}

	res := yfResp.Chart.Result[0]
	timestamps := res.Timestamp
	if len(timestamps) == 0 {
		return []PriceCandle{}, nil
	}

	var quotes struct {
		Open   []*float64
		High   []*float64
		Low    []*float64
		Close  []*float64
		Volume []*float64
	}
	if len(res.Indicators.Quote) > 0 {
		quotes.Open = res.Indicators.Quote[0].Open
		quotes.High = res.Indicators.Quote[0].High
		quotes.Low = res.Indicators.Quote[0].Low
		quotes.Close = res.Indicators.Quote[0].Close
		quotes.Volume = res.Indicators.Quote[0].Volume
	}

	var adjCloses []*float64
	if len(res.Indicators.Adjclose) > 0 {
		adjCloses = res.Indicators.Adjclose[0].Adjclose
	}

	candles := make([]PriceCandle, 0, len(timestamps))
	now := time.Now().UTC()

	for i, ts := range timestamps {
		if ts <= 0 {
			continue
		}

		var open, high, low, closeVal float64
		if i < len(quotes.Open) && quotes.Open[i] != nil {
			open = *quotes.Open[i]
		}
		if i < len(quotes.High) && quotes.High[i] != nil {
			high = *quotes.High[i]
		}
		if i < len(quotes.Low) && quotes.Low[i] != nil {
			low = *quotes.Low[i]
		}
		if i < len(quotes.Close) && quotes.Close[i] != nil {
			closeVal = *quotes.Close[i]
		}

		// Skip completely null/empty quote intervals
		var volRaw float64
		if i < len(quotes.Volume) && quotes.Volume[i] != nil {
			volRaw = *quotes.Volume[i]
		}
		if open == 0 && high == 0 && low == 0 && closeVal == 0 && volRaw == 0 {
			continue
		}

		// If close is nil/zero (common for the unfinalized current session candle), use regular market price or open
		if closeVal == 0 {
			if i == len(timestamps)-1 && res.Meta.RegularMarketPrice > 0 {
				closeVal = res.Meta.RegularMarketPrice
			} else if open > 0 {
				closeVal = open
			}
		}

		// Skip rows where all price fields are completely zero
		if closeVal == 0 && open == 0 && high == 0 && low == 0 {
			continue
		}

		adjClose := closeVal
		if i < len(adjCloses) && adjCloses[i] != nil && *adjCloses[i] > 0 {
			adjClose = *adjCloses[i]
		} else if adjClose == 0 {
			adjClose = closeVal
		}

		var volume int64
		if i < len(quotes.Volume) && quotes.Volume[i] != nil {
			volume = int64(*quotes.Volume[i])
		}

		candleDate := time.Unix(ts, 0).UTC()
		candles = append(candles, PriceCandle{
			Ticker:    cleanTicker,
			Date:      candleDate,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     closeVal,
			AdjClose:  adjClose,
			Volume:    volume,
			CreatedAt: now,
			UpdatedAt: now,
		})
	}

	return candles, nil
}

// FetchLatestPrice returns the latest regular market price or latest closing price.
func (c *Client) FetchLatestPrice(ctx context.Context, ticker string) (float64, error) {
	candles, err := c.FetchHistoricalPricesWithContext(ctx, ticker, "5d")
	if err != nil {
		return 0, err
	}
	if len(candles) == 0 {
		return 0, fmt.Errorf("no price data found for %s", ticker)
	}
	for i := len(candles) - 1; i >= 0; i-- {
		if candles[i].Close > 0 {
			return candles[i].Close, nil
		}
	}
	return 0, fmt.Errorf("no positive close price found for %s", ticker)
}

// FetchUSDIDR fetches the latest USD to IDR exchange rate from Yahoo Finance.
func (c *Client) FetchUSDIDR(ctx context.Context) (float64, error) {
	return c.FetchLatestPrice(ctx, "USDIDR=X")
}

// DefaultClient is the package-level default Yahoo client.
var defaultClient = NewClient()

// FetchHistoricalPrices is a package-level helper using the default client.
func FetchHistoricalPrices(ticker string, rangePeriod string) ([]PriceCandle, error) {
	return defaultClient.FetchHistoricalPrices(ticker, rangePeriod)
}
