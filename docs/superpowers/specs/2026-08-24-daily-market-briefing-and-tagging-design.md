# Design Specification: Daily Market Briefing & Ticker/Industry Intelligence Engine

**Date**: 2026-08-24  
**Status**: Approved  

---

## 1. Overview & Goals

This feature enhances the news processing pipeline in `go-idx-scraper` with automated market intelligence:
1. **7 AM GMT+8 Date Window**: Automatically calculates yesterday and today's date window in GMT+8 (`Asia/Makassar` / `UTC+8`) and skips articles already persisted in MongoDB.
2. **Ticker & Industry Categorization**: OpenRouter Gemini 3.7 Flash extracts IDX stock tickers (`tickers: []string`, e.g. `["BBRI", "BBCA"]`) and sector classifications (`industry: string`, `is_industry_wide: bool`) for granular search and filtering.
3. **Daily Morning Briefing ("Today's Summarization of Yesterday")**: Generates a comprehensive Value Investor Daily Briefing synthesizing the last 24h of news into actionable insights:
   - **Executive Macro & Market Pulse**
   - **🟢 Stocks to Watch (Buy / Bullish Lookout)**
   - **🔴 Stocks to Avoid / Risk Alerts (Bearish Lookout)**
   - **🏭 Sector & Industry Highlights**
   - **🎯 Value Manager Action Plan for Today**
4. **Dual Delivery**: Persists the daily briefing to MongoDB (`daily_briefings` collection) for Nuxt 4 API exposure and sends formatted HTML/Markdown emails to the configured mailing list.

---

## 2. Architecture & Data Flow

```
┌────────────────────────────────────────────────────────┐
│  7 AM GMT+8 Ingestion (Yesterday & Today Window)       │
└──────────────────────────┬─────────────────────────────┘
                           │
                           ▼
┌────────────────────────────────────────────────────────┐
│  Kontan Scraper (`investasi` + `keuangan`)             │
│  - Checks MongoDB `link` to skip already ingested news │
└──────────────────────────┬─────────────────────────────┘
                           │
                           ▼
┌────────────────────────────────────────────────────────┐
│  Per-Article Value Investing Analysis (Gemini 3.7)     │
│  - Title & 3-sentence Summary                          │
│  - ValueScore (-10 to +10), ImpactDirection, Takeaway  │
│  - Tickers (`["BBRI", "BBCA"]`)                        │
│  - Industry (e.g. "Banking", "Poultry") & Sector-Wide  │
└──────────────────────────┬─────────────────────────────┘
                           │
                           ▼
┌────────────────────────────────────────────────────────┐
│  Daily Briefing Synthesis Engine (Gemini 3.7 Flash)    │
│  - Aggregates all 24h news articles                    │
│  - Categorizes bullish picks, bearish alerts, sectors  │
│  - Synthesizes actionable Value Investor Plan          │
└──────────────────────────┬─────────────────────────────┘
                           │
              ┌────────────┴────────────┐
              ▼                         ▼
   ┌──────────────────────┐  ┌──────────────────────┐
   │ MongoDB Persistence  │  │ Email Delivery       │
   │ (`daily_briefings`)  │  │ (Formatted HTML/MD   │
   │ & Nuxt API Server    │  │  via helper.SendMail)│
   └──────────────────────┘  └──────────────────────┘
```

---

## 3. Domain Model Specifications (`internal/feature/news`)

### 3.1 Updated `News` Entity
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
	Tickers            []string      `bson:"tickers" json:"tickers"`
	Industry           string        `bson:"industry" json:"industry"`
	IsIndustryWide     bool          `bson:"is_industry_wide" json:"is_industry_wide"`
}
```

### 3.2 New `Briefing` Entity
```go
type BriefingItem struct {
	Ticker             string `bson:"ticker,omitempty" json:"ticker,omitempty"`
	IssuerName         string `bson:"issuer_name,omitempty" json:"issuer_name,omitempty"`
	Headline           string `bson:"headline" json:"headline"`
	Rationale          string `bson:"rationale" json:"rationale"`
	ValueScore         int    `bson:"value_score" json:"value_score"`
	InvestmentTakeaway string `bson:"investment_takeaway" json:"investment_takeaway"`
}

type SectorHighlight struct {
	Sector    string `bson:"sector" json:"sector"`
	Summary   string `bson:"summary" json:"summary"`
	Sentiment string `bson:"sentiment" json:"sentiment"` // "Bullish", "Bearish", "Neutral"
}

type Briefing struct {
	ID               bson.ObjectID     `bson:"_id,omitempty" json:"id"`
	Date             time.Time         `bson:"date" json:"date"`
	Title            string            `bson:"title" json:"title"`
	MacroPulse       string            `bson:"macro_pulse" json:"macro_pulse"`
	BullishLookout   []BriefingItem    `bson:"bullish_lookout" json:"bullish_lookout"`
	BearishLookout   []BriefingItem    `bson:"bearish_lookout" json:"bearish_lookout"`
	SectorHighlights []SectorHighlight `bson:"sector_highlights" json:"sector_highlights"`
	ActionPlan       string            `bson:"action_plan" json:"action_plan"`
	RawMarkdown      string            `bson:"raw_markdown" json:"raw_markdown"`
	CreatedAt        time.Time         `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time         `bson:"updated_at" json:"updated_at"`
}

type BriefingRepository interface {
	Create(ctx context.Context, b *Briefing) error
	FindByDate(ctx context.Context, date time.Time) (*Briefing, error)
	FindLatest(ctx context.Context) (*Briefing, error)
	FindAll(ctx context.Context, filter any, opts ...options.Lister[options.FindOptions]) ([]*Briefing, error)
}
```

---

## 4. Per-Article Extraction & JSON Schema (`internal/feature/news/service.go`)

### 4.1 Structured OpenRouter JSON Schema
Update `NewsSummary` struct with JSON Schema descriptions:
```go
type NewsSummary struct {
	Title              string   `json:"title" jsonschema:"description=An updated engaging and objective headline"`
	Summary            string   `json:"summary" jsonschema:"description=Concise 3-sentence summary covering core facts and operational impact"`
	Priority           int      `json:"priority" jsonschema:"description=Market impact urgency from 1 (critical) to 10 (routine)"`
	ValueScore         int      `json:"value_score" jsonschema:"description=Value investing impact score strictly between -10 and +10 (-10 is severe impairment, 0 is neutral, +10 is high value creation)"`
	ImpactDirection    string   `json:"impact_direction" jsonschema:"enum=Bullish,enum=Bearish,enum=Neutral,description=Direction of impact on intrinsic business value"`
	InvestmentTakeaway string   `json:"investment_takeaway" jsonschema:"description=1-2 sentence takeaway for a disciplined long-term value investor"`
	Tickers            []string `json:"tickers" jsonschema:"description=List of 4-letter IDX stock ticker symbols explicitly mentioned or directly affected, e.g. ['BBRI', 'BBCA']. Empty array if none."`
	Industry           string   `json:"industry" jsonschema:"description=Primary industry or sector (e.g. Banking, Poultry, Mining, Energy, Consumer Goods, Infrastructure, Macroeconomics)"`
	IsIndustryWide     bool     `json:"is_industry_wide" jsonschema:"description=True if the news affects an entire industry or sector rather than one specific company"`
}
```

---

## 5. Daily Briefing Synthesis Service (`internal/feature/news/briefing_service.go`)

### 5.1 Aggregation & Synthesis Process
1. Query all `News` records whose `date` or `created_at` falls in the target 24h window.
2. Formulate synthesis prompt to OpenRouter Gemini 3.7 Flash:
   - Evaluates all collected news collectively.
   - Highlights top high-conviction bullish candidates (Buy lookout).
   - Highlights top risk alerts and governance concerns (Bearish lookout).
   - Groups industry-wide developments by sector.
   - Formats a polished Markdown / HTML report.
3. Persists the resulting `Briefing` record in MongoDB `daily_briefings` collection.
4. Sends email using `helper.SendMail` to the configured `mailing.list`.

---

## 6. Scraper Ingestion Runner (`cmd/scraper/main.go`)

- **Default Timezone**: GMT+8 (`time.FixedZone("GMT+8", 8*3600)`).
- **Date Window Computation**:
  - `endDate`: Current date in GMT+8.
  - `startDate`: Yesterday in GMT+8.
- **Idempotency**:
  - Checks if `repo.ExistsByLink(ctx, link)` returns true before inserting.
  - If news already exists in MongoDB, skips scraping/inserting.

---

## 7. Nuxt 4 API Integration (`idx-web`)

1. **TypeScript Definitions (`idx-web/src/server/utils/types.ts`)**:
   - Update `News` interface with `tickers?: string[]`, `industry?: string`, `is_industry_wide?: boolean`.
   - Add `Briefing`, `BriefingItem`, and `SectorHighlight` interfaces.
2. **API Endpoints**:
   - `GET /api/v1/briefings/latest`: Returns the most recent daily market briefing.
   - `GET /api/v1/briefings`: List past briefings with pagination.
   - Support `ticker` filter in `GET /api/v1/news?ticker=BBRI`.

---

## 8. Testing Strategy

1. **Unit Tests (`internal/feature/news/service_test.go`)**:
   - Verify `NewsSummary` JSON schema unmarshaling for `tickers`, `industry`, `is_industry_wide`.
   - Verify `Briefing` serialization and schema validation.
2. **MongoDB Repository Tests (`internal/infra/db/mongo/briefing_repo_test.go`)**:
   - Test CRUD operations for `BriefingRepository`.
3. **End-to-End Build & Run Tests**:
   - Run `go test -count=1 ./...`.
   - Build binaries `go build -o bin/ ./cmd/...`.
