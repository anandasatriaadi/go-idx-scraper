# Design Document: Facebook Group Car Listing Scraper (`jualbelimobilbekasbali`)

**Date:** 2026-08-26  
**Target Repository:** `/Users/ananda/gitrepo/fb-group-scrape`  
**Go Module:** `github.com/anandasatriaadi/fb-group-scrape`  
**Architecture:** Hexagonal Architecture + Tactical Domain-Driven Design (DDD)  
**Database:** MongoDB via `go.mongodb.org/mongo-driver/v2`

---

## 1. Overview & Objectives

Build a dedicated, high-performance scraper for the Facebook group `https://www.facebook.com/groups/jualbelimobilbekasbali` to collect used car listings.

### Core Capabilities:
1. **Interactive Login with Persistent Profile**:
   - Spawns Chrome in visible mode on first run or when `-login` is specified.
   - Prompts the user to log in via browser window, saving all session tokens and cookies into `--user-data-dir=./profiles/facebook`.
   - Subsequent scraper runs execute in **headless mode** without user intervention.
2. **Chronological & Time-Bounded Scrolling**:
   - Enforces newest-to-oldest post sorting via `?sorting_setting=CHRONOLOGICAL`.
   - Scrolls down the feed dynamically until posts older than **2 days** (`-days=2`) are reached.
   - Automatically detects previously scraped posts and stops upon encountering a **known post ID** from MongoDB (incremental scraping).
3. **DOM & Listing Extraction**:
   - Injects an in-browser JavaScript evaluation parser to extract post ID, permalink, author name & URL, timestamp, listing description, structured price tags, and photo CDN URLs.
4. **Hexagonal Architecture (Ports & Adapters)**:
   - Clean layer separation: Domain Core (no external I/O), Application Use Cases (orchestration), and Infrastructure Adapters (Selenium Browser & MongoDB v2 persistence).
5. **MongoDB v2 Persistence**:
   - Stores scraped posts into `facebook_group_posts` collection with compound unique indexing `{ "group_slug": 1, "id": 1 }` and sorted query indexing `{ "group_slug": 1, "posted_at": -1 }`.

---

## 2. Directory Layout & Hexagonal DDD Architecture

```
/Users/ananda/gitrepo/fb-group-scrape/
├── cmd/
│   ├── example/main.go                 # Scraper engine demo
│   └── scrape_group/main.go            # Driver Adapter: Scraper CLI application
├── internal/
│   ├── domain/                         # Pure Domain Layer (No external dependencies)
│   │   └── post/
│   │       ├── entity.go               # Post aggregate root, Author, Price value objects
│   │       ├── repository.go           # PostRepository interface (Driven Port)
│   │       └── service.go              # Timestamp parsing, deduplication, price normalization
│   ├── application/                    # Application Layer (Use Cases)
│   │   └── scrape/
│   │       ├── command.go              # ScrapeGroupCommand, ScrapeGroupResult DTOs
│   │       ├── port.go                 # FacebookBrowserPort interface (Driven Port)
│   │       └── handler.go              # ScrapeGroupHandler use case
│   └── infra/                          # Infrastructure Layer (Driven Adapters)
│       ├── browser/
│       │   ├── adapter.go              # FacebookBrowserAdapter implementing FacebookBrowserPort
│       │   ├── auth.go                 # Login status detection & interactive login flow
│       │   └── extractor.go            # In-browser JS extractor & DOM element parsing
│       └── storage/
│           └── mongo_repo.go           # MongoPostRepository implementing PostRepository port
├── pkg/
│   └── scraper/                        # Reusable Selenium Chrome Automation Engine
├── profiles/                           # Persistent Chrome User Data Directory
└── output/                             # Export artifacts
```

---

## 3. Domain Model (`internal/domain/post`)

### Aggregate & Value Objects (`entity.go`)

```go
type Post struct {
	ID           string    `json:"id" bson:"id"`
	GroupSlug    string    `json:"group_slug" bson:"group_slug"`
	URL          string    `json:"url" bson:"url"`
	AuthorName   string    `json:"author_name" bson:"author_name"`
	AuthorURL    string    `json:"author_url" bson:"author_url"`
	PostedAt     time.Time `json:"posted_at" bson:"posted_at"`
	RawTimestamp string    `json:"raw_timestamp" bson:"raw_timestamp"`
	Text         string    `json:"text" bson:"text"`
	PriceRaw     string    `json:"price_raw" bson:"price_raw"`
	PriceNumeric int64     `json:"price_numeric" bson:"price_numeric"`
	ImageURLs    []string  `json:"image_urls" bson:"image_urls"`
	ScrapedAt    time.Time `json:"scraped_at" bson:"scraped_at"`
}
```

### Driven Port (`repository.go`)

```go
type Repository interface {
	UpsertMany(ctx context.Context, posts []Post) (int, error)
	GetLatestPostID(ctx context.Context, groupSlug string) (string, error)
	FindByGroup(ctx context.Context, groupSlug string, limit int) ([]Post, error)
}
```

### Domain Logic (`service.go`)
- `ParseTimestamp(raw string, now time.Time) (time.Time, bool)`: Parses Indonesian/English relative times ("5 jam", "2 hari", "Kemarin", "24 Agustus").
- `ParsePrice(raw string) (int64, string)`: Normalizes "Rp 125jt", "125.000.000", "125 juta" into `125000000`.
- `DeduplicateAndSort(posts []Post) []Post`: Deduplicates by Post ID and sorts descending by `PostedAt`.

---

## 4. Application Use Case (`internal/application/scrape`)

### Driven Port (`port.go`)

```go
type FacebookBrowserPort interface {
	CheckLoggedIn(ctx context.Context) (bool, error)
	InteractiveLogin(ctx context.Context) error
	ScrapeGroupFeed(ctx context.Context, req FeedRequest) ([]domain.Post, error)
}

type FeedRequest struct {
	GroupURL   string
	MaxDaysAgo int
	UntilID    string
	MaxScrolls int
	Pause      time.Duration
}
```

### Use Case Handler (`handler.go`)
1. Resolves `UntilID` from MongoDB `repo.GetLatestPostID(ctx, groupSlug)`.
2. Validates session with `browser.CheckLoggedIn(ctx)`. If false, triggers `browser.InteractiveLogin(ctx)`.
3. Calls `browser.ScrapeGroupFeed(ctx, req)`.
4. Normalizes and deduplicates posts via domain service.
5. Saves new/updated posts via `repo.UpsertMany(ctx, posts)`.
6. Returns `ScrapeGroupResult` with counts and statistics.

---

## 5. Infrastructure Adapters

### 1. Browser Adapter (`internal/infra/browser`)
- Uses `pkg/scraper` with `--user-data-dir` profile persistence.
- Injects JS evaluation script extracting feed units (`div[role="feed"] > div`, `div[role="article"]`).
- Evaluates stopping condition after each scroll:
  - Stop if any extracted post has `PostedAt < (now - maxDaysAgo)`.
  - Stop if any extracted post matches `UntilID`.
  - Stop if scroll count reaches `MaxScrolls`.

### 2. MongoDB Adapter (`internal/infra/storage/mongo_repo.go`)
- Implements `domain.Repository` with `go.mongodb.org/mongo-driver/v2`.
- Creates compound unique index: `{ "group_slug": 1, "id": 1 }`.
- Creates query index: `{ "group_slug": 1, "posted_at": -1 }`.
- Executes bulk upserts via `mongo.NewUpdateOneModel().SetFilter(...).SetUpdate(...).SetUpsert(true)`.

---

## 6. CLI Driver (`cmd/scrape_group/main.go`)

### Flags:
- `-url`: Target Facebook group URL (default: `https://www.facebook.com/groups/jualbelimobilbekasbali`)
- `-profile`: Chrome user data directory (default: `./profiles/facebook`)
- `-mongo-uri`: MongoDB connection string (default: `mongodb://localhost:27017`)
- `-mongo-db`: MongoDB database name (default: `fb_scraper`)
- `-days`: Maximum days of posts to scrape (default: `2`)
- `-max-scrolls`: Maximum feed scroll iterations (default: `50`)
- `-pause-sec`: Seconds between scroll actions (default: `2`)
- `-login`: Launch interactive visible browser to perform one-time login
- `-no-headless`: Run Chrome with visible UI
- `-export-json`: Also export scraped posts to JSON file in `./output/`

---

## 7. Verification & Testing Strategy

1. **Domain Unit Tests**:
   - `service_test.go`: Test timestamp parsing across Indonesian & English variants, price extraction, and post deduplication.
2. **Repository Unit / Integration Tests**:
   - `mongo_repo_test.go`: Test index initialization and bulk upsert operations.
3. **Browser Adapter Unit Tests**:
   - `extractor_test.go`: Test parsing mock DOM payloads and JS return structures.
4. **End-to-End CLI Verification**:
   - `go test -v ./...` and `go vet ./...` in `/Users/ananda/gitrepo/fb-group-scrape`.
