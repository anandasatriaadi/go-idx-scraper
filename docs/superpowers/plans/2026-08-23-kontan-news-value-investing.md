# Kontan Multi-Channel Scraper & Value Investing Analysis Engine Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend the Kontan news scraper to automatically ingest articles across both `investasi` and `keuangan` channels with deduplication, and enhance the OpenRouter Gemini summarizer to evaluate articles strictly through a Value Investing framework with objective `-10` to `+10` scoring, impact direction, and investment takeaways.

**Architecture:** Hexagonal DDD architecture where the Scraper Adapter (`internal/infra/scraper/kontan`) queries multiple channels and forwards cleaned articles to the News Service (`internal/feature/news`). The service invokes OpenRouter Gemini using a strict JSON schema for structured Value Investing metrics (`value_score`, `impact_direction`, `investment_takeaway`), persists them in MongoDB, and exposes them across Go domain models and Nuxt 4 API TypeScript types.

**Tech Stack:** Go 1.24+, MongoDB v2 Driver (`go.mongodb.org/mongo-driver/v2`), OpenRouter SDK (`github.com/revrost/go-openrouter`), Selenium WebDriver, Goquery, Bluemonday, html-to-markdown/v2, TypeScript (Nuxt 4).

## Global Constraints

- Never use `context.Background()` in domain logic; propagate context from callers.
- Wrap errors with `%w` for error context across layer boundaries.
- Use `zap.Logger` structured logging (`zap.String`, `zap.Int`, `zap.Error`).
- Maintain existing repository port interfaces in `internal/feature/news/entity.go`.
- Value score scale: integer strictly from `-10` to `+10`.
- Impact direction: exactly `"Bullish"`, `"Bearish"`, or `"Neutral"`.

---

### Task 1: Extend Domain Entity & API Interfaces (Go & TypeScript)

**Files:**
- Modify: `internal/feature/news/entity.go:1-32`
- Modify: `idx-web/src/server/utils/types.ts:30-45`
- Test: `internal/feature/news/service_test.go`

**Interfaces:**
- Produces: `News` entity with `ValueScore`, `ImpactDirection`, `InvestmentTakeaway` fields.

- [ ] **Step 1: Write test verifying entity fields serialization and structure**

Add `TestNewsEntity_Fields` in `internal/feature/news/service_test.go`:

```go
func TestNewsEntity_Fields(t *testing.T) {
	id := bson.NewObjectID()
	n := &News{
		ID:                 id,
		Title:              "Sample News",
		Summary:            "3 sentence summary.",
		Content:            "Full markdown content.",
		Priority:           5,
		ValueScore:         7,
		ImpactDirection:    "Bullish",
		InvestmentTakeaway: "Strong cash flow and expansion potential.",
	}

	if n.ValueScore != 7 {
		t.Errorf("Expected ValueScore 7, got %d", n.ValueScore)
	}
	if n.ImpactDirection != "Bullish" {
		t.Errorf("Expected ImpactDirection 'Bullish', got '%s'", n.ImpactDirection)
	}
	if n.InvestmentTakeaway != "Strong cash flow and expansion potential." {
		t.Errorf("Expected InvestmentTakeaway to match, got '%s'", n.InvestmentTakeaway)
	}
}
```

- [ ] **Step 2: Run test to verify it fails (fields not defined on News)**

Run: `go test -v ./internal/feature/news -run TestNewsEntity_Fields`  
Expected: Compilation failure due to undefined fields `ValueScore`, `ImpactDirection`, `InvestmentTakeaway`.

- [ ] **Step 3: Update `internal/feature/news/entity.go` and `idx-web/src/server/utils/types.ts`**

Update `internal/feature/news/entity.go`:
```go
type News struct {
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
	ID                 bson.ObjectID `bson:"_id,omitempty" json:"id"`
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
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -v ./internal/feature/news -run TestNewsEntity_Fields`  
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/feature/news/entity.go internal/feature/news/service_test.go idx-web/src/server/utils/types.ts
git commit -m "feat(news): add value investing score and takeaway fields to news entity"
```

---

### Task 2: Multi-Channel Kontan Scraping & URL Deduplication

**Files:**
- Modify: `internal/infra/scraper/kontan/scraper.go:100-195`
- Modify: `internal/infra/scraper/kontan/scraper_test.go`

**Interfaces:**
- Consumes: `Browser` interface (`FetchList`, `FetchContent`).
- Produces: `Scraper.Scrape(ctx, startDate, endDate, onNewsFound)` querying both `investasi` and `keuangan` channels.

- [ ] **Step 1: Write the failing tests for multi-channel scraping and deduplication**

Add `TestScrape_MultiChannelAndDeduplication` in `internal/infra/scraper/kontan/scraper_test.go`:

```go
func TestScrape_MultiChannelAndDeduplication(t *testing.T) {
	logger := zap.NewNop()
	scraper := NewScraper(logger, nil)

	var fetchedURLs []string
	investasiListHTML := `
		<li>
			<div class="ket">
				<h1><a href="https://investasi.kontan.co.id/news/news-1">News 1</a></h1>
				<div class="fs14"><span>Investasi</span><span>| 05 Februari 2024</span></div>
			</div>
		</li>
		<li>
			<div class="ket">
				<h1><a href="https://investasi.kontan.co.id/news/shared-news">Shared News</a></h1>
				<div class="fs14"><span>Investasi</span><span>| 05 Februari 2024</span></div>
			</div>
		</li>
	`
	keuanganListHTML := `
		<li>
			<div class="ket">
				<h1><a href="https://keuangan.kontan.co.id/news/shared-news">Shared News</a></h1>
				<div class="fs14"><span>Keuangan</span><span>| 05 Februari 2024</span></div>
			</div>
		</li>
		<li>
			<div class="ket">
				<h1><a href="https://keuangan.kontan.co.id/news/news-2">News 2</a></h1>
				<div class="fs14"><span>Keuangan</span><span>| 05 Februari 2024</span></div>
			</div>
		</li>
	`

	mock := &RefinedMockBrowser{
		FetchListFunc: func(ctx context.Context, url string) (string, error) {
			fetchedURLs = append(fetchedURLs, url)
			if strings.Contains(url, "kanal=investasi") {
				return investasiListHTML, nil
			}
			if strings.Contains(url, "kanal=keuangan") {
				return keuanganListHTML, nil
			}
			return "", nil
		},
		FetchContentFunc: func(ctx context.Context, url string) (string, string, error) {
			return "<p>Clean content</p>", "", nil
		},
	}
	scraper.WithBrowser(mock)

	date := time.Date(2024, 2, 5, 0, 0, 0, 0, time.UTC)
	var processedTitles []string
	err := scraper.Scrape(context.Background(), date, date, func(n *news.News) error {
		processedTitles = append(processedTitles, n.Title)
		return nil
	})

	if err != nil {
		t.Fatalf("Scrape failed: %v", err)
	}

	// Verify both channels were fetched
	hasInvestasi := false
	hasKeuangan := false
	for _, u := range fetchedURLs {
		if strings.Contains(u, "kanal=investasi") {
			hasInvestasi = true
		}
		if strings.Contains(u, "kanal=keuangan") {
			hasKeuangan = true
		}
	}
	if !hasInvestasi || !hasKeuangan {
		t.Errorf("Expected both investasi and keuangan channels to be fetched, got URLs: %v", fetchedURLs)
	}

	// Verify deduplication (Shared News should only be processed once if identical link, or handled properly)
	if len(processedTitles) != 3 {
		// news-1, shared-news, news-2
		t.Errorf("Expected 3 unique news articles, got %d: %v", len(processedTitles), processedTitles)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v ./internal/infra/scraper/kontan/ -run TestScrape_MultiChannelAndDeduplication`  
Expected: FAIL (scraper currently only queries `kanal=investasi`).

- [ ] **Step 3: Update `scraper.go` to iterate over both channels and deduplicate seen URLs**

In `internal/infra/scraper/kontan/scraper.go`:
1. Update `fetchArticleListHTML` signature or implementation to accept `kanal string`:
```go
func (s *Scraper) fetchArticleListHTML(ctx context.Context, kanal string, date time.Time, perPage int) (string, error) {
	var perPageStr string
	if perPage == 0 {
		perPageStr = ""
	} else {
		perPageStr = fmt.Sprintf("%d", perPage)
	}
	day := fmt.Sprintf("%02d", date.Day())
	month := fmt.Sprintf("%02d", int(date.Month()))
	year := date.Year()
	url := fmt.Sprintf("https://www.kontan.co.id/search/indeks?kanal=%s&tanggal=%s&bulan=%s&tahun=%d&pos=indeks&per_page=%s", kanal, day, month, year, perPageStr)
	htmlContent, err := s.browser.FetchList(ctx, url)
	if err != nil {
		return "", fmt.Errorf("browser fetch error for list: %w", err)
	}
	return htmlContent, nil
}
```
2. Update `Scrape` method to loop over `channels := []string{"investasi", "keuangan"}` and maintain `seenLinks := make(map[string]bool)`. Normalize links (e.g. stripping query/trailing slashes if needed, or by path key) so duplicate articles across channels are skipped.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test -v ./internal/infra/scraper/kontan/...`  
Expected: PASS across all tests.

- [ ] **Step 5: Commit**

```bash
git add internal/infra/scraper/kontan/scraper.go internal/infra/scraper/kontan/scraper_test.go
git commit -m "feat(scraper): support multi-channel scraping for investasi and keuangan with deduplication"
```

---

### Task 3: Value Investing AI Analysis Engine & OpenRouter Gemini Integration

**Files:**
- Modify: `internal/feature/news/service.go:35-115`
- Modify: `internal/feature/news/service_test.go`

**Interfaces:**
- Produces: `NewsSummary` with Value Investing schema and `Service.Summarize(ctx, ids)` populating `priority`, `value_score`, `impact_direction`, and `investment_takeaway`.

- [ ] **Step 1: Write failing unit test for `NewsSummary` schema and serialization**

Add `TestNewsSummary_SchemaAndPrompt` in `internal/feature/news/service_test.go`:

```go
func TestNewsSummary_SchemaAndPrompt(t *testing.T) {
	schema, err := jsonschema.GenerateSchemaForType(NewsSummary{})
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}
	if schema == nil {
		t.Fatal("Expected non-nil schema")
	}

	jsonSample := `{
		"title": "Emiten BBRI Perkuat Pendanaan",
		"summary": "BBRI membukukan pertumbuhan laba bersih dan peningkatan margin bunga bersih. Pertumbuhan ini didorong oleh segmen mikro yang solid. Manajemen mempertahankan rasio dividen tinggi.",
		"priority": 2,
		"value_score": 8,
		"impact_direction": "Bullish",
		"investment_takeaway": "Fundamental kuat dengan moat kokoh di pembiayaan mikro dan valuasi menarik."
	}`

	var summary NewsSummary
	if err := json.Unmarshal([]byte(jsonSample), &summary); err != nil {
		t.Fatalf("Failed to unmarshal NewsSummary: %v", err)
	}

	if summary.ValueScore != 8 {
		t.Errorf("Expected ValueScore 8, got %d", summary.ValueScore)
	}
	if summary.ImpactDirection != "Bullish" {
		t.Errorf("Expected ImpactDirection 'Bullish', got '%s'", summary.ImpactDirection)
	}
	if summary.InvestmentTakeaway == "" {
		t.Errorf("Expected non-empty InvestmentTakeaway")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v ./internal/feature/news/ -run TestNewsSummary_SchemaAndPrompt`  
Expected: FAIL (missing fields in `NewsSummary` struct).

- [ ] **Step 3: Update `NewsSummary` struct and `Summarize` prompt in `internal/feature/news/service.go`**

1. Update `NewsSummary` struct:
```go
type NewsSummary struct {
	Title              string `json:"title" jsonschema:"description=An updated engaging and objective title capturing the essence of the article"`
	Summary            string `json:"summary" jsonschema:"description=Concise 3-sentence summary highlighting financial facts figures and immediate market implications"`
	Priority           int    `json:"priority" jsonschema:"description=Market impact priority from 1 (highest market urgency/impact) to 10 (routine/low market impact)"`
	ValueScore         int    `json:"value_score" jsonschema:"description=Fundamental value investing impact score strictly between -10 and +10 (-10 is severe fundamental impairment, 0 is neutral/macro noise, +10 is massive fundamental value creation)"`
	ImpactDirection    string `json:"impact_direction" jsonschema:"enum=Bullish,enum=Bearish,enum=Neutral,description=Directional impact on underlying business intrinsic value"`
	InvestmentTakeaway string `json:"investment_takeaway" jsonschema:"description=1-2 sentence actionable takeaway strictly from the perspective of a disciplined long-term value investor focusing on moat, capital allocation, cash flows, and intrinsic value"`
}
```

2. Update System Prompt & OpenRouter request in `Summarize`:
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

Article:
"""
%s
"""`, n.Content)
```

3. Update the MongoDB `$set` payload in `Summarize` to persist `value_score`, `impact_direction`, and `investment_takeaway`:
```go
update := map[string]any{
	"$set": map[string]any{
		"title":               summary.Title,
		"summary":             summary.Summary,
		"priority":            summary.Priority,
		"value_score":         summary.ValueScore,
		"impact_direction":    summary.ImpactDirection,
		"investment_takeaway": summary.InvestmentTakeaway,
		"updated_at":          time.Now(),
	},
}
```

- [ ] **Step 4: Run unit tests to verify they pass**

Run: `go test -v ./internal/feature/news/...`  
Expected: PASS

- [ ] **Step 5: Run full project test suite**

Run: `go test ./...`  
Expected: PASS across all packages.

- [ ] **Step 6: Commit**

```bash
git add internal/feature/news/service.go internal/feature/news/service_test.go
git commit -m "feat(news): implement objective value investing analysis engine with OpenRouter Gemini"
```

---

### Task 4: Verification & End-to-End Build

**Files:**
- Test all CLI entry points and packages.

- [ ] **Step 1: Build all CLI binaries**

Run: `go build -o bin/ ./cmd/...`  
Expected: Clean build without errors for `bin/downloader`, `bin/scraper`, `bin/issuer`, `bin/announcement`.

- [ ] **Step 2: Run all tests in repository**

Run: `go test -v ./...`  
Expected: All tests PASS.

- [ ] **Step 3: Verify git status is clean**

Run: `git status`  
Expected: Working tree clean.
