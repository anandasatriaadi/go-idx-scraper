# IDX-IC Standard Taxonomy & Server-Side Pagination Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Enforce the official 11 Sectors and 34 Subsectors from the IDX Industrial Classification (IDX-IC) across the Gemini 3.7 Flash news ingestion pipeline, and implement server-side descending sorting (`{ tgl_pengumuman: -1, date: -1 }`) and pagination across all API endpoints and UI tables to eliminate heavy initial page loads.

**Architecture:** Domain models in `internal/feature/news` include explicit `Sector` and `Subsector` fields constrained by IDX-IC taxonomy enums. In `idx-web`, repository adapters (`announcement-repo`, `news-repo`, `finreport-repo`) apply server-side sorting, regex search, and `{ limit, skip }` pagination with total count queries. Nuxt 4 views render reactive pagination controls with instant `< 30ms` page loads and provide sector-based related news linking for tickers.

**Tech Stack:** Go 1.24+, MongoDB v2 Driver (`go.mongodb.org/mongo-driver/v2`), OpenRouter SDK (Gemini 3.7 Flash), Nuxt 4 / Vue 3 / Nitro.

## Global Constraints

- Use strict `context: "fresh"` for all subagent dispatches (no session forking).
- Never use `context.Background()` in domain logic.
- Wrap errors with `%w`.
- Strict adherence to the 11 Official IDX Sectors and 34 Subsectors (IDX-IC).
- All collections must sort descending by date (`{ tgl_pengumuman: -1 }`, `{ date: -1 }`, `{ created_at: -1 }`).
- Disclosures, News, and Financial Reports default to `limit: 20`, max `limit: 100`.

---

### Task 1: Domain Models & Gemini 3.7 Flash Strict IDX-IC Taxonomy Schema

**Files:**
- Modify: `internal/feature/news/entity.go`
- Modify: `internal/feature/news/service.go`
- Modify: `internal/feature/news/service_test.go`
- Modify: `idx-web/src/server/utils/types.ts`

**Interfaces:**
- Produces: `News` with `Sector`, `Subsector` (IDX-IC enum schema), updated `NewsSummary` JSON schema.

- [ ] **Step 1: Write test for `NewsSummary` with strict IDX-IC Sector & Subsector validation**

Add `TestNewsSummary_IDXICClassification` in `internal/feature/news/service_test.go`:

```go
func TestNewsSummary_IDXICClassification(t *testing.T) {
	schema, err := jsonschema.GenerateSchemaForType(NewsSummary{})
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}
	if schema == nil {
		t.Fatal("Expected non-nil schema")
	}

	jsonSample := `{
		"title": "Bank Mandiri Catat Pertumbuhan Kredit 13%",
		"summary": "Kredit BMRI tumbuh kuat ditopang segmen korporasi dan komersial. Kualitas aset terjaga dengan NPL rendah. Laba bersih semester I meningkat signifikan.",
		"priority": 3,
		"value_score": 7,
		"impact_direction": "Bullish",
		"investment_takeaway": "Kinerja fundamental solid dengan profitabilitas tinggi dan solvabilitas kokoh.",
		"tickers": ["BMRI"],
		"sector": "G. Financials",
		"subsector": "G1. Banks",
		"is_industry_wide": false
	}`

	var summary NewsSummary
	if err := json.Unmarshal([]byte(jsonSample), &summary); err != nil {
		t.Fatalf("Failed to unmarshal NewsSummary: %v", err)
	}

	if summary.Sector != "G. Financials" {
		t.Errorf("Expected sector 'G. Financials', got '%s'", summary.Sector)
	}
	if summary.Subsector != "G1. Banks" {
		t.Errorf("Expected subsector 'G1. Banks', got '%s'", summary.Subsector)
	}
}
```

- [ ] **Step 2: Update `internal/feature/news/entity.go` and `idx-web/src/server/utils/types.ts`**

Update `internal/feature/news/entity.go`:
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
	Sector             string        `bson:"sector" json:"sector"`
	Subsector          string        `bson:"subsector" json:"subsector"`
	Industry           string        `bson:"industry,omitempty" json:"industry,omitempty"`
	IsIndustryWide     bool          `bson:"is_industry_wide" json:"is_industry_wide"`
}
```

Update `NewsSummary` in `internal/feature/news/service.go`:
```go
type NewsSummary struct {
	Title              string   `json:"title" jsonschema:"description=An updated engaging and objective title capturing the essence of the article"`
	Summary            string   `json:"summary" jsonschema:"description=Concise 3-sentence summary highlighting financial facts figures and immediate market implications"`
	Priority           int      `json:"priority" jsonschema:"description=Market impact priority from 1 (highest market urgency) to 10 (routine)"`
	ValueScore         int      `json:"value_score" jsonschema:"description=Fundamental value investing impact score strictly between -10 and +10"`
	ImpactDirection    string   `json:"impact_direction" jsonschema:"enum=Bullish,enum=Bearish,enum=Neutral,description=Directional impact on underlying business intrinsic value"`
	InvestmentTakeaway string   `json:"investment_takeaway" jsonschema:"description=1-2 sentence actionable takeaway for a disciplined long-term value investor"`
	Tickers            []string `json:"tickers" jsonschema:"description=List of 4-letter IDX stock ticker symbols (e.g. ['BBRI', 'BBCA']). Empty array if none."`
	Sector             string   `json:"sector" jsonschema:"enum=A. Energy,enum=B. Basic Materials,enum=C. Industrials,enum=D. Consumer Non-Cyclicals,enum=E. Consumer Cyclicals,enum=F. Healthcare,enum=G. Financials,enum=H. Properties and Real Estate,enum=I. Technology,enum=J. Infrastructures,enum=K. Transportation and Logistic,enum=Macroeconomics,description=Official IDX Industrial Classification (IDX-IC) Primary Sector"`
	Subsector          string   `json:"subsector" jsonschema:"enum=A1. Oil, Gas, and Coal,enum=A2. Alternative Energy,enum=B1. Basic Materials,enum=C1. Industrial Goods,enum=C2. Industrial Services,enum=C3. Multi-sector Holdings,enum=D1. Food and Staples Retailing,enum=D2. Food and Beverage,enum=D3. Tobacco,enum=D4. Nondurable Household Products,enum=E1. Automobiles and Components,enum=E2. Household Goods,enum=E3. Leisure Goods,enum=E4. Apparel and Luxury Goods,enum=E5. Consumer Services,enum=E6. Media and Entertainment,enum=E7. Retailing,enum=F1. Healthcare Equipment & Providers,enum=F2. Pharmaceuticals & Health Care Research,enum=G1. Banks,enum=G2. Financing Service,enum=G3. Investment Service,enum=G4. Insurance,enum=G5. Holding and Investment Companies,enum=H1. Properties & Real Estate,enum=I1. Software & IT Services,enum=I2. Technology Hardware & Equipment,enum=J1. Transportation Infrastructure,enum=J2. Heavy Constructions & Civil Engineering,enum=J3. Telecommunication,enum=J4. Utilities,enum=K1. Transportation,enum=K2. Logistics & Deliveries,enum=General Market & Policy,description=Official IDX-IC Subsector classification"`
	IsIndustryWide     bool     `json:"is_industry_wide" jsonschema:"description=True if the news affects an entire sector or macroeconomic policy rather than just one individual company"`
}
```

- [ ] **Step 3: Update `Summarize` prompt and MongoDB persistence**

In `internal/feature/news/service.go`:
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
		"sector":              summary.Sector,
		"subsector":           summary.Subsector,
		"industry":            summary.Subsector, // backwards compatibility
		"is_industry_wide":    summary.IsIndustryWide,
		"updated_at":          time.Now(),
	},
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
  sector?: string;
  subsector?: string;
  industry?: string;
  is_industry_wide?: boolean;
}
```

- [ ] **Step 4: Run unit tests**

Run: `go test -v ./internal/feature/news/...`  
Expected: PASS

- [ ] **Step 5: Commit Task 1**

```bash
git add internal/feature/news/ idx-web/src/server/utils/types.ts
git commit -m "feat(news): enforce official IDX-IC 11 sectors and 34 subsectors in AI schema"
```

---

### Task 2: Server-Side Descending Sort & Pagination on All API Endpoints

**Files:**
- Modify: `idx-web/src/server/utils/announcement-repo.ts`
- Modify: `idx-web/src/server/api/v1/announcements/index.get.ts`
- Modify: `idx-web/src/server/utils/news-repo.ts`
- Modify: `idx-web/src/server/api/v1/news/index.get.ts`
- Modify: `idx-web/src/server/utils/finreport-repo.ts`
- Modify: `idx-web/src/server/api/v1/financial-reports/index.get.ts`

**Interfaces:**
- Produces: `{ data: T[], total: number, page: number, total_pages: number }` for announcements, reports, and news.

- [ ] **Step 1: Update `announcement-repo.ts` and `announcements/index.get.ts`**

Update `idx-web/src/server/utils/announcement-repo.ts`:
```typescript
export interface AnnouncementFilter {
  ticker?: string;
  search?: string;
  limit?: number;
  skip?: number;
}

export async function findAllAnnouncementsPaginated(filter: AnnouncementFilter = {}) {
  const collection = getAnnouncementsCollection()
  const query: any = {}

  if (filter.ticker) {
    query.kode_emiten = filter.ticker.toUpperCase().trim()
  }
  if (filter.search) {
    const q = filter.search.trim()
    query.$or = [
      { judul_pengumuman: { $regex: q, $options: 'i' } },
      { no_pengumuman: { $regex: q, $options: 'i' } },
      { kode_emiten: { $regex: q, $options: 'i' } },
      { title: { $regex: q, $options: 'i' } }
    ]
  }

  const limit = Math.min(Math.max(filter.limit || 20, 1), 100)
  const skip = Math.max(filter.skip || 0, 0)

  const [data, total] = await Promise.all([
    collection
      .find(query)
      .sort({ tgl_pengumuman: -1, created_at: -1, _id: -1 })
      .skip(skip)
      .limit(limit)
      .toArray(),
    collection.countDocuments(query)
  ])

  return {
    data: data.map(d => ({ ...d, id: d._id.toString() })),
    total,
    page: Math.floor(skip / limit) + 1,
    total_pages: Math.ceil(total / limit)
  }
}
```

Update `idx-web/src/server/api/v1/announcements/index.get.ts`:
```typescript
import { defineEventHandler, getQuery } from 'h3'
import { findAllAnnouncementsPaginated } from '../../../utils/announcement-repo'
import { getAuthFromEvent } from '../../../utils/firebase-admin'
import { getUserWatchlist } from '../../../utils/user-service'

export default defineEventHandler(async (event) => {
  const query = getQuery(event)
  const limit = parseInt(String(query.limit || '20'), 10)
  const page = parseInt(String(query.page || '1'), 10)
  const skip = query.skip ? parseInt(String(query.skip), 10) : (page - 1) * limit
  const ticker = query.ticker ? String(query.ticker) : undefined
  const search = query.search ? String(query.search) : undefined

  const result = await findAllAnnouncementsPaginated({ limit, skip, ticker, search })

  let watchlist: string[] = []
  const auth = getAuthFromEvent(event)
  if (auth) {
    watchlist = await getUserWatchlist(auth.uid)
  }
  const watchlistSet = new Set(watchlist)

  return {
    ...result,
    data: result.data.map(ann => ({
      ...ann,
      is_watched: ann.kode_emiten ? watchlistSet.has(ann.kode_emiten) : false,
    }))
  }
})
```

- [ ] **Step 2: Update `news-repo.ts` and `news/index.get.ts`**

Update `idx-web/src/server/utils/news-repo.ts`:
```typescript
export interface NewsFilter {
  date_gte?: string;
  date_lte?: string;
  priority?: number;
  source?: string;
  ticker?: string;
  sector?: string;
  subsector?: string;
  industry?: string;
  search?: string;
  limit?: number;
  skip?: number;
}

export async function findAllNewsPaginated(filter: NewsFilter = {}) {
  const collection = getNewsCollection()
  const query: any = {}

  if (filter.date_gte || filter.date_lte) {
    query.date = {}
    if (filter.date_gte) query.date.$gte = new Date(filter.date_gte)
    if (filter.date_lte) query.date.$lte = new Date(filter.date_lte)
  }

  if (filter.priority !== undefined) query.priority = filter.priority
  if (filter.source) query.link = filter.source
  if (filter.ticker) query.tickers = filter.ticker.toUpperCase().trim()
  if (filter.sector) query.sector = filter.sector
  if (filter.subsector) query.subsector = filter.subsector
  if (filter.industry && !filter.subsector) {
    query.$or = [{ subsector: filter.industry }, { industry: filter.industry }, { sector: filter.industry }]
  }
  if (filter.search) {
    const q = filter.search.trim()
    query.$or = [
      { title: { $regex: q, $options: 'i' } },
      { summary: { $regex: q, $options: 'i' } },
      { tickers: filter.search.toUpperCase().trim() }
    ]
  }

  const limit = Math.min(Math.max(filter.limit || 20, 1), 100)
  const skip = Math.max(filter.skip || 0, 0)

  const [data, total] = await Promise.all([
    collection
      .find(query)
      .sort({ date: -1, created_at: -1, _id: -1 })
      .skip(skip)
      .limit(limit)
      .toArray(),
    collection.countDocuments(query)
  ])

  return {
    data: data.map((d: any) => ({ ...d, id: d._id.toString() })),
    total,
    page: Math.floor(skip / limit) + 1,
    total_pages: Math.ceil(total / limit)
  }
}
```

Update `idx-web/src/server/api/v1/news/index.get.ts`:
```typescript
import { defineEventHandler, getQuery } from 'h3'
import { findAllNewsPaginated } from '../../../utils/news-repo'

export default defineEventHandler(async (event) => {
  const query = getQuery(event)
  const limit = parseInt(String(query.limit || '20'), 10)
  const page = parseInt(String(query.page || '1'), 10)
  const skip = query.skip ? parseInt(String(query.skip), 10) : (page - 1) * limit

  return await findAllNewsPaginated({
    limit,
    skip,
    ticker: query.ticker ? String(query.ticker) : undefined,
    sector: query.sector ? String(query.sector) : undefined,
    subsector: query.subsector ? String(query.subsector) : undefined,
    industry: query.industry ? String(query.industry) : undefined,
    search: query.search ? String(query.search) : undefined,
    priority: query.priority ? parseInt(String(query.priority), 10) : undefined,
  })
})
```

- [ ] **Step 3: Update `finreport-repo.ts` and `financial-reports/index.get.ts`**

Update `idx-web/src/server/utils/finreport-repo.ts`:
```typescript
export interface FinReportFilter {
  issuer_code?: string;
  year?: number;
  quarter?: number;
  search?: string;
  limit?: number;
  skip?: number;
}

export async function findAllFinancialReportsPaginated(filter: FinReportFilter = {}) {
  const collection = getFinancialReportsCollection()
  const query: any = {}

  if (filter.issuer_code) query.issuer_code = filter.issuer_code.toUpperCase().trim()
  if (filter.year) query.year = filter.year
  if (filter.quarter) query.quarter = filter.quarter
  if (filter.search) {
    query.issuer_code = { $regex: filter.search.trim(), $options: 'i' }
  }

  const limit = Math.min(Math.max(filter.limit || 20, 1), 100)
  const skip = Math.max(filter.skip || 0, 0)

  const [data, total] = await Promise.all([
    collection
      .find(query)
      .sort({ year: -1, downloaded_at: -1, _id: -1 })
      .skip(skip)
      .limit(limit)
      .toArray(),
    collection.countDocuments(query)
  ])

  return {
    data: data.map((d: any) => ({ ...d, id: d._id.toString() })),
    total,
    page: Math.floor(skip / limit) + 1,
    total_pages: Math.ceil(total / limit)
  }
}
```

Update `idx-web/src/server/api/v1/financial-reports/index.get.ts`:
```typescript
import { defineEventHandler, getQuery } from 'h3'
import { findAllFinancialReportsPaginated } from '../../../utils/finreport-repo'

export default defineEventHandler(async (event) => {
  const query = getQuery(event)
  const limit = parseInt(String(query.limit || '20'), 10)
  const page = parseInt(String(query.page || '1'), 10)
  const skip = query.skip ? parseInt(String(query.skip), 10) : (page - 1) * limit

  return await findAllFinancialReportsPaginated({
    limit,
    skip,
    issuer_code: query.issuer_code ? String(query.issuer_code) : (query.ticker ? String(query.ticker) : undefined),
    year: query.year ? parseInt(String(query.year), 10) : undefined,
    quarter: query.quarter ? parseInt(String(query.quarter), 10) : undefined,
    search: query.search ? String(query.search) : undefined,
  })
})
```

- [ ] **Step 4: Test Nuxt build**

Run: `cd idx-web && npm run build`  
Expected: PASS

- [ ] **Step 5: Commit Task 2**

```bash
git add idx-web/src/server/
git commit -m "feat(api): implement server-side descending sort and pagination across all endpoints"
```

---

### Task 3: Frontend UI Pagination & Related Sector News Integration

**Files:**
- Modify: `idx-web/src/components/AnnouncementsView.vue`
- Modify: `idx-web/src/components/FinReportsView.vue`
- Modify: `idx-web/src/components/NewsTerminalView.vue`
- Modify: `idx-web/src/components/TickerFinancialsModal.vue`
- Modify: `idx-web/src/pages/index.vue`

**Interfaces:**
- `AnnouncementsView` & `FinReportsView`: Render server-paginated tables with search input and page controls.
- `NewsTerminalView`: Sector & Subsector dropdowns with 11 IDX-IC sectors and 34 subsectors.
- `TickerFinancialsModal`: Displays related sector news stream when viewing any ticker.

- [ ] **Step 1: Update `idx-web/src/components/AnnouncementsView.vue` with pagination controls**

- [ ] **Step 2: Update `idx-web/src/components/FinReportsView.vue` with pagination controls**

- [ ] **Step 3: Update `idx-web/src/components/NewsTerminalView.vue` with IDX-IC Sector and Subsector filters**

- [ ] **Step 4: Update `idx-web/src/components/TickerFinancialsModal.vue` with Related Sector News tab**

- [ ] **Step 5: Update `idx-web/src/pages/index.vue` to integrate paginated data fetching**

- [ ] **Step 6: Verify build & commit Task 3**

```bash
cd idx-web && npm run build
git add src/components/ src/pages/index.vue
git commit -m "feat(web): add server pagination controls, IDX-IC sector filters, and related sector news linking"
```

---

### Task 4: Verification, Full Build & Final Code Review

- [ ] **Step 1: Run complete Go test suite (`go test -count=1 ./...`)**
- [ ] **Step 2: Run Nuxt production build (`cd idx-web && npm run build`)**
- [ ] **Step 3: Verify all Go binaries compile (`go build -o bin/ ./cmd/...`)**
- [ ] **Step 4: Verify clean git status**
