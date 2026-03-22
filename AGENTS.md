# Agent Guide: go-idx-scraper

High-level project context, development standards, and actionable guidance for AI agents working on this repository.

---

## Project Context

**go-idx-scraper** is a Go-based scraper for fetching and processing financial data from the Indonesia Stock Exchange (IDX). It automates the collection of stock issuer lists, financial reports, and news.

**Important:** The API was migrated to `idx-web/` (Nuxt 4 Server API). The Go codebase is now **data collection only** - do not add API/HTTP server code here.

## Repository Structure

```
go-idx-scraper/          # Go scraper (data collection only)
├── cmd/                  # CLI entry points (downloader, scraper, issuer, announcement)
├── internal/
│   ├── feature/          # Business logic (announcement, news, finreport, stock, system)
│   ├── infra/            # Adapters (MongoDB, scrapers, IDX adapter, browser)
│   ├── browser/          # Selenium browser setup
│   ├── config/           # Configuration loading
│   └── helper/          # Utilities (email, excel, file, logger)
├── config/               # config.yml
├── tools/                # Code generators
└── idx-web/             # Nuxt 4 API server (SEPARATE PROJECT)

idx-web/                 # Vue.js/Nuxt 4 API server
├── src/
│   ├── server/
│   │   ├── api/v1/      # API endpoints
│   │   ├── middleware/   # CORS, auth
│   │   ├── plugins/      # MongoDB plugin
│   │   └── utils/        # Repos, types, Firebase admin
│   └── app.vue
└── nuxt.config.ts
```

---

## CORE RULES

### Must Follow

1. **Propagate context.Context** - Never use `context.Background()` in business logic. Pass context from entry points.
2. **Wrap errors with `%w`** - Add context when errors propagate across boundaries.
3. **Check errors immediately** - Handle or propagate every error right after the call.
4. **Use zap.Logger** - Inject logger into structs; use structured logging (`zap.Error`, `zap.String`).
5. **Use interfaces for ports** - Repository and external service interfaces live in `internal/feature/<name>/`.
6. **One concern per file** - Keep functions small and focused.

### What to Avoid

1. **Do NOT add HTTP servers** - The API lives in `idx-web/`. The Go codebase is scraper-only.
2. **Do NOT add Chi/gin/etc routers** - No HTTP frameworks in this Go project.
3. **Do NOT use `fmt.Printf`/`log.Printf`** - Use zap structured logging.
4. **Do NOT hardcode credentials** - Use `config/config.yml` for configuration.
5. **Do NOT create goroutines without context cancellation** - Always tie to `ctx.Done()`.
6. **Do NOT add Firebase/MongoDB code here** - These belong in `idx-web/`.

---

## Go Standards (This Codebase)

### Hexagonal DDD Architecture

Business logic organized into **features** under `internal/feature/`:
- `entity.go` - Domain models and repository interfaces
- `service.go` - Business logic implementation

**External adapters** under `internal/infra/`:
- `db/mongo/` - MongoDB repository implementations
- `scraper/` - Web scraping implementations

### Naming Conventions

| Element | Convention | Example |
|---------|------------|---------|
| Variables | camelCase | `srv`, `repo`, `ctx` |
| Functions | PascalCase (exported), camelCase (unexported) | `FindAll`, `parseHTML` |
| Interfaces | `Repository` or `-er` suffix | `Repository`, `Scraper` |
| Receivers | 1-3 letter abbreviation | `(s *Service)`, `(r *Repository)` |
| Packages | lowercase, single word | `announcement`, `mongo` |

### Entity Standards

```go
type Announcement struct {
    CreatedAt time.Time `json:"created_at" bson:"created_at"`
    ID        string    `json:"id" bson:"_id,omitempty"`
    Title     *string   `json:"title" bson:"title"`  // pointer for optional
}
```

- Use pointers (`*string`) for optional fields
- Include `json` and `bson` tags (snake_case)
- Use `omitempty` for `_id` field

### Error Handling

```go
// Good: wrap with context
if err != nil {
    return fmt.Errorf("finding announcement: %w", err)
}

// Good: nil check
user, err := s.repo.FindByUID(ctx, uid)
if err != nil {
    return err
}
```

---

## Feature Implementation

### Adding a New Feature

1. Create `internal/feature/<name>/entity.go` - Define entity and repository interface
2. Create `internal/feature/<name>/service.go` - Implement business logic
3. Create `internal/infra/db/mongo/<name>_repo.go` - Implement repository adapter
4. Call from CLI entry point in `cmd/<command>/main.go`

### Data Persistence (MongoDB)

- Use `go.mongodb.org/mongo-driver/v2`
- Repositories in `internal/infra/db/mongo/`
- Generate boilerplate: `go generate ./internal/feature/...`

### Scraping Logic

- Scrapers in `internal/infra/scraper/`
- Use `selenium` for JavaScript-heavy sites
- Use `goquery` for static HTML
- `NewScraper` constructor with `*zap.Logger` and `Browser` interface

---

## Presentation Layer (CLI Only)

CLI entry points in `cmd/<command>/main.go`:
- `cmd/downloader/` - Downloads financial reports
- `cmd/scraper/` - Scrapes announcements/news
- `cmd/issuer/` - Manages issuer list
- `cmd/announcement/` - Announcement operations

```go
func main() {
    var configPath string
    flag.StringVar(&configPath, "config", "config/config.yml", "Path to config")
    flag.Parse()
    
    // ... setup logger, config, services
}
```

---

## Logging & Concurrency

### Logging (zap)

```go
logger.Info("processing started", zap.String("issuer", code))
logger.Error("fetch failed", zap.Error(err), zap.String("url", url))
```

### Context Propagation

```go
func (s *Service) FindAll(ctx context.Context, filter any) ([]*Announcement, error) {
    return s.repo.FindAll(ctx, filter)
}
```

### Goroutines

```go
select {
case <-ctx.Done():
    return ctx.Err()
case result := <-ch:
    // process
}
```

---

## Testing

- `_test.go` files in same package as code
- Table-driven tests for complex logic
- `zap.NewNop()` to silence logs in tests
- Use local HTML samples for scraper tests

---

## Development Environment

| Requirement | Details |
|-------------|---------|
| Go | 1.24.0+ |
| MongoDB | Local or Docker on `localhost:27017` |
| ChromeDriver | Must match Chrome version |
| Config | `config/config.yml` (copy from `example_config.yml`) |

---

## Automation Commands

```bash
# Run all tests
go test ./...

# Build CLI tools
go build -o bin/ ./cmd/...

# Lint
go vet ./...

# Generate repository boilerplate
go generate ./internal/feature/...
```

---

## idx-web (Nuxt API Server)

The API server is a **separate Nuxt 4 project** in `idx-web/`.

### API Endpoints (idx-web)

| Endpoint | Description |
|----------|-------------|
| `GET /api/v1/announcements` | All announcements with `is_watched` |
| `GET /api/v1/announcements/{id}` | Single announcement |
| `GET /api/v1/news` | News with filters |
| `GET /api/v1/news/{id}` | Single news |
| `GET /api/v1/financial-reports` | All financial reports |
| `GET /api/v1/user/watchlist` | User's watchlist (auth required) |
| `PUT /api/v1/user/watchlist` | Update watchlist (auth required) |

### idx-web Commands

```bash
cd idx-web
npm install
npm run dev      # Development
npm run build    # Production build
```

### idx-web Environment Variables

```bash
FIREBASE_CREDENTIALS_PATH=/path/to/firebase-credentials.json
FIREBASE_API_KEY=
FIREBASE_AUTH_DOMAIN=
FIREBASE_PROJECT_ID=
MONGODB_URI=mongodb://localhost:27017
MONGODB_DB_NAME=idx_scraper
```

---

## Glossary

- **Issuer** - Company listed on IDX
- **Announcement** - Official disclosure from issuer/IDX
- **Financial Report** - Quarterly/annual statement from issuer
- **Kontan** - Indonesian financial news source used for scraping

---

## Commit Guidelines

- Use imperative mood: "Add feature" not "Added feature"
- Reference issues/tickets when applicable
- Keep commits focused (one logical change per commit)
- Write clear, concise commit messages
