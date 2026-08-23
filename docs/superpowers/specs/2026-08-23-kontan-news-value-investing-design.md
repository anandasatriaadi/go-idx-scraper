# Design Specification: Kontan Multi-Channel Scraper & Value Investing Analysis Engine

**Date**: 2026-08-23  
**Status**: Approved  

---

## 1. Overview & Goals

This feature enhances the news scraping and intelligence pipeline in `go-idx-scraper` by:
1. **Multi-Channel Scraping**: Expanding Kontan web scraping to automatically cover both `investasi` and `keuangan` channels.
2. **Value Investing Scoring & Analysis**: Upgrading the OpenRouter Gemini summarization service to act as an objective Personal Investment Manager. It evaluates news articles from a strict Value Investing perspective (moat, intrinsic value, earnings quality, capital allocation, governance) and outputs a structured `-10` to `+10` `value_score`, an `impact_direction` (`Bullish`/`Bearish`/`Neutral`), and a concise `investment_takeaway`.
3. **End-to-End Type Synchronization**: Updating Go domain entities (`internal/feature/news`) and Nuxt 4 API TypeScript definitions (`idx-web/src/server/utils/types.ts`) so API consumers receive the new financial metrics.

---

## 2. Architecture & Data Flow

```
[Kontan Web]
   │
   ├─► kanal=investasi ──┐
   └─► kanal=keuangan  ──┴─► [Selenium Scraper] (Deduplicates by article URL)
                                 │
                                 ▼
                     [Markdown HTML Processor] (Cleaned Content)
                                 │
                                 ▼
                     [MongoDB persistence] (Initial News Record)
                                 │
                                 ▼
                 [OpenRouter Gemini Analysis Engine]
                     - Role: Personal Investment Manager
                     - Framework: Value Investing
                     - Structured JSON Schema Output
                                 │
                                 ▼
                     [MongoDB Document Update]
                         ├── value_score (-10 to +10)
                         ├── impact_direction (Bullish/Bearish/Neutral)
                         └── investment_takeaway
                                 │
                                 ▼
                    [idx-web Nuxt API Server]
```

---

## 3. Scraper Enhancements (`internal/infra/scraper/kontan/scraper.go`)

* **Channel Execution**:
  * Hardcode channel list: `channels := []string{"investasi", "keuangan"}`.
  * For each target date, `Scrape()` loops through both `investasi` and `keuangan` index search URLs:
    `https://www.kontan.co.id/search/indeks?kanal=<CHANNEL>&tanggal=DD&bulan=MM&tahun=YYYY&pos=indeks&per_page=N`
* **Deduplication**:
  * Maintain an in-memory `seenLinks := make(map[string]bool)` per `Scrape` run to ensure an article appearing in both channels is fetched and stored only once.

---

## 4. Domain Model Update (`internal/feature/news/entity.go`)

Update `News` struct:

```go
type News struct {
	ID                 bson.ObjectID `bson:"_id,omitempty" json:"id"`
	CreatedAt          time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt          time.Time     `bson:"updated_at" json:"updated_at"`
	Date               time.Time     `bson:"date" json:"date"`
	Title              string        `bson:"title" json:"title"`
	Summary            string        `bson:"summary" json:"summary"`
	Content            string        `bson:"content" json:"content"`
	Link               string        `bson:"link" json:"link"`
	Priority           int           `bson:"priority" json:"priority"`
	ValueScore         int           `bson:"value_score" json:"value_score"`
	ImpactDirection    string        `bson:"impact_direction" json:"impact_direction"`
	InvestmentTakeaway string        `bson:"investment_takeaway" json:"investment_takeaway"`
}
```

---

## 5. Value Investing AI Analysis Engine (`internal/feature/news/service.go`)

### 5.1 Response Structure & JSON Schema
Update `NewsSummary` struct used for OpenRouter JSON Schema generation:

```go
type NewsSummary struct {
	Title              string `json:"title"`
	Summary            string `json:"summary"`
	Priority           int    `json:"priority"`
	ValueScore         int    `json:"value_score"`
	ImpactDirection    string `json:"impact_direction"`
	InvestmentTakeaway string `json:"investment_takeaway"`
}
```

### 5.2 System Prompt Specification
Prompt persona: **Objective Personal Investment Manager applying Value Investing principles**.

Prompt requirements:
1. **Title**: Concise, objective, market-relevant title.
2. **Summary**: Exactly 3 sentences highlighting core financial facts, figures, and immediate implications.
3. **Priority**: `1` (highest priority market event) to `10` (routine / low market impact).
4. **ValueScore**: Integer strictly between `-10` and `+10`:
   * `-10` to `-1`: Negative impact on intrinsic value, competitive moat, balance sheet, or governance.
   * `0`: Neutral, macro noise, or no fundamental impact on business valuation.
   * `+1` to `+10`: Positive impact on long-term cash flows, earnings power, moat, or capital allocation.
5. **ImpactDirection**: Exactly `"Bullish"`, `"Bearish"`, or `"Neutral"`.
6. **InvestmentTakeaway**: 1-2 sentence objective advice on why this news matters (or doesn't) to a long-term value investor.

### 5.3 OpenRouter Request & Model
* Model: `google/gemini-2.5-flash` (or as configured in `OpenrouterApiKey`).
* ResponseFormat: `ChatCompletionResponseFormatTypeJSONSchema` with `Strict: true`.

---

## 6. Nuxt 4 API Integration (`idx-web/src/server/utils/types.ts`)

Update `News` interface in Nuxt API server:

```typescript
export interface News {
  _id?: string;
  id: string;
  created_at?: string;
  updated_at?: string;
  date?: string;
  title?: string;
  summary?: string;
  content?: string;
  link?: string;
  priority?: number;
  value_score?: number;
  impact_direction?: 'Bullish' | 'Bearish' | 'Neutral';
  investment_takeaway?: string;
}
```

---

## 7. Testing Strategy

1. **Scraper Unit Tests (`internal/infra/scraper/kontan/scraper_test.go`)**:
   * Verify `Scrape()` queries both `investasi` and `keuangan` channels using `MockBrowser`.
   * Verify duplicate links across channels are skipped.
2. **News Service Unit Tests (`internal/feature/news/service_test.go`)**:
   * Test `NewsSummary` JSON schema unmarshaling with `value_score`, `impact_direction`, and `investment_takeaway`.
3. **Regression Tests**:
   * Run `go test ./...` across all packages.
