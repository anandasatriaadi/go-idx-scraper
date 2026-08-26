# `fb-group-scrape` Facebook Group Car Listing Scraper Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement a Hexagonal DDD scraper for Facebook group `https://www.facebook.com/groups/jualbelimobilbekasbali` with interactive auth, 2-day chronological scrolling cutoff / known ID incremental stop, and MongoDB v2 driver persistence.

**Architecture:** Hexagonal Architecture + DDD Tactical Patterns. Domain layer (`internal/domain/post`), Application layer (`internal/application/scrape`), Infrastructure adapters (`internal/infra/browser`, `internal/infra/storage/mongo_repo.go`), and CLI driver (`cmd/scrape_group/main.go`).

**Tech Stack:** Go 1.24+, `go.mongodb.org/mongo-driver/v2`, `github.com/tebeka/selenium v0.9.9`, `pkg/scraper`.

## Global Constraints

- Target directory: `/Users/ananda/gitrepo/fb-group-scrape`
- Module: `github.com/anandasatriaadi/fb-group-scrape`
- MongoDB driver: `go.mongodb.org/mongo-driver/v2`
- Headed mode: Support running in headed mode (`headless: false`) for interactive login and visual live site debugging.
- Chronological sorting: Enforce `?sorting_setting=CHRONOLOGICAL`.
- Stop conditions: Stop when a post $\ge 2$ days ago is reached OR an already-known post ID is found in MongoDB.
- Safe Golang principles: maximum 70 lines per function, bounded loops with `ctx.Done()`, error wrapping with `%w`, zero unhandled errors.

---

### Task 1: Add MongoDB v2 Driver & Implement Domain Layer (`internal/domain/post`)

**Files:**
- Modify: `/Users/ananda/gitrepo/fb-group-scrape/go.mod`
- Create: `/Users/ananda/gitrepo/fb-group-scrape/internal/domain/post/entity.go`
- Create: `/Users/ananda/gitrepo/fb-group-scrape/internal/domain/post/repository.go`
- Create: `/Users/ananda/gitrepo/fb-group-scrape/internal/domain/post/service.go`
- Create: `/Users/ananda/gitrepo/fb-group-scrape/internal/domain/post/service_test.go`

**Interfaces:**
- Produces: `Post` entity, `Repository` port interface, `ParseTimestamp(raw string, now time.Time) (time.Time, bool)`, `ParsePrice(raw string) (int64, string)`, `DeduplicateAndSort(posts []Post) []Post`.

- [ ] **Step 1: Add `go.mongodb.org/mongo-driver/v2` to `go.mod`**

```bash
cd /Users/ananda/gitrepo/fb-group-scrape
go get go.mongodb.org/mongo-driver/v2/mongo@v2.0.1
go mod tidy
```

- [ ] **Step 2: Write failing unit tests for Domain Service (`service_test.go`)**

Create `/Users/ananda/gitrepo/fb-group-scrape/internal/domain/post/service_test.go`:
```go
package post

import (
	"testing"
	"time"
)

func TestParseTimestamp(t *testing.T) {
	refTime := time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		input       string
		expectedAge time.Duration
		ok          bool
	}{
		{"Baru saja", 0, true},
		{"Just now", 0, true},
		{"10 menit", 10 * time.Minute, true},
		{"10 mins", 10 * time.Minute, true},
		{"2 jam", 2 * time.Hour, true},
		{"3 hrs", 3 * time.Hour, true},
		{"Kemarin pukul 10:00", 24 * time.Hour, true},
		{"Yesterday at 10:00", 24 * time.Hour, true},
		{"1 hari", 24 * time.Hour, true},
		{"2 hari", 48 * time.Hour, true},
		{"3 days", 72 * time.Hour, true},
		{"invalid format", 0, false},
	}

	for _, tt := range tests {
		parsed, ok := ParseTimestamp(tt.input, refTime)
		if ok != tt.ok {
			t.Errorf("ParseTimestamp(%q) ok = %v, expected %v", tt.input, ok, tt.ok)
			continue
		}
		if ok {
			diff := refTime.Sub(parsed)
			// Tolerance of 1 minute
			if diff < tt.expectedAge-time.Minute || diff > tt.expectedAge+time.Minute {
				t.Errorf("ParseTimestamp(%q) age diff = %v, expected ~%v", tt.input, diff, tt.expectedAge)
			}
		}
	}
}

func TestParsePrice(t *testing.T) {
	tests := []struct {
		input       string
		expectedNum int64
		expectedRaw string
	}{
		{"Rp 125.000.000", 125000000, "Rp 125.000.000"},
		{"Rp150jt", 150000000, "Rp150jt"},
		{"85 juta", 85000000, "85 juta"},
		{"Free", 0, "Free"},
	}

	for _, tt := range tests {
		num, raw := ParsePrice(tt.input)
		if num != tt.expectedNum || raw != tt.expectedRaw {
			t.Errorf("ParsePrice(%q) = (%d, %q), expected (%d, %q)", tt.input, num, raw, tt.expectedNum, tt.expectedRaw)
		}
	}
}

func TestDeduplicateAndSort(t *testing.T) {
	t1 := time.Date(2026, 8, 26, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 8, 26, 11, 0, 0, 0, time.UTC)
	t3 := time.Date(2026, 8, 26, 9, 0, 0, 0, time.UTC)

	posts := []Post{
		{ID: "1", Text: "Post 1", PostedAt: t1},
		{ID: "2", Text: "Post 2", PostedAt: t2},
		{ID: "1", Text: "Post 1 Duplicate", PostedAt: t1},
		{ID: "3", Text: "Post 3", PostedAt: t3},
	}

	res := DeduplicateAndSort(posts)
	if len(res) != 3 {
		t.Fatalf("expected 3 unique posts, got %d", len(res))
	}
	if res[0].ID != "2" || res[1].ID != "1" || res[2].ID != "3" {
		t.Errorf("expected sorted descending [2, 1, 3], got [%s, %s, %s]", res[0].ID, res[1].ID, res[2].ID)
	}
}
```

- [ ] **Step 3: Run test to verify it fails**

Run: `cd /Users/ananda/gitrepo/fb-group-scrape && go test ./internal/domain/post -v`
Expected: FAIL (undefined: Post, ParseTimestamp, ParsePrice, DeduplicateAndSort)

- [ ] **Step 4: Implement `internal/domain/post/entity.go`**

Create `/Users/ananda/gitrepo/fb-group-scrape/internal/domain/post/entity.go`:
```go
package post

import "time"

// Post represents a Facebook group car listing aggregate root.
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

- [ ] **Step 5: Implement `internal/domain/post/repository.go`**

Create `/Users/ananda/gitrepo/fb-group-scrape/internal/domain/post/repository.go`:
```go
package post

import "context"

// Repository is the driven port for persisting and querying Post entities.
type Repository interface {
	UpsertMany(ctx context.Context, posts []Post) (int, error)
	GetLatestPostID(ctx context.Context, groupSlug string) (string, error)
	FindByGroup(ctx context.Context, groupSlug string, limit int) ([]Post, error)
}
```

- [ ] **Step 6: Implement `internal/domain/post/service.go`**

Create `/Users/ananda/gitrepo/fb-group-scrape/internal/domain/post/service.go`:
```go
package post

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	reDigits   = regexp.MustCompile(`[0-9]+`)
	reJtJuta   = regexp.MustCompile(`(?i)([0-9]+(?:\.[0-9]+)?)\s*(?:jt|juta)`)
	reCleanNum = regexp.MustCompile(`[^0-9]`)
)

// ParseTimestamp parses relative or absolute Facebook timestamps into a UTC time.Time.
func ParseTimestamp(raw string, now time.Time) (time.Time, bool) {
	s := strings.TrimSpace(strings.ToLower(raw))
	if s == "" {
		return time.Time{}, false
	}
	if s == "baru saja" || s == "just now" {
		return now, true
	}
	if d, ok := matchRelativeTime(s); ok {
		return now.Add(-d), true
	}
	return time.Time{}, false
}

func matchRelativeTime(s string) (time.Duration, bool) {
	if strings.Contains(s, "kemarin") || strings.Contains(s, "yesterday") {
		return 24 * time.Hour, true
	}
	digits := reDigits.FindString(s)
	if digits == "" {
		return 0, false
	}
	n, err := strconv.Atoi(digits)
	if err != nil {
		return 0, false
	}
	if strings.Contains(s, "menit") || strings.Contains(s, "min") {
		return time.Duration(n) * time.Minute, true
	}
	if strings.Contains(s, "jam") || strings.Contains(s, "hr") || strings.Contains(s, "hour") {
		return time.Duration(n) * time.Hour, true
	}
	if strings.Contains(s, "hari") || strings.Contains(s, "day") {
		return time.Duration(n) * 24 * time.Hour, true
	}
	return 0, false
}

// ParsePrice extracts numeric price value and raw string.
func ParsePrice(raw string) (int64, string) {
	rawTrimmed := strings.TrimSpace(raw)
	if m := reJtJuta.FindStringSubmatch(rawTrimmed); len(m) > 1 {
		if f, err := strconv.ParseFloat(m[1], 64); err == nil {
			return int64(f * 1000000), rawTrimmed
		}
	}
	cleaned := reCleanNum.ReplaceAllString(rawTrimmed, "")
	if cleaned != "" {
		if num, err := strconv.ParseInt(cleaned, 10, 64); err == nil {
			return num, rawTrimmed
		}
	}
	return 0, rawTrimmed
}

// DeduplicateAndSort deduplicates posts by ID and sorts descending by PostedAt.
func DeduplicateAndSort(posts []Post) []Post {
	seen := make(map[string]bool, len(posts))
	unique := make([]Post, 0, len(posts))

	for _, p := range posts {
		if p.ID == "" || seen[p.ID] {
			continue
		}
		seen[p.ID] = true
		unique = append(unique, p)
	}

	sort.Slice(unique, func(i, j int) bool {
		return unique[i].PostedAt.After(unique[j].PostedAt)
	})

	return unique
}
```

- [ ] **Step 7: Run domain tests to verify they pass**

Run: `cd /Users/ananda/gitrepo/fb-group-scrape && go test ./internal/domain/post -v`
Expected: PASS

- [ ] **Step 8: Commit domain layer**

```bash
cd /Users/ananda/gitrepo/fb-group-scrape
git add go.mod go.sum internal/domain/post/
git commit -m "feat(domain): implement Post entity, repository port, and domain service"
```

---

### Task 2: Implement Application Layer Use Case (`internal/application/scrape`)

**Files:**
- Create: `/Users/ananda/gitrepo/fb-group-scrape/internal/application/scrape/port.go`
- Create: `/Users/ananda/gitrepo/fb-group-scrape/internal/application/scrape/command.go`
- Create: `/Users/ananda/gitrepo/fb-group-scrape/internal/application/scrape/handler.go`
- Create: `/Users/ananda/gitrepo/fb-group-scrape/internal/application/scrape/handler_test.go`

**Interfaces:**
- Produces: `FacebookBrowserPort` interface, `ScrapeGroupCommand`, `ScrapeGroupResult`, `Handler` use case.

- [ ] **Step 1: Write tests for Application Handler (`handler_test.go`)**

Create `/Users/ananda/gitrepo/fb-group-scrape/internal/application/scrape/handler_test.go`:
```go
package scrape

import (
	"context"
	"testing"
	"time"

	"github.com/anandasatriaadi/fb-group-scrape/internal/domain/post"
)

type mockBrowserPort struct {
	checkLoggedInFunc    func(ctx context.Context) (bool, error)
	interactiveLoginFunc func(ctx context.Context) error
	scrapeFeedFunc       func(ctx context.Context, req FeedRequest) ([]post.Post, error)
}

func (m *mockBrowserPort) CheckLoggedIn(ctx context.Context) (bool, error) {
	if m.checkLoggedInFunc != nil {
		return m.checkLoggedInFunc(ctx)
	}
	return true, nil
}

func (m *mockBrowserPort) InteractiveLogin(ctx context.Context) error {
	if m.interactiveLoginFunc != nil {
		return m.interactiveLoginFunc(ctx)
	}
	return nil
}

func (m *mockBrowserPort) ScrapeGroupFeed(ctx context.Context, req FeedRequest) ([]post.Post, error) {
	if m.scrapeFeedFunc != nil {
		return m.scrapeFeedFunc(ctx, req)
	}
	return []post.Post{
		{ID: "p1", GroupSlug: "jualbelimobilbekasbali", Text: "Mobil 1", PostedAt: time.Now()},
	}, nil
}

type mockPostRepo struct {
	upsertManyFunc      func(ctx context.Context, posts []post.Post) (int, error)
	getLatestPostIDFunc func(ctx context.Context, groupSlug string) (string, error)
	findByGroupFunc     func(ctx context.Context, groupSlug string, limit int) ([]post.Post, error)
}

func (m *mockPostRepo) UpsertMany(ctx context.Context, posts []post.Post) (int, error) {
	if m.upsertManyFunc != nil {
		return m.upsertManyFunc(ctx, posts)
	}
	return len(posts), nil
}

func (m *mockPostRepo) GetLatestPostID(ctx context.Context, groupSlug string) (string, error) {
	if m.getLatestPostIDFunc != nil {
		return m.getLatestPostIDFunc(ctx, groupSlug)
	}
	return "known-id-123", nil
}

func (m *mockPostRepo) FindByGroup(ctx context.Context, groupSlug string, limit int) ([]post.Post, error) {
	if m.findByGroupFunc != nil {
		return m.findByGroupFunc(ctx, groupSlug, limit)
	}
	return nil, nil
}

func TestHandler_Execute(t *testing.T) {
	browser := &mockBrowserPort{}
	repo := &mockPostRepo{}
	h := NewHandler(browser, repo)

	cmd := ScrapeGroupCommand{
		GroupURL:   "https://www.facebook.com/groups/jualbelimobilbekasbali",
		GroupSlug:  "jualbelimobilbekasbali",
		MaxDaysAgo: 2,
		MaxScrolls: 10,
		Pause:      1 * time.Second,
	}

	res, err := h.Execute(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if res.ScrapedCount != 1 || res.SavedCount != 1 {
		t.Errorf("expected 1 scraped & saved, got scraped=%d, saved=%d", res.ScrapedCount, res.SavedCount)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/ananda/gitrepo/fb-group-scrape && go test ./internal/application/scrape -v`
Expected: FAIL (undefined: FacebookBrowserPort, ScrapeGroupCommand, NewHandler, etc.)

- [ ] **Step 3: Implement `internal/application/scrape/port.go`**

Create `/Users/ananda/gitrepo/fb-group-scrape/internal/application/scrape/port.go`:
```go
package scrape

import (
	"context"
	"time"

	"github.com/anandasatriaadi/fb-group-scrape/internal/domain/post"
)

// FeedRequest specifies parameters for the browser scrape feed operation.
type FeedRequest struct {
	GroupURL   string
	GroupSlug  string
	MaxDaysAgo int
	UntilID    string
	MaxScrolls int
	Pause      time.Duration
}

// FacebookBrowserPort is the driven port interface for browser automation.
type FacebookBrowserPort interface {
	CheckLoggedIn(ctx context.Context) (bool, error)
	InteractiveLogin(ctx context.Context) error
	ScrapeGroupFeed(ctx context.Context, req FeedRequest) ([]post.Post, error)
}
```

- [ ] **Step 4: Implement `internal/application/scrape/command.go`**

Create `/Users/ananda/gitrepo/fb-group-scrape/internal/application/scrape/command.go`:
```go
package scrape

import (
	"time"

	"github.com/anandasatriaadi/fb-group-scrape/internal/domain/post"
)

// ScrapeGroupCommand defines inputs for the scrape group use case.
type ScrapeGroupCommand struct {
	GroupURL   string
	GroupSlug  string
	MaxDaysAgo int
	MaxScrolls int
	Pause      time.Duration
	ForceLogin bool
}

// ScrapeGroupResult summarizes the output of the use case execution.
type ScrapeGroupResult struct {
	GroupSlug    string
	ScrapedCount int
	SavedCount   int
	Posts        []post.Post
	Duration     time.Duration
}
```

- [ ] **Step 5: Implement `internal/application/scrape/handler.go`**

Create `/Users/ananda/gitrepo/fb-group-scrape/internal/application/scrape/handler.go`:
```go
package scrape

import (
	"context"
	"fmt"
	"time"

	"github.com/anandasatriaadi/fb-group-scrape/internal/domain/post"
)

// Handler executes the ScrapeGroupCommand use case.
type Handler struct {
	browser FacebookBrowserPort
	repo    post.Repository
}

// NewHandler creates a new ScrapeGroupHandler instance.
func NewHandler(browser FacebookBrowserPort, repo post.Repository) *Handler {
	return &Handler{
		browser: browser,
		repo:    repo,
	}
}

// Execute orchestrates the full group scraping pipeline.
func (h *Handler) Execute(ctx context.Context, cmd ScrapeGroupCommand) (*ScrapeGroupResult, error) {
	start := time.Now()

	if cmd.ForceLogin {
		if err := h.browser.InteractiveLogin(ctx); err != nil {
			return nil, fmt.Errorf("interactive login failed: %w", err)
		}
	} else {
		isLoggedIn, err := h.browser.CheckLoggedIn(ctx)
		if err != nil {
			return nil, fmt.Errorf("checking login status: %w", err)
		}
		if !isLoggedIn {
			if err := h.browser.InteractiveLogin(ctx); err != nil {
				return nil, fmt.Errorf("interactive login failed: %w", err)
			}
		}
	}

	untilID, err := h.repo.GetLatestPostID(ctx, cmd.GroupSlug)
	if err != nil {
		untilID = ""
	}

	req := FeedRequest{
		GroupURL:   cmd.GroupURL,
		GroupSlug:  cmd.GroupSlug,
		MaxDaysAgo: cmd.MaxDaysAgo,
		UntilID:    untilID,
		MaxScrolls: cmd.MaxScrolls,
		Pause:      cmd.Pause,
	}

	rawPosts, err := h.browser.ScrapeGroupFeed(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("scraping group feed: %w", err)
	}

	posts := post.DeduplicateAndSort(rawPosts)
	savedCount := 0
	if len(posts) > 0 {
		saved, err := h.repo.UpsertMany(ctx, posts)
		if err != nil {
			return nil, fmt.Errorf("persisting scraped posts: %w", err)
		}
		savedCount = saved
	}

	return &ScrapeGroupResult{
		GroupSlug:    cmd.GroupSlug,
		ScrapedCount: len(posts),
		SavedCount:   savedCount,
		Posts:        posts,
		Duration:     time.Since(start),
	}, nil
}
```

- [ ] **Step 6: Run application tests to verify they pass**

Run: `cd /Users/ananda/gitrepo/fb-group-scrape && go test ./internal/application/scrape -v`
Expected: PASS

- [ ] **Step 7: Commit application layer**

```bash
cd /Users/ananda/gitrepo/fb-group-scrape
git add internal/application/scrape/
git commit -m "feat(application): implement scrape group use case and browser port"
```

---

### Task 3: Implement MongoDB v2 Repository Adapter (`internal/infra/storage`)

**Files:**
- Create: `/Users/ananda/gitrepo/fb-group-scrape/internal/infra/storage/mongo_repo.go`
- Create: `/Users/ananda/gitrepo/fb-group-scrape/internal/infra/storage/mongo_repo_test.go`

**Interfaces:**
- Produces: `MongoPostRepository` implementing `post.Repository` port.

- [ ] **Step 1: Implement `internal/infra/storage/mongo_repo.go`**

Create `/Users/ananda/gitrepo/fb-group-scrape/internal/infra/storage/mongo_repo.go`:
```go
package storage

import (
	"context"
	"fmt"

	"github.com/anandasatriaadi/fb-group-scrape/internal/domain/post"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const CollectionPosts = "facebook_group_posts"

// MongoPostRepository implements the post.Repository port using MongoDB v2.
type MongoPostRepository struct {
	collection *mongo.Collection
}

// NewMongoPostRepository creates a repository and ensures unique compound indexes.
func NewMongoPostRepository(ctx context.Context, db *mongo.Database) (*MongoPostRepository, error) {
	coll := db.Collection(CollectionPosts)
	repo := &MongoPostRepository{collection: coll}
	if err := repo.initIndexes(ctx); err != nil {
		return nil, fmt.Errorf("initializing mongo indexes: %w", err)
	}
	return repo, nil
}

func (r *MongoPostRepository) initIndexes(ctx context.Context) error {
	models := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "group_slug", Value: 1}, {Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "group_slug", Value: 1}, {Key: "posted_at", Value: -1}},
		},
	}
	_, err := r.collection.Indexes().CreateMany(ctx, models)
	if err != nil {
		return fmt.Errorf("creating indexes: %w", err)
	}
	return nil
}

// UpsertMany bulk upserts posts into MongoDB.
func (r *MongoPostRepository) UpsertMany(ctx context.Context, posts []post.Post) (int, error) {
	if len(posts) == 0 {
		return 0, nil
	}
	writes := make([]mongo.WriteModel, 0, len(posts))
	for _, p := range posts {
		filter := bson.D{{Key: "group_slug", Value: p.GroupSlug}, {Key: "id", Value: p.ID}}
		update := bson.D{{Key: "$set", Value: p}}
		model := mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(true)
		writes = append(writes, model)
	}

	opts := options.BulkWrite().SetOrdered(false)
	res, err := r.collection.BulkWrite(ctx, writes, opts)
	if err != nil {
		return 0, fmt.Errorf("bulk writing posts: %w", err)
	}
	return int(res.UpsertedCount + res.ModifiedCount + res.MatchedCount), nil
}

// GetLatestPostID returns the most recently posted post ID for a group slug.
func (r *MongoPostRepository) GetLatestPostID(ctx context.Context, groupSlug string) (string, error) {
	filter := bson.D{{Key: "group_slug", Value: groupSlug}}
	opts := options.FindOne().SetSort(bson.D{{Key: "posted_at", Value: -1}}).SetProjection(bson.D{{Key: "id", Value: 1}})

	var p post.Post
	if err := r.collection.FindOne(ctx, filter, opts).Decode(&p); err != nil {
		return "", fmt.Errorf("finding latest post: %w", err)
	}
	return p.ID, nil
}

// FindByGroup queries posts for a group ordered from newest to oldest.
func (r *MongoPostRepository) FindByGroup(ctx context.Context, groupSlug string, limit int) ([]post.Post, error) {
	filter := bson.D{{Key: "group_slug", Value: groupSlug}}
	opts := options.Find().SetSort(bson.D{{Key: "posted_at", Value: -1}}).SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("finding posts: %w", err)
	}
	defer cursor.Close(ctx)

	var posts []post.Post
	if err := cursor.All(ctx, &posts); err != nil {
		return nil, fmt.Errorf("decoding posts: %w", err)
	}
	return posts, nil
}
```

- [ ] **Step 2: Write unit test verifying struct signatures & index models (`mongo_repo_test.go`)**

Create `/Users/ananda/gitrepo/fb-group-scrape/internal/infra/storage/mongo_repo_test.go`:
```go
package storage

import (
	"testing"

	"github.com/anandasatriaadi/fb-group-scrape/internal/domain/post"
)

func TestMongoPostRepository_InterfaceCompliance(t *testing.T) {
	var _ post.Repository = (*MongoPostRepository)(nil)
}
```

- [ ] **Step 3: Run test to verify compilation & interface compliance**

Run: `cd /Users/ananda/gitrepo/fb-group-scrape && go test ./internal/infra/storage -v`
Expected: PASS

- [ ] **Step 4: Commit MongoDB adapter**

```bash
cd /Users/ananda/gitrepo/fb-group-scrape
git add internal/infra/storage/
git commit -m "feat(infra): implement MongoPostRepository adapter with MongoDB v2"
```

---

### Task 4: Implement Browser Adapter, Interactive Auth & DOM Extractor (`internal/infra/browser`)

**Files:**
- Create: `/Users/ananda/gitrepo/fb-group-scrape/internal/infra/browser/auth.go`
- Create: `/Users/ananda/gitrepo/fb-group-scrape/internal/infra/browser/extractor.go`
- Create: `/Users/ananda/gitrepo/fb-group-scrape/internal/infra/browser/adapter.go`
- Create: `/Users/ananda/gitrepo/fb-group-scrape/internal/infra/browser/extractor_test.go`

**Interfaces:**
- Produces: `FacebookBrowserAdapter` implementing `scrape.FacebookBrowserPort`.

- [ ] **Step 1: Implement `internal/infra/browser/extractor.go`**

Create `/Users/ananda/gitrepo/fb-group-scrape/internal/infra/browser/extractor.go`:
```go
package browser

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/anandasatriaadi/fb-group-scrape/internal/domain/post"
	"github.com/anandasatriaadi/fb-group-scrape/pkg/scraper"
)

const extractionScript = `
return (function() {
    var articles = document.querySelectorAll('div[role="feed"] > div, div[role="article"], div[data-pagelet*="FeedUnit"]');
    var results = [];
    for (var i = 0; i < articles.length; i++) {
        var el = articles[i];
        var text = el.innerText || "";
        if (text.trim().length < 10) continue;

        var links = el.querySelectorAll('a');
        var postUrl = "";
        var authorName = "";
        var authorUrl = "";
        var timestamp = "";

        for (var j = 0; j < links.length; j++) {
            var href = links[j].href || "";
            if (!postUrl && (href.indexOf('/posts/') !== -1 || href.indexOf('story_fbid') !== -1 || href.indexOf('/permalink/') !== -1)) {
                postUrl = href;
                timestamp = links[j].innerText || "";
            }
            if (!authorName && href.indexOf('facebook.com/') !== -1 && href.indexOf('/groups/') === -1 && links[j].innerText.trim().length > 1) {
                authorName = links[j].innerText.trim();
                authorUrl = href;
            }
        }

        var images = [];
        var imgTags = el.querySelectorAll('img');
        for (var k = 0; k < imgTags.length; k++) {
            var src = imgTags[k].src || "";
            if (src.indexOf('scontent') !== -1) {
                images.push(src);
            }
        }

        var idMatch = postUrl.match(/\/posts\/([0-9]+)/) || postUrl.match(/story_fbid=([0-9]+)/) || postUrl.match(/\/permalink\/([0-9]+)/);
        var postId = idMatch ? idMatch[1] : "";

        results.push({
            id: postId,
            url: postUrl,
            author_name: authorName,
            author_url: authorUrl,
            timestamp: timestamp,
            text: text,
            images: images
        });
    }
    return JSON.stringify(results);
})();
`

type rawPostDTO struct {
	ID         string   `json:"id"`
	URL        string   `json:"url"`
	AuthorName string   `json:"author_name"`
	AuthorURL  string   `json:"author_url"`
	Timestamp  string   `json:"timestamp"`
	Text       string   `json:"text"`
	Images     []string `json:"images"`
}

// ExtractVisiblePosts runs the JavaScript extraction script on the active page.
func ExtractVisiblePosts(ctx context.Context, s *scraper.Scraper, groupSlug string, now time.Time) ([]post.Post, error) {
	val, err := s.ExecuteScript(ctx, extractionScript)
	if err != nil {
		return nil, fmt.Errorf("evaluating extraction script: %w", err)
	}
	jsonStr, ok := val.(string)
	if !ok || jsonStr == "" {
		return nil, nil
	}

	var rawList []rawPostDTO
	if err := json.Unmarshal([]byte(jsonStr), &rawList); err != nil {
		return nil, fmt.Errorf("unmarshaling raw post DTOs: %w", err)
	}

	posts := make([]post.Post, 0, len(rawList))
	for _, r := range rawList {
		if r.ID == "" && r.URL == "" {
			continue
		}
		id := r.ID
		if id == "" {
			id = r.URL
		}
		postedAt, _ := post.ParseTimestamp(r.Timestamp, now)
		if postedAt.IsZero() {
			postedAt = now
		}
		priceNum, priceRaw := post.ParsePrice(r.Text)

		p := post.Post{
			ID:           id,
			GroupSlug:    groupSlug,
			URL:          r.URL,
			AuthorName:   r.AuthorName,
			AuthorURL:    r.AuthorURL,
			PostedAt:     postedAt,
			RawTimestamp: r.Timestamp,
			Text:         r.Text,
			PriceRaw:     priceRaw,
			PriceNumeric: priceNum,
			ImageURLs:    r.Images,
			ScrapedAt:    now,
		}
		posts = append(posts, p)
	}
	return posts, nil
}
```

- [ ] **Step 2: Implement `internal/infra/browser/auth.go`**

Create `/Users/ananda/gitrepo/fb-group-scrape/internal/infra/browser/auth.go`:
```go
package browser

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/anandasatriaadi/fb-group-scrape/pkg/scraper"
)

// CheckLoggedIn checks if the active browser session has a valid Facebook session cookie.
func CheckLoggedIn(ctx context.Context, s *scraper.Scraper) (bool, error) {
	cookies, err := s.GetCookies(ctx)
	if err != nil {
		return false, fmt.Errorf("retrieving session cookies: %w", err)
	}
	for _, c := range cookies {
		if c.Name == "c_user" && c.Value != "" {
			return true, nil
		}
	}
	return false, nil
}

// InteractiveLogin spawns a visible Chrome window and prompts the user to log in.
func InteractiveLogin(ctx context.Context, profilePath, chromeDriverPath string) error {
	cfg := scraper.DefaultConfig()
	cfg.Headless = false
	cfg.ChromeProfilePath = profilePath
	cfg.ChromeDriverPath = chromeDriverPath

	s, err := scraper.New(cfg)
	if err != nil {
		return fmt.Errorf("launching visible browser for login: %w", err)
	}
	defer s.Close()

	if err := s.Navigate(ctx, "https://www.facebook.com/login"); err != nil {
		return fmt.Errorf("navigating to facebook login: %w", err)
	}

	fmt.Println("================================================================================")
	fmt.Println("[ACTION REQUIRED] Interactive Facebook Login")
	fmt.Printf("1. Visible Chrome has opened with profile at: %s\n", profilePath)
	fmt.Println("2. Log in with your Facebook credentials in the browser window.")
	fmt.Println("3. Complete 2FA / checkpoints if prompted until your feed is visible.")
	fmt.Println("================================================================================")
	fmt.Print("Press [Enter] when login is complete in the browser: ")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()

	isLoggedIn, _ := CheckLoggedIn(ctx, s)
	if !isLoggedIn {
		title, _ := s.GetTitle(ctx)
		if strings.Contains(strings.ToLower(title), "login") {
			return fmt.Errorf("session cookie c_user not found; please verify login completed")
		}
	}
	fmt.Println("[SUCCESS] Facebook session detected and saved to profile!")
	time.Sleep(1 * time.Second)
	return nil
}
```

- [ ] **Step 3: Implement `internal/infra/browser/adapter.go`**

Create `/Users/ananda/gitrepo/fb-group-scrape/internal/infra/browser/adapter.go`:
```go
package browser

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/anandasatriaadi/fb-group-scrape/internal/application/scrape"
	"github.com/anandasatriaadi/fb-group-scrape/internal/domain/post"
	"github.com/anandasatriaadi/fb-group-scrape/pkg/scraper"
)

// FacebookBrowserAdapter implements scrape.FacebookBrowserPort.
type FacebookBrowserAdapter struct {
	profilePath      string
	chromeDriverPath string
	headless         bool
}

// NewFacebookBrowserAdapter creates a new browser adapter.
func NewFacebookBrowserAdapter(profilePath, chromeDriverPath string, headless bool) *FacebookBrowserAdapter {
	return &FacebookBrowserAdapter{
		profilePath:      profilePath,
		chromeDriverPath: chromeDriverPath,
		headless:         headless,
	}
}

// CheckLoggedIn checks for active Facebook session.
func (a *FacebookBrowserAdapter) CheckLoggedIn(ctx context.Context) (bool, error) {
	cfg := scraper.DefaultConfig()
	cfg.Headless = true
	cfg.ChromeProfilePath = a.profilePath
	cfg.ChromeDriverPath = a.chromeDriverPath

	s, err := scraper.New(cfg)
	if err != nil {
		return false, fmt.Errorf("starting scraper for auth check: %w", err)
	}
	defer s.Close()

	return CheckLoggedIn(ctx, s)
}

// InteractiveLogin runs interactive login flow.
func (a *FacebookBrowserAdapter) InteractiveLogin(ctx context.Context) error {
	return InteractiveLogin(ctx, a.profilePath, a.chromeDriverPath)
}

// ScrapeGroupFeed navigates to group feed with chronological sorting and scrolls until stop condition.
func (a *FacebookBrowserAdapter) ScrapeGroupFeed(ctx context.Context, req scrape.FeedRequest) ([]post.Post, error) {
	cfg := scraper.DefaultConfig()
	cfg.Headless = a.headless
	cfg.ChromeProfilePath = a.profilePath
	cfg.ChromeDriverPath = a.chromeDriverPath

	s, err := scraper.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("starting scraper for group feed: %w", err)
	}
	defer s.Close()

	targetURL := formatChronologicalURL(req.GroupURL)
	if err := s.Navigate(ctx, targetURL); err != nil {
		return nil, fmt.Errorf("navigating to %s: %w", targetURL, err)
	}

	time.Sleep(3 * time.Second)

	now := time.Now().UTC()
	cutoff := now.Add(-time.Duration(req.MaxDaysAgo) * 24 * time.Hour)

	var allPosts []post.Post
	seen := make(map[string]bool)

	for scroll := 0; scroll < req.MaxScrolls; scroll++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		visible, err := ExtractVisiblePosts(ctx, s, req.GroupSlug, now)
		if err != nil {
			return nil, fmt.Errorf("extracting visible posts: %w", err)
		}

		shouldStop := false
		for _, p := range visible {
			if seen[p.ID] {
				continue
			}
			seen[p.ID] = true
			allPosts = append(allPosts, p)

			if req.UntilID != "" && p.ID == req.UntilID {
				shouldStop = true
				break
			}
			if !p.PostedAt.IsZero() && p.PostedAt.Before(cutoff) {
				shouldStop = true
				break
			}
		}

		if shouldStop {
			break
		}

		_ = s.ScrollBy(ctx, 0, 1200)
		time.Sleep(req.Pause)
	}

	return allPosts, nil
}

func formatChronologicalURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	q := u.Query()
	q.Set("sorting_setting", "CHRONOLOGICAL")
	u.RawQuery = q.Encode()
	return u.String()
}
```

- [ ] **Step 4: Write unit tests for extraction DTO parsing & URL formatting (`extractor_test.go`)**

Create `/Users/ananda/gitrepo/fb-group-scrape/internal/infra/browser/extractor_test.go`:
```go
package browser

import (
	"testing"
)

func TestFormatChronologicalURL(t *testing.T) {
	raw := "https://www.facebook.com/groups/jualbelimobilbekasbali"
	formatted := formatChronologicalURL(raw)
	expected := "https://www.facebook.com/groups/jualbelimobilbekasbali?sorting_setting=CHRONOLOGICAL"
	if formatted != expected {
		t.Errorf("formatChronologicalURL(%q) = %q, expected %q", raw, formatted, expected)
	}
}
```

- [ ] **Step 5: Run tests to verify compilation & tests pass**

Run: `cd /Users/ananda/gitrepo/fb-group-scrape && go test ./internal/infra/browser -v`
Expected: PASS

- [ ] **Step 6: Commit browser adapter**

```bash
cd /Users/ananda/gitrepo/fb-group-scrape
git add internal/infra/browser/
git commit -m "feat(infra): implement FacebookBrowserAdapter, auth, and DOM extractor"
```

---

### Task 5: Implement CLI Driver (`cmd/scrape_group/main.go`)

**Files:**
- Create: `/Users/ananda/gitrepo/fb-group-scrape/cmd/scrape_group/main.go`

**Interfaces:**
- Produces: Standalone binary `cmd/scrape_group` executing the scrape use case.

- [ ] **Step 1: Implement `cmd/scrape_group/main.go`**

Create `/Users/ananda/gitrepo/fb-group-scrape/cmd/scrape_group/main.go`:
```go
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/anandasatriaadi/fb-group-scrape/internal/application/scrape"
	"github.com/anandasatriaadi/fb-group-scrape/internal/infra/browser"
	"github.com/anandasatriaadi/fb-group-scrape/internal/infra/storage"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	var (
		groupURL   string
		groupSlug  string
		profileDir string
		driverPath string
		mongoURI   string
		mongoDB    string
		days       int
		maxScrolls int
		pauseSec   int
		loginOnly  bool
		noHeadless bool
		exportJSON bool
	)

	flag.StringVar(&groupURL, "url", "https://www.facebook.com/groups/jualbelimobilbekasbali", "Facebook group URL")
	flag.StringVar(&groupSlug, "slug", "jualbelimobilbekasbali", "Facebook group slug")
	flag.StringVar(&profileDir, "profile", "./profiles/facebook", "Path to Chrome user data profile")
	flag.StringVar(&driverPath, "chromedriver", "", "Explicit chromedriver path (auto-discovered if empty)")
	flag.StringVar(&mongoURI, "mongo-uri", "mongodb://localhost:27017", "MongoDB connection URI")
	flag.StringVar(&mongoDB, "mongo-db", "fb_scraper", "MongoDB database name")
	flag.IntVar(&days, "days", 2, "Stop scrolling when posts older than N days are found")
	flag.IntVar(&maxScrolls, "max-scrolls", 50, "Safety maximum number of scroll iterations")
	flag.IntVar(&pauseSec, "pause-sec", 2, "Seconds to pause between feed scrolls")
	flag.BoolVar(&loginOnly, "login", false, "Launch interactive visible Chrome to perform one-time login")
	flag.BoolVar(&noHeadless, "no-headless", false, "Run Chrome in visible GUI mode")
	flag.BoolVar(&exportJSON, "export-json", true, "Export scraped posts to JSON file in ./output")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Received termination signal, shutting down...")
		cancel()
	}()

	headless := !noHeadless
	if loginOnly {
		headless = false
	}

	browserAdapter := browser.NewFacebookBrowserAdapter(profileDir, driverPath, headless)

	if loginOnly {
		if err := browserAdapter.InteractiveLogin(ctx); err != nil {
			log.Fatalf("Interactive login failed: %v", err)
		}
		return
	}

	clientOpts := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(clientOpts)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	repo, err := storage.NewMongoPostRepository(ctx, client.Database(mongoDB))
	if err != nil {
		log.Fatalf("Failed to initialize Mongo repository: %v", err)
	}

	handler := scrape.NewHandler(browserAdapter, repo)
	cmd := scrape.ScrapeGroupCommand{
		GroupURL:   groupURL,
		GroupSlug:  groupSlug,
		MaxDaysAgo: days,
		MaxScrolls: maxScrolls,
		Pause:      time.Duration(pauseSec) * time.Second,
		ForceLogin: false,
	}

	log.Printf("Starting scrape for group %s (days cutoff: %d, headless: %v)...", groupSlug, days, headless)
	res, err := handler.Execute(ctx, cmd)
	if err != nil {
		log.Fatalf("Scraping execution failed: %v", err)
	}

	log.Printf("Scraping completed in %v! Found: %d posts, Persisted in DB: %d", res.Duration, res.ScrapedCount, res.SavedCount)

	if exportJSON && len(res.Posts) > 0 {
		if err := exportToFile(groupSlug, res.Posts); err != nil {
			log.Printf("Warning: failed to export JSON: %v", err)
		}
	}
}

func exportToFile(groupSlug string, posts any) error {
	if err := os.MkdirAll("output", 0755); err != nil {
		return err
	}
	fileName := filepath.Join("output", fmt.Sprintf("%s_%s.json", groupSlug, time.Now().Format("20060102_150405")))
	data, err := json.MarshalIndent(posts, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(fileName, data, 0644); err != nil {
		return err
	}
	log.Printf("Exported JSON result to %s", fileName)
	return nil
}
```

- [ ] **Step 2: Build CLI binary**

Run: `cd /Users/ananda/gitrepo/fb-group-scrape && go build -o bin/scrape_group ./cmd/scrape_group`
Expected: PASS (binary created in `bin/scrape_group`)

- [ ] **Step 3: Commit CLI runner**

```bash
cd /Users/ananda/gitrepo/fb-group-scrape
git add cmd/scrape_group/
git commit -m "feat(cmd): implement scrape_group CLI entry point with MongoDB v2 integration"
```

---

### Task 6: End-to-End Verification & Documentation Update

**Files:**
- Modify: `/Users/ananda/gitrepo/fb-group-scrape/README.md`
- Create: `/Users/ananda/gitrepo/fb-group-scrape/Makefile`

**Interfaces:**
- Produces: Complete build targets and updated documentation.

- [ ] **Step 1: Create `Makefile` in `fb-group-scrape`**

Create `/Users/ananda/gitrepo/fb-group-scrape/Makefile`:
```makefile
.PHONY: all build test vet login scrape

build:
	go build -o bin/scrape_group ./cmd/scrape_group

test:
	go test -v ./...

vet:
	go vet ./...

login:
	go run ./cmd/scrape_group -login

scrape:
	go run ./cmd/scrape_group -days 2

scrape-headed:
	go run ./cmd/scrape_group -days 2 -no-headless
```

- [ ] **Step 2: Update `README.md` with Facebook Group Scraper instructions**

Update `/Users/ananda/gitrepo/fb-group-scrape/README.md`:
```markdown
# fb-group-scrape

High-performance Facebook web scraper and data collection engine with persistent Chrome profiles, Hexagonal DDD architecture, and MongoDB v2 persistence.

## Architecture

- **Domain Layer (`internal/domain/post`)**: `Post` aggregate root, timestamp & price parsers, post deduplication.
- **Application Layer (`internal/application/scrape`)**: `ScrapeGroupCommand` use case orchestrating browser port and repository port.
- **Infrastructure Layer (`internal/infra`)**:
  - `internal/infra/browser`: Selenium Chrome browser adapter, interactive login, and in-browser JS DOM extractor.
  - `internal/infra/storage`: MongoDB v2 repository with unique indexing and bulk upserts.
- **Core Package (`pkg/scraper`)**: Decoupled, reusable Selenium Chrome automation engine.

## Quick Start

### 1. Interactive One-Time Login (Headed Mode)

```bash
make login
# or: go run ./cmd/scrape_group -login
```
A visible Chrome window will open. Complete your login in the browser window and press `<Enter>` in the terminal when logged in. The session is saved to `./profiles/facebook`.

### 2. Run Headless Group Scraper

```bash
make scrape
# or: go run ./cmd/scrape_group -days 2
```

### 3. Run Headed (Visible) Group Scraper

```bash
make scrape-headed
# or: go run ./cmd/scrape_group -days 2 -no-headless
```

## Running Tests

```bash
make test
make vet
```
```

- [ ] **Step 3: Run full verification suite**

Run: `cd /Users/ananda/gitrepo/fb-group-scrape && go test -v ./... && go vet ./...`
Expected: PASS with 0 warnings, 0 errors.

- [ ] **Step 4: Commit Makefile and README**

```bash
cd /Users/ananda/gitrepo/fb-group-scrape
git add Makefile README.md
git commit -m "docs: add Makefile and update README for Facebook group scraper"
```
