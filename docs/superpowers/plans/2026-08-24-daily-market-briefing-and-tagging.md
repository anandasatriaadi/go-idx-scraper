# Daily Market Briefing & Ticker/Industry Intelligence Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement automated 7 AM GMT+8 multi-channel ingestion with MongoDB link deduplication, structured Gemini 3.7 Flash extraction of stock tickers and industry sectors, and a Daily Morning Market Briefing generator ("Today's Summarization of Yesterday") that evaluates bullish stock opportunities and bearish risk alerts, storing them in MongoDB and delivering them via email.

**Architecture:** Domain-Driven Design (Hexagonal) inside `internal/feature/news`. The Scraper (`internal/infra/scraper/kontan`) ingests `investasi` and `keuangan` articles, checking `NewsRepository.ExistsByLink` to skip already processed articles. OpenRouter Gemini 3.7 Flash extracts per-article value investing scores, tickers, and industry metadata. A Daily Briefing engine aggregates 24h news into structured Macro, Bullish Watchlist, Bearish Alerts, and Sector Highlights, persisting to MongoDB (`daily_briefings`) and emailing subscribers via `helper.SendMail`.

**Tech Stack:** Go 1.24+, MongoDB v2 Driver (`go.mongodb.org/mongo-driver/v2`), OpenRouter SDK (`github.com/revrost/go-openrouter`), Selenium WebDriver, Goquery, Gomail v2, TypeScript / Nuxt 4.

## Global Constraints

- Never use `context.Background()` in domain logic; propagate context from callers.
- Wrap errors with `%w` for error context across layer boundaries.
- Use `zap.Logger` structured logging.
- Target timezone for daily window: GMT+8 (`time.FixedZone("GMT+8", 8*3600)`).
- Value score scale: integer strictly between `-10` and `+10`.
- Tickers: array of uppercase 4-letter IDX codes (e.g. `["BBRI", "BBCA"]`).
- Impact direction: `"Bullish"`, `"Bearish"`, or `"Neutral"`.

---

### Task 1: Domain Entities & Repository Ports (Go & TypeScript)

**Files:**
- Modify: `internal/feature/news/entity.go`
- Modify: `idx-web/src/server/utils/types.ts`
- Test: `internal/feature/news/service_test.go`

**Interfaces:**
- Produces: `News` (with `Tickers`, `Industry`, `IsIndustryWide`), `Briefing`, `BriefingItem`, `SectorHighlight`, `BriefingRepository`.

- [ ] **Step 1: Write test for News & Briefing entity fields**

Add `TestBriefingEntity_Fields` in `internal/feature/news/service_test.go`:

```go
func TestBriefingEntity_Fields(t *testing.T) {
	id := bson.NewObjectID()
	now := time.Now()
	b := &Briefing{
		ID:         id,
		Date:       now,
		Title:      "Daily Market Briefing",
		MacroPulse: "Positive macro sentiment.",
		BullishLookout: []BriefingItem{
			{
				Ticker:             "BBRI",
				IssuerName:         "PT Bank Rakyat Indonesia Tbk",
				Headline:           "Expanding Net Interest Margin",
				Rationale:          "Strong micro loan disbursement.",
				ValueScore:         8,
				InvestmentTakeaway: "Attractive valuation and solid moat.",
			},
		},
		BearishLookout: []BriefingItem{
			{
				Ticker:             "ASBI",
				IssuerName:         "PT Asuransi Bintang Tbk",
				Headline:           "Embezzlement Investigation",
				Rationale:          "Internal controls failure.",
				ValueScore:         -7,
				InvestmentTakeaway: "Avoid until balance sheet risk is cleared.",
			},
		},
		SectorHighlights: []SectorHighlight{
			{
				Sector:    "Banking",
				Summary:   "Liquidity tightening persists.",
				Sentiment: "Neutral",
			},
		},
		ActionPlan: "Accumulate high ROE banks on weakness.",
	}

	if len(b.BullishLookout) != 1 || b.BullishLookout[0].Ticker != "BBRI" {
		t.Errorf("Expected BullishLookout with BBRI")
	}
	if len(b.BearishLookout) != 1 || b.BearishLookout[0].Ticker != "ASBI" {
		t.Errorf("Expected BearishLookout with ASBI")
	}
	if len(b.SectorHighlights) != 1 || b.SectorHighlights[0].Sector != "Banking" {
		t.Errorf("Expected SectorHighlights with Banking")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v ./internal/feature/news -run TestBriefingEntity_Fields`  
Expected: Compilation failure due to missing `Briefing` types.

- [ ] **Step 3: Update `internal/feature/news/entity.go` and `idx-web/src/server/utils/types.ts`**

Update `internal/feature/news/entity.go`:
```go
package news

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

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
	Sentiment string `bson:"sentiment" json:"sentiment"`
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

// Port: Repository defines news data persistence
type Repository interface {
	Create(ctx context.Context, news *News) error
	FindAll(ctx context.Context, filter any, opts ...options.Lister[options.FindOptions]) ([]*News, error)
	FindByID(ctx context.Context, id bson.ObjectID) (*News, error)
	UpdateByID(ctx context.Context, id bson.ObjectID, update any) error
	ExistsByLink(ctx context.Context, link string) (bool, error)
}

// Port: BriefingRepository defines briefing data persistence
type BriefingRepository interface {
	Create(ctx context.Context, b *Briefing) error
	FindByDate(ctx context.Context, date time.Time) (*Briefing, error)
	FindLatest(ctx context.Context) (*Briefing, error)
	FindAll(ctx context.Context, filter any, opts ...options.Lister[options.FindOptions]) ([]*Briefing, error)
}

// Port: Scraper defines news scraping interface
type Scraper interface {
	Scrape(ctx context.Context, startDate, endDate time.Time, onNewsFound func(*News) error) error
}
```

Update `idx-web/src/server/utils/types.ts`:
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
  tickers?: string[];
  industry?: string;
  is_industry_wide?: boolean;
}

export interface BriefingItem {
  ticker?: string;
  issuer_name?: string;
  headline: string;
  rationale: string;
  value_score: number;
  investment_takeaway: string;
}

export interface SectorHighlight {
  sector: string;
  summary: string;
  sentiment: 'Bullish' | 'Bearish' | 'Neutral';
}

export interface Briefing {
  _id?: string;
  id: string;
  date: string;
  title: string;
  macro_pulse: string;
  bullish_lookout: BriefingItem[];
  bearish_lookout: BriefingItem[];
  sector_highlights: SectorHighlight[];
  action_plan: string;
  raw_markdown?: string;
  created_at?: string;
  updated_at?: string;
}
```

- [ ] **Step 4: Update `MockRepository` in `internal/feature/news/service_test.go` and run test**

Add `ExistsByLink` to `MockRepository`:
```go
func (m *MockRepository) ExistsByLink(ctx context.Context, link string) (bool, error) {
	return false, m.Err
}
```
Run: `go test -v ./internal/feature/news -run TestBriefingEntity_Fields`  
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/feature/news/entity.go internal/feature/news/service_test.go idx-web/src/server/utils/types.ts
git commit -m "feat(news): add briefing entities, tickers/industry fields and repository ports"
```

---

### Task 2: MongoDB Repositories Implementation (News & Briefing)

**Files:**
- Modify: `internal/infra/db/mongo/news_repo.go`
- Create: `internal/infra/db/mongo/briefing_repo.go`
- Create: `internal/infra/db/mongo/briefing_repo_test.go`

**Interfaces:**
- Produces: `BriefingRepository` implementation, `NewsRepository.ExistsByLink`.

- [ ] **Step 1: Write test for `ExistsByLink` and `BriefingRepository`**

Create `internal/infra/db/mongo/briefing_repo_test.go`:
```go
package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/news"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type MockBriefingRepo struct {
	Briefing *news.Briefing
	List     []*news.Briefing
	Err      error
}

func (m *MockBriefingRepo) Create(ctx context.Context, b *news.Briefing) error {
	m.Briefing = b
	return m.Err
}
func (m *MockBriefingRepo) FindByDate(ctx context.Context, date time.Time) (*news.Briefing, error) {
	return m.Briefing, m.Err
}
func (m *MockBriefingRepo) FindLatest(ctx context.Context) (*news.Briefing, error) {
	return m.Briefing, m.Err
}
func (m *MockBriefingRepo) FindAll(ctx context.Context, filter any, opts ...any) ([]*news.Briefing, error) {
	return m.List, m.Err
}

func TestBriefingStruct(t *testing.T) {
	b := &news.Briefing{
		ID:    bson.NewObjectID(),
		Title: "Morning Briefing Test",
	}
	if b.Title != "Morning Briefing Test" {
		t.Errorf("Unexpected title: %s", b.Title)
	}
}
```

- [ ] **Step 2: Implement `ExistsByLink` in `internal/infra/db/mongo/news_repo.go`**

```go
// ExistsByLink checks if an article with the given link already exists in MongoDB
func (r *NewsRepository) ExistsByLink(ctx context.Context, link string) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"link": link}, options.Count().SetLimit(1))
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
```

- [ ] **Step 3: Create `internal/infra/db/mongo/briefing_repo.go`**

```go
package mongo

import (
	"context"
	"time"

	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/news"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type BriefingRepository struct {
	collection *mongo.Collection
}

func NewBriefingRepository(db *mongo.Database) news.BriefingRepository {
	return &BriefingRepository{
		collection: db.Collection("daily_briefings"),
	}
}

func (r *BriefingRepository) Create(ctx context.Context, model *news.Briefing) error {
	now := time.Now()
	if model.CreatedAt.IsZero() {
		model.CreatedAt = now
	}
	model.UpdatedAt = now
	_, err := r.collection.InsertOne(ctx, model)
	return err
}

func (r *BriefingRepository) FindByDate(ctx context.Context, date time.Time) (*news.Briefing, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.AddDate(0, 0, 1)

	filter := bson.M{
		"date": bson.M{
			"$gte": startOfDay,
			"$lt":  endOfDay,
		},
	}
	var result news.Briefing
	err := r.collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *BriefingRepository) FindLatest(ctx context.Context) (*news.Briefing, error) {
	opts := options.FindOne().SetSort(bson.D{{Key: "date", Value: -1}})
	var result news.Briefing
	err := r.collection.FindOne(ctx, bson.M{}, opts).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *BriefingRepository) FindAll(ctx context.Context, filter any, opts ...options.Lister[options.FindOptions]) ([]*news.Briefing, error) {
	cursor, err := r.collection.Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*news.Briefing
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}
```

- [ ] **Step 4: Run tests to verify compilation and execution**

Run: `go test -v ./internal/infra/db/mongo/...`  
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/infra/db/mongo/news_repo.go internal/infra/db/mongo/briefing_repo.go internal/infra/db/mongo/briefing_repo_test.go
git commit -m "feat(mongo): implement BriefingRepository and ExistsByLink on NewsRepository"
```

---

### Task 3: Per-Article Ticker & Industry Extraction (Gemini 3.7 Flash)

**Files:**
- Modify: `internal/feature/news/service.go`
- Modify: `internal/feature/news/service_test.go`

**Interfaces:**
- Produces: `NewsSummary` with `tickers`, `industry`, `is_industry_wide` fields, populating `News` document in MongoDB.

- [ ] **Step 1: Write failing test for `NewsSummary` schema containing tickers and industry**

Add test in `internal/feature/news/service_test.go`:

```go
func TestNewsSummary_TickersAndIndustry(t *testing.T) {
	schema, err := jsonschema.GenerateSchemaForType(NewsSummary{})
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}
	if schema == nil {
		t.Fatal("Expected non-nil schema")
	}

	jsonSample := `{
		"title": "Kinerja Emiten Unggas Menguat",
		"summary": "Emiten peternakan ayam mencatat pemulihan margin di semester II. Permintaan stabil menopang pertumbuhan laba. Efisiensi pakan menjadi faktor pendukung utama.",
		"priority": 4,
		"value_score": 6,
		"impact_direction": "Bullish",
		"investment_takeaway": "CPIN dan JPFA memiliki posisi pasar dominan dan efisiensi rantai pasok terintegrasi.",
		"tickers": ["CPIN", "JPFA", "MAIN"],
		"industry": "Poultry",
		"is_industry_wide": true
	}`

	var summary NewsSummary
	if err := json.Unmarshal([]byte(jsonSample), &summary); err != nil {
		t.Fatalf("Failed to unmarshal NewsSummary: %v", err)
	}

	if len(summary.Tickers) != 3 || summary.Tickers[0] != "CPIN" {
		t.Errorf("Expected tickers with CPIN, got %v", summary.Tickers)
	}
	if summary.Industry != "Poultry" {
		t.Errorf("Expected industry 'Poultry', got '%s'", summary.Industry)
	}
	if !summary.IsIndustryWide {
		t.Errorf("Expected is_industry_wide to be true")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v ./internal/feature/news -run TestNewsSummary_TickersAndIndustry`  
Expected: FAIL (`summary.Tickers`, `summary.Industry`, `summary.IsIndustryWide` undefined).

- [ ] **Step 3: Update `NewsSummary` and Prompt in `internal/feature/news/service.go`**

Update `NewsSummary`:
```go
type NewsSummary struct {
	Title              string   `json:"title" jsonschema:"description=An updated engaging and objective title capturing the essence of the article"`
	Summary            string   `json:"summary" jsonschema:"description=Concise 3-sentence summary highlighting financial facts figures and immediate market implications"`
	Priority           int      `json:"priority" jsonschema:"description=Market impact priority from 1 (highest market urgency/impact) to 10 (routine/low market impact)"`
	ValueScore         int      `json:"value_score" jsonschema:"description=Fundamental value investing impact score strictly between -10 and +10 (-10 is severe fundamental impairment, 0 is neutral/macro noise, +10 is massive fundamental value creation)"`
	ImpactDirection    string   `json:"impact_direction" jsonschema:"enum=Bullish,enum=Bearish,enum=Neutral,description=Directional impact on underlying business intrinsic value"`
	InvestmentTakeaway string   `json:"investment_takeaway" jsonschema:"description=1-2 sentence actionable takeaway strictly from the perspective of a disciplined long-term value investor focusing on moat, capital allocation, cash flows, and intrinsic value"`
	Tickers            []string `json:"tickers" jsonschema:"description=List of 4-letter IDX stock ticker symbols explicitly mentioned or directly affected (e.g. ['BBRI', 'BBCA']). Empty array if none."`
	Industry           string   `json:"industry" jsonschema:"description=Primary industry or sector classification (e.g. Banking, Poultry, Mining, Energy, Consumer Goods, Infrastructure, Technology, Macroeconomics)"`
	IsIndustryWide     bool     `json:"is_industry_wide" jsonschema:"description=True if the news affects an entire industry sector or macroeconomic policy rather than just one individual company"`
}
```

Update Prompt:
```go
prompt := fmt.Sprintf(`You are an expert Personal Investment Manager and seasoned Value Investor adhering strictly to the fundamental principles of Benjamin Graham, Warren Buffett, and Charlie Munger.

Analyze the provided financial news article with super objective, disciplined rigor. Evaluate whether this event impacts intrinsic business value, economic moats, capital allocation, balance sheet durability, or free cash flow generation.

Provide your evaluation adhering to the following rules:
- Title: Clear, concise, and professional headline.
- Summary: Exactly 3 sentences. Cover core facts, financial metrics, and operational impact.
- Priority: Integer from 1 to 10 (1 = critical high-impact market event, 10 = routine noise).
- ValueScore: Integer from -10 to +10:
  * -10 to -1: Fundamental destruction (deteriorating moat, dilutive acquisitions, high debt risk, governance red flags).
  * 0: Neutral / Macro noise / Speculative price movements with no underlying business value change.
  * +1 to +10: Fundamental enhancement (widening moat, high ROIC reinvestment, robust organic growth, disciplined capital allocation).
- ImpactDirection: Exactly "Bullish", "Bearish", or "Neutral".
- InvestmentTakeaway: 1 to 2 sentences summarizing the bottom-line takeaway for a long-term value investor.
- Tickers: Array of uppercase 4-letter Indonesian stock tickers explicitly mentioned or impacted (e.g. ["BBRI", "BBCA"]). Empty array [] if no specific company is mentioned.
- Industry: Sector category (e.g., "Banking", "Poultry", "Mining", "Energy", "Consumer Goods", "Technology", "Infrastructure", "Macroeconomics").
- IsIndustryWide: Boolean true if the news affects the whole sector or macro economy rather than an isolated company.

Article:
"""
%s
"""`, n.Content)
```

Update MongoDB update in `Summarize`:
```go
update := bson.M{
	"$set": bson.M{
		"title":               summary.Title,
		"summary":             summary.Summary,
		"priority":            summary.Priority,
		"value_score":         summary.ValueScore,
		"impact_direction":    summary.ImpactDirection,
		"investment_takeaway": summary.InvestmentTakeaway,
		"tickers":             summary.Tickers,
		"industry":            summary.Industry,
		"is_industry_wide":    summary.IsIndustryWide,
		"updated_at":          time.Now(),
	},
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test -v ./internal/feature/news/...`  
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/feature/news/service.go internal/feature/news/service_test.go
git commit -m "feat(news): extract tickers and industry classification in per-article AI summary"
```

---

### Task 4: Daily Briefing Generator & Email Delivery Service

**Files:**
- Modify: `internal/feature/news/service.go`
- Modify: `internal/feature/news/service_test.go`

**Interfaces:**
- Produces: `Service.GenerateDailyBriefing(ctx context.Context, targetDate time.Time, briefingRepo BriefingRepository) (*Briefing, error)`

- [ ] **Step 1: Write unit test for Daily Briefing Schema & formatting**

Add `TestDailyBriefing_SchemaGeneration` in `internal/feature/news/service_test.go`:

```go
func TestDailyBriefing_SchemaGeneration(t *testing.T) {
	schema, err := jsonschema.GenerateSchemaForType(BriefingSchemaOutput{})
	if err != nil {
		t.Fatalf("Failed to generate BriefingSchemaOutput schema: %v", err)
	}
	if schema == nil {
		t.Fatal("Expected non-nil schema")
	}

	jsonSample := `{
		"title": "Morning Market Intelligence Briefing - 24 August 2026",
		"macro_pulse": "IHSG consolidates as foreign flows stabilize and commodity prices rebound.",
		"bullish_lookout": [
			{
				"ticker": "BBRI",
				"issuer_name": "PT Bank Rakyat Indonesia Tbk",
				"headline": "Micro Lending Expansion",
				"rationale": "High NIM and robust capital adequacy ratio.",
				"value_score": 8,
				"investment_takeaway": "Long-term compounding opportunity."
			}
		],
		"bearish_lookout": [
			{
				"ticker": "ASBI",
				"issuer_name": "PT Asuransi Bintang Tbk",
				"headline": "Embezzlement Scandal",
				"rationale": "Governance breakdown.",
				"value_score": -7,
				"investment_takeaway": "High governance risk; avoid."
			}
		],
		"sector_highlights": [
			{
				"sector": "Banking",
				"summary": "Liquidity remains tight but top tier banks maintain pricing power.",
				"sentiment": "Neutral"
			}
		],
		"action_plan": "Focus on high-quality big cap banks and poultry leaders with widening moats."
	}`

	var output BriefingSchemaOutput
	if err := json.Unmarshal([]byte(jsonSample), &output); err != nil {
		t.Fatalf("Failed to unmarshal BriefingSchemaOutput: %v", err)
	}

	if output.Title == "" || len(output.BullishLookout) != 1 {
		t.Errorf("Unexpected unmarshaled briefing: %+v", output)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v ./internal/feature/news -run TestDailyBriefing_SchemaGeneration`  
Expected: FAIL (`BriefingSchemaOutput` undefined).

- [ ] **Step 3: Implement `BriefingSchemaOutput` and `GenerateDailyBriefing` in `internal/feature/news/service.go`**

Define `BriefingSchemaOutput`:
```go
type BriefingSchemaOutput struct {
	Title            string            `json:"title" jsonschema:"description=Engaging title for today's market briefing"`
	MacroPulse       string            `json:"macro_pulse" jsonschema:"description=2-3 sentence summary of broader market sentiment and macro developments"`
	BullishLookout   []BriefingItem    `json:"bullish_lookout" jsonschema:"description=Top high-conviction companies with positive fundamental catalysts, expanding moats, or earnings growth"`
	BearishLookout   []BriefingItem    `json:"bearish_lookout" jsonschema:"description=Companies facing serious fundamental risks, governance issues, or severe headwind alerts"`
	SectorHighlights []SectorHighlight `json:"sector_highlights" jsonschema:"description=Industry-wide and sector developments grouped by industry"`
	ActionPlan       string            `json:"action_plan" jsonschema:"description=Disciplined, concrete 2-3 sentence action plan for long-term value investors entering today's session"`
}
```

Implement `GenerateDailyBriefing`:
```go
func (s *Service) GenerateDailyBriefing(ctx context.Context, targetDate time.Time, bRepo BriefingRepository) (*Briefing, error) {
	s.logger.Info("Starting Daily Briefing generation", zap.Time("date", targetDate))
	if s.cfg == nil {
		return nil, fmt.Errorf("config not loaded")
	}

	// Calculate 24h window (yesterday 00:00 to targetDate 23:59:59)
	startWindow := targetDate.AddDate(0, 0, -1)
	startOfWindow := time.Date(startWindow.Year(), startWindow.Month(), startWindow.Day(), 0, 0, 0, 0, targetDate.Location())
	endOfWindow := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 23, 59, 59, 999999999, targetDate.Location())

	filter := bson.M{
		"created_at": bson.M{
			"$gte": startOfWindow,
			"$lte": endOfWindow,
		},
	}
	newsList, err := s.repo.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("fetching 24h news: %w", err)
	}

	if len(newsList) == 0 {
		s.logger.Warn("No news found in 24h window for briefing", zap.Time("start", startOfWindow), zap.Time("end", endOfWindow))
		return nil, nil
	}

	s.logger.Info("Aggregating news for briefing", zap.Int("count", len(newsList)))

	// Format news summaries for LLM prompt
	var sb strings.Builder
	for i, n := range newsList {
		sb.WriteString(fmt.Sprintf("%d. [%s] (Tickers: %v, Score: %+d, Direction: %s, Sector: %s)\nTitle: %s\nSummary: %s\nTakeaway: %s\n\n",
			i+1, n.Date.Format("2006-01-02"), n.Tickers, n.ValueScore, n.ImpactDirection, n.Industry, n.Title, n.Summary, n.InvestmentTakeaway))
	}

	client := openrouter.NewClient(s.cfg.OpenrouterApiKey)
	schema, err := jsonschema.GenerateSchemaForType(BriefingSchemaOutput{})
	if err != nil {
		return nil, fmt.Errorf("GenerateSchemaForType for briefing: %w", err)
	}

	prompt := fmt.Sprintf(`You are an elite Personal Investment Manager preparing the Daily Morning Market Intelligence Briefing for a disciplined Value Investor (Graham-Buffett-Munger school).

Based on all the news collected over the past 24 hours below, synthesize an authoritative, super-objective Daily Briefing answering:
1. Executive Macro & Market Pulse.
2. Stocks to Watch / Buy Lookout (Companies with strong moats, positive catalysts, high ROIC potential, or attractive fundamental developments).
3. Stocks to Avoid / Risk Lookout (Companies with governance red flags, balance sheet damage, or regulatory headwinds).
4. Sector & Industry Highlights (Macro or industry-wide trends e.g. Banking, Poultry, Mining, Energy).
5. Value Investor Action Plan for today's market session.

24-Hour News Digest:
"""
%s
"""`, sb.String())

	request := openrouter.ChatCompletionRequest{
		Model: "google/gemini-3.7-flash",
		Messages: []openrouter.ChatCompletionMessage{
			{
				Role:    openrouter.ChatMessageRoleUser,
				Content: openrouter.Content{Text: prompt},
			},
		},
		ResponseFormat: &openrouter.ChatCompletionResponseFormat{
			Type: openrouter.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: &openrouter.ChatCompletionResponseFormatJSONSchema{
				Name:   "daily_market_briefing",
				Schema: schema,
				Strict: true,
			},
		},
	}

	res, err := client.CreateChatCompletion(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("openrouter briefing completion: %w", err)
	}

	if len(res.Choices) == 0 {
		return nil, fmt.Errorf("no response choices from openrouter")
	}

	var output BriefingSchemaOutput
	if err := json.Unmarshal([]byte(res.Choices[0].Message.Content.Text), &output); err != nil {
		return nil, fmt.Errorf("unmarshaling briefing output: %w", err)
	}

	// Format Raw Markdown for email/rendering
	rawMarkdown := formatBriefingMarkdown(output, targetDate)

	briefing := &Briefing{
		ID:               bson.NewObjectID(),
		Date:             targetDate,
		Title:            output.Title,
		MacroPulse:       output.MacroPulse,
		BullishLookout:   output.BullishLookout,
		BearishLookout:   output.BearishLookout,
		SectorHighlights: output.SectorHighlights,
		ActionPlan:       output.ActionPlan,
		RawMarkdown:      rawMarkdown,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if bRepo != nil {
		if err := bRepo.Create(ctx, briefing); err != nil {
			s.logger.Error("Failed to persist briefing in MongoDB", zap.Error(err))
		} else {
			s.logger.Info("Successfully saved Daily Briefing in MongoDB", zap.String("id", briefing.ID.Hex()))
		}
	}

	return briefing, nil
}

func formatBriefingMarkdown(b BriefingSchemaOutput, d time.Time) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n\n", b.Title))
	sb.WriteString(fmt.Sprintf("**Date:** %s (7:00 AM GMT+8)\n\n", d.Format("Monday, 02 January 2006")))
	sb.WriteString("## 🌐 Executive Macro & Market Pulse\n")
	sb.WriteString(b.MacroPulse + "\n\n")

	sb.WriteString("## 🟢 Stocks on Lookout (Buy / Opportunities)\n")
	if len(b.BullishLookout) == 0 {
		sb.WriteString("_No high-conviction bullish candidates today._\n\n")
	} else {
		for _, item := range b.BullishLookout {
			sb.WriteString(fmt.Sprintf("### %s (%s) — Value Score: %+d\n", item.Ticker, item.IssuerName, item.ValueScore))
			sb.WriteString(fmt.Sprintf("**Headline:** %s\n\n", item.Headline))
			sb.WriteString(fmt.Sprintf("**Rationale:** %s\n\n", item.Rationale))
			sb.WriteString(fmt.Sprintf("**Takeaway:** %s\n\n", item.InvestmentTakeaway))
		}
	}

	sb.WriteString("## 🔴 Stocks on Lookout (Risk Alerts / Headwinds)\n")
	if len(b.BearishLookout) == 0 {
		sb.WriteString("_No major risk alerts today._\n\n")
	} else {
		for _, item := range b.BearishLookout {
			sb.WriteString(fmt.Sprintf("### %s (%s) — Value Score: %+d\n", item.Ticker, item.IssuerName, item.ValueScore))
			sb.WriteString(fmt.Sprintf("**Headline:** %s\n\n", item.Headline))
			sb.WriteString(fmt.Sprintf("**Rationale:** %s\n\n", item.Rationale))
			sb.WriteString(fmt.Sprintf("**Takeaway:** %s\n\n", item.InvestmentTakeaway))
		}
	}

	sb.WriteString("## 🏭 Sector & Industry Highlights\n")
	for _, sec := range b.SectorHighlights {
		sb.WriteString(fmt.Sprintf("- **%s** [%s]: %s\n", sec.Sector, sec.Sentiment, sec.Summary))
	}
	sb.WriteString("\n")

	sb.WriteString("## 🎯 Value Investor Action Plan\n")
	sb.WriteString(b.ActionPlan + "\n")

	return sb.String()
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test -v ./internal/feature/news/...`  
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/feature/news/service.go internal/feature/news/service_test.go
git commit -m "feat(news): implement Daily Briefing generator with Gemini 3.7 Flash and Markdown formatter"
```

---

### Task 5: Scraper CLI Runner with 7 AM GMT+8 Window & Nuxt API Integration

**Files:**
- Modify: `cmd/scraper/main.go`
- Modify: `idx-web/src/server/utils/news-repo.ts`
- Modify: `idx-web/src/server/api/v1/news/index.get.ts`
- Create: `idx-web/src/server/utils/briefing-repo.ts`
- Create: `idx-web/src/server/api/v1/briefings/latest.get.ts`
- Create: `idx-web/src/server/api/v1/briefings/index.get.ts`

**Interfaces:**
- Scraper CLI runs yesterday $\rightarrow$ today GMT+8 window, skips duplicate links via MongoDB check, summarizes articles, generates Daily Briefing, persists and emails.
- Nuxt API exposes `/api/v1/briefings/latest`, `/api/v1/briefings`, and `GET /api/v1/news?ticker=BBRI`.

- [ ] **Step 1: Update `cmd/scraper/main.go`**

```go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	br "github.com/anandasatriaadi/go-idx-scraper/internal/browser"
	"github.com/anandasatriaadi/go-idx-scraper/internal/config"
	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/news"
	"github.com/anandasatriaadi/go-idx-scraper/internal/helper"
	newsRepo "github.com/anandasatriaadi/go-idx-scraper/internal/infra/db/mongo"
	"github.com/anandasatriaadi/go-idx-scraper/internal/infra/scraper/kontan"
	"go.mongodb.org/mongo-driver/v2/bson"
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
	var scrapeStartDate string
	var scrapeEndDate string
	var skipBriefing bool

	flag.StringVar(&configPath, "config", "config/config.yml", "Path to configuration file")
	flag.BoolVar(&noHeadless, "no-headless", false, "Disable headless mode for browser")
	flag.StringVar(&scrapeStartDate, "start-date", "", "Start date (YYYY-MM-DD), default yesterday GMT+8")
	flag.StringVar(&scrapeEndDate, "end-date", "", "End date (YYYY-MM-DD), default today GMT+8")
	flag.BoolVar(&skipBriefing, "skip-briefing", false, "Skip generating daily briefing")
	flag.Parse()

	logger, err := helper.NewLogger("scraper")
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer logger.Sync()

	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Error("Failed to load config", zap.Error(err))
		return err
	}

	if noHeadless {
		cfg.SetHeadless(false)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		logger.Info("Received shutdown signal", zap.String("signal", sig.String()))
		cancel()
	}()

	dbClient, err := newsRepo.NewClient(logger)
	if err != nil {
		logger.Error("Failed to connect to MongoDB", zap.Error(err))
		return err
	}

	db := dbClient.Database(cfg.Database.DbName)
	repo := newsRepo.NewNewsRepository(db)
	briefingRepo := newsRepo.NewBriefingRepository(db)
	service := news.NewService(repo, logger, cfg)

	browser, err := br.SetupSelenium(cfg)
	if err != nil {
		logger.Error("Failed to setup selenium", zap.Error(err))
		return err
	}
	defer browser.Close()

	scraper := kontan.NewScraper(logger, kontan.NewDefaultBrowser(browser.Driver))

	// GMT+8 (WITA / Singapore / Perth timezone)
	locGMT8 := time.FixedZone("GMT+8", 8*3600)
	nowGMT8 := time.Now().In(locGMT8)

	var startDate, endDate time.Time

	if scrapeStartDate == "" {
		startDate = nowGMT8.AddDate(0, 0, -1) // Yesterday
	} else {
		startDate, err = time.ParseInLocation("2006-01-02", scrapeStartDate, locGMT8)
		if err != nil {
			logger.Error("Failed to parse start-date", zap.Error(err))
			return err
		}
	}

	if scrapeEndDate == "" {
		endDate = nowGMT8 // Today
	} else {
		endDate, err = time.ParseInLocation("2006-01-02", scrapeEndDate, locGMT8)
		if err != nil {
			logger.Error("Failed to parse end-date", zap.Error(err))
			return err
		}
	}

	logger.Info("Starting scrape window in GMT+8", zap.Time("start", startDate), zap.Time("end", endDate))

	var ids []bson.ObjectID
	err = scraper.Scrape(ctx, startDate, endDate, func(n *news.News) error {
		// Idempotency: skip if already in MongoDB
		exists, err := repo.ExistsByLink(ctx, n.Link)
		if err != nil {
			logger.Warn("Failed to check if news exists by link", zap.String("link", n.Link), zap.Error(err))
		} else if exists {
			logger.Debug("News article already exists in DB, skipping", zap.String("link", n.Link))
			return nil
		}

		if err := service.Create(ctx, n); err != nil {
			return err
		}
		ids = append(ids, n.ID)
		return nil
	})

	if err != nil {
		logger.Error("Scraping finished with error", zap.Error(err))
	}

	if len(ids) > 0 {
		logger.Info("Summarizing newly fetched articles", zap.Int("count", len(ids)))
		if err := service.Summarize(ctx, ids); err != nil {
			logger.Error("Failed to summarize news", zap.Error(err))
		}
	}

	// Generate Daily Briefing and send email
	if !skipBriefing {
		briefing, err := service.GenerateDailyBriefing(ctx, nowGMT8, briefingRepo)
		if err != nil {
			logger.Error("Failed to generate Daily Briefing", zap.Error(err))
		} else if briefing != nil {
			logger.Info("Sending Daily Market Briefing email", zap.String("title", briefing.Title))
			if err := helper.SendMail(briefing.RawMarkdown, briefing.Title, cfg); err != nil {
				logger.Error("Failed to send Daily Briefing email", zap.Error(err))
			} else {
				logger.Info("Daily Briefing email successfully dispatched")
			}
		}
	}

	return nil
}
```

- [ ] **Step 2: Update Nuxt 4 backend utils & API endpoints**

Create `idx-web/src/server/utils/briefing-repo.ts`:
```typescript
import { Db, ObjectId } from 'mongodb'
import { Briefing } from './types'

export class BriefingRepository {
  private collection: any

  constructor(db: Db) {
    this.collection = db.collection('daily_briefings')
  }

  async findLatest(): Promise<Briefing | null> {
    const doc = await this.collection.find().sort({ date: -1 }).limit(1).toArray()
    if (!doc || doc.length === 0) return null
    return { ...doc[0], id: doc[0]._id.toString() }
  }

  async findAll(limit: number = 20, skip: number = 0): Promise<{ briefings: Briefing[]; total: number }> {
    const [briefings, total] = await Promise.all([
      this.collection.find().sort({ date: -1 }).skip(skip).limit(limit).toArray(),
      this.collection.countDocuments()
    ])
    return {
      briefings: briefings.map((b: any) => ({ ...b, id: b._id.toString() })),
      total
    }
  }
}
```

Create `idx-web/src/server/api/v1/briefings/latest.get.ts`:
```typescript
import { defineEventHandler } from 'h3'
import { useDb } from '../../../plugins/mongodb'
import { BriefingRepository } from '../../../utils/briefing-repo'

export default defineEventHandler(async (event) => {
  const db = useDb()
  const repo = new BriefingRepository(db)
  const latest = await repo.findLatest()
  if (!latest) {
    throw createError({ statusCode: 404, statusMessage: 'No briefing found' })
  }
  return latest
})
```

Create `idx-web/src/server/api/v1/briefings/index.get.ts`:
```typescript
import { defineEventHandler, getQuery } from 'h3'
import { useDb } from '../../../plugins/mongodb'
import { BriefingRepository } from '../../../utils/briefing-repo'

export default defineEventHandler(async (event) => {
  const query = getQuery(event)
  const limit = parseInt(String(query.limit || '20'), 10)
  const skip = parseInt(String(query.skip || '0'), 10)

  const db = useDb()
  const repo = new BriefingRepository(db)
  return await repo.findAll(limit, skip)
})
```

Update `idx-web/src/server/utils/news-repo.ts` to support `ticker` and `industry` queries:
```typescript
if (filter.ticker) {
  query.tickers = filter.ticker.toUpperCase()
}
if (filter.industry) {
  query.industry = filter.industry
}
```

- [ ] **Step 3: Run Go build & test**

Run: `go build -o bin/ ./cmd/...`  
Run: `go test -count=1 ./...`  
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add cmd/scraper/main.go idx-web/src/server/
git commit -m "feat: integrate 7 AM GMT+8 window, daily briefing emailing, and Nuxt 4 briefing endpoints"
```

---

### Task 6: Verification & End-to-End Build

**Files:**
- Full repository verification.

- [ ] **Step 1: Run complete test suite**

Run: `go test -count=1 ./...`  
Expected: All tests pass.

- [ ] **Step 2: Verify binary compilation**

Run: `go build -o bin/ ./cmd/...`  
Expected: `bin/scraper`, `bin/downloader`, `bin/issuer`, `bin/announcement` built cleanly.

- [ ] **Step 3: Verify git status is clean**

Run: `git status`  
Expected: Working tree clean.
