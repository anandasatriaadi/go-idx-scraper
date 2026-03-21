# Hexagonal DDD Architecture - go-idx-scraper

This document describes the Hexagonal (Ports & Adapters) architecture with Domain-Driven Design (DDD) principles implemented in the go-idx-scraper project.

## Architecture Overview

The project is organized into three main layers following the Hexagonal Architecture pattern:

```
┌─────────────────────────────────────────────────────────┐
│                   Presentation Layer                    │
│                 (Entry Points & Handlers)                │
│  - cmd/*/main.go (CLI commands)                         │
│  - internal/presentation/http/* (HTTP Handlers)         │
└──────────────────────┬──────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────┐
│                   Feature Layer (Core)                  │
│              (Business Logic & Entities)                │
│  - internal/feature/announcement/                       │
│  - internal/feature/news/                               │
│  - internal/feature/finreport/                          │
│  - internal/feature/system/                             │
│  - internal/feature/stock/                              │
└──────────────────────┬──────────────────────────────────┘
                       │
        ┌──────────────┼──────────────┐
        ▼              ▼              ▼
    Repositories   Services      Ports (Interfaces)
    (Database)    (Business     (External Services)
                   Logic)
        │              │              │
        └──────────────┼──────────────┘
                       │
┌──────────────────────▼──────────────────────────────────┐
│                Infrastructure Layer                     │
│          (Adapters & External Implementations)          │
│  - internal/infra/db/mongo/* (Repository Adapters)    │
│  - internal/infra/idx/* (IDX API Adapter)              │
│  - internal/infra/scraper/kontan/* (Scraper Adapter)  │
│  - internal/infra/logger/* (Logging Adapter)           │
│  - internal/browser/* (Selenium Browser Adapter)       │
└─────────────────────────────────────────────────────────┘
```

## Layer Responsibilities

### 1. Feature Layer (Core - `internal/feature/`)

**Purpose:** Contains pure business logic, entities, and port definitions.

**Key Components:**

Each feature (announcement, news, finreport, system, stock) contains:

- **entity.go**: 
  - Defines domain entities (structs with business fields)
  - Defines Port interfaces (Repository, external services)
  - Ports use `json` and `bson` tags for persistence
  - Optional fields use pointers to distinguish zero-values from missing data

- **service.go**:
  - Implements business logic
  - Depends on Repository ports (injected via constructor)
  - Takes `*zap.Logger` for structured logging
  - Pure business logic independent of frameworks

#### Example Feature Structure: Announcement

```go
// entity.go
type Announcement struct {
    ID                string              `json:"id" bson:"_id,omitempty"`
    JudulPengumuman   *string             `json:"judul_pengumuman" bson:"judul_pengumuman"`
    CreatedAt         time.Time           `json:"created_at" bson:"created_at"`
}

// Port: Output Port (Data Persistence)
type Repository interface {
    Create(ctx context.Context, announcement *Announcement) error
    FindByID(ctx context.Context, id string) (*Announcement, error)
}

// Port: External Service Port
type IDXDataProvider interface {
    Fetch(ctx context.Context, dateFrom, dateTo string) ([]*Announcement, error)
}

// service.go
type Service struct {
    repo   Repository
    logger *zap.Logger
}

func NewService(repo Repository, logger *zap.Logger) *Service {
    return &Service{repo: repo, logger: logger}
}
```

### 2. Infrastructure Layer (`internal/infra/`)

**Purpose:** Implements ports defined in the feature layer.

**Key Components:**

- **db/mongo/**: Repository implementations (Output Port Adapters)
  - `announcement_repo.go`: Implements announcement.Repository
  - `news_repo.go`: Implements news.Repository
  - `financial_report_repo.go`: Implements finreport.Repository
  - `system_repo.go`: Implements system.Repository
  
- **idx/**: IDX API adapter (External Service Adapter)
  - `adapter.go`: Contains IDXProvider implementing announcement.IDXDataProvider
  - Handles HTTP requests to IDX API
  - Parses responses into domain entities

- **scraper/kontan/**: News scraper implementation
  - Implements the news.Scraper port
  - Uses Selenium for dynamic content

- **browser/**: Browser abstraction for web automation
  - Selenium WebDriver setup and utilities

- **logger/**: Logging configuration

#### Example Adapter: IDXProvider

```go
// internal/infra/idx/adapter.go
type IDXProvider struct {
    logger *zap.Logger
    driver selenium.WebDriver
}

func NewIDXProvider(logger *zap.Logger, driver selenium.WebDriver) announcement.IDXDataProvider {
    return &IDXProvider{logger: logger, driver: driver}
}

func (p *IDXProvider) Fetch(ctx context.Context, dateFrom, dateTo string) ([]*announcement.Announcement, error) {
    // Implementation using the port interface
}
```

### 3. Presentation Layer (`internal/presentation/`)

**Purpose:** Entry points for the application (CLI commands, HTTP handlers).

**Key Components:**

- **http/**: HTTP request handlers
  - `announcement_handler.go`: HTTP endpoints for announcements
  - `news_handler.go`: HTTP endpoints for news
  - `financial_report_handler.go`: HTTP endpoints for financial reports

- **cmd/**: CLI entry points
  - `cmd/announcement/main.go`: Fetch announcements from IDX
  - `cmd/scraper/main.go`: Scrape news from Kontan
  - `cmd/downloader/main.go`: Download financial reports
  - `cmd/issuer/main.go`: Fetch issuer list
  - `cmd/server/main.go`: HTTP server

## Dependency Injection Pattern

Dependencies flow inward from infrastructure to core:

```
HTTP Request
    ↓
Presentation Layer (Handler)
    ↓
Feature Layer (Service)
    ↓
Infrastructure Layer (Adapters)
    ↓
External Systems (Database, APIs)
```

### Initialization Example (from cmd/server/main.go)

```go
// 1. Create adapters (Infrastructure)
dbClient := mongo.NewClient(logger)
database := dbClient.Database("db_name")

// 2. Create repository adapters
announcementRepo := mongo.NewAnnouncementRepository(database)

// 3. Create feature services
announcementService := announcement.NewService(announcementRepo, logger)

// 4. Create presentation handlers
announcementHandler := handlers.NewAnnouncementHandler(announcementService)

// 5. Wire up HTTP routes
r.Mount("/announcements", announcementHandler.Routes())
```

## Key Principles Applied

### 1. **Separation of Concerns**
- Feature layer: Pure business logic
- Infrastructure layer: Technical details
- Presentation layer: User interaction

### 2. **Dependency Inversion**
- Features depend on abstractions (Ports/Interfaces)
- Infrastructure implements these abstractions
- No upward dependencies

### 3. **Testability**
- Ports enable mock implementations for testing
- Services receive dependencies via constructor injection
- No global state or singletons

### 4. **Framework Independence**
- Core business logic has zero framework dependencies
- Easy to swap implementations (e.g., MongoDB → PostgreSQL)
- Ports define contracts, not implementations

## Directory Structure

```
go-idx-scraper/
├── cmd/                          # CLI Entry Points
│   ├── announcement/main.go       # Announcement scraper CLI
│   ├── scraper/main.go            # News scraper CLI
│   ├── downloader/main.go         # Report downloader CLI
│   ├── issuer/main.go             # Issuer fetcher CLI
│   └── server/main.go             # HTTP API server
│
├── internal/
│   ├── feature/                   # CORE LAYER - Business Logic
│   │   ├── announcement/          # Announcement feature
│   │   │   ├── entity.go          # Entities + Ports
│   │   │   └── service.go         # Business logic
│   │   ├── news/                  # News feature
│   │   ├── finreport/             # Financial report feature
│   │   ├── system/                # System metadata feature
│   │   ├── stock/                 # Stock data feature
│   │   └── common/                # Shared types
│   │
│   ├── infra/                     # INFRASTRUCTURE LAYER
│   │   ├── db/mongo/              # Database adapters
│   │   │   ├── announcement_repo.go
│   │   │   ├── news_repo.go
│   │   │   ├── financial_report_repo.go
│   │   │   └── system_repo.go
│   │   ├── idx/                   # IDX API adapter
│   │   │   └── adapter.go         # IDXProvider implementation
│   │   ├── scraper/               # Web scraper adapters
│   │   │   └── kontan/            # Kontan news scraper
│   │   ├── browser/               # Browser automation
│   │   ├── logger/                # Logging setup
│   │   └── external/              # Other external services
│   │
│   ├── presentation/              # PRESENTATION LAYER
│   │   └── http/                  # HTTP handlers
│   │       ├── announcement_handler.go
│   │       ├── news_handler.go
│   │       └── financial_report_handler.go
│   │
│   ├── config/                    # Configuration management
│   ├── browser/                   # Browser abstraction
│   └── helper/                    # Utility functions
│
├── config/                        # Configuration files
├── AGENTS.md                      # Agent guidelines
├── ARCHITECTURE.md                # This file
└── README.md                      # Project overview
```

## Feature Examples

### Announcement Feature

**Ports:**
- `Repository`: Persist/retrieve announcements (Output)
- `IDXDataProvider`: Fetch announcement data from IDX API (External Service)

**Service:** Business logic for managing announcements

**Adapters:**
- `mongo.AnnouncementRepository`: MongoDB implementation
- `idx.IDXProvider`: IDX API implementation

### News Feature

**Ports:**
- `Repository`: Persist/retrieve news (Output)
- `Scraper`: Fetch news from sources (External Service)

**Service:** Business logic including AI-powered summarization

**Adapters:**
- `mongo.NewsRepository`: MongoDB implementation
- `kontan.Scraper`: Kontan website scraper

## Design Patterns Used

### 1. **Ports & Adapters (Hexagonal Architecture)**
- Define contracts as interfaces in feature layer
- Implement contracts in infrastructure layer
- Enables easy testing and swapping implementations

### 2. **Dependency Injection**
- Constructor injection of repositories and services
- Loose coupling between layers
- Explicit dependencies

### 3. **Repository Pattern**
- Data access abstraction
- Consistent interface across all entities
- Easy to test with mock repositories

### 4. **Service Layer Pattern**
- Business logic encapsulation
- Orchestrates between repositories and ports
- Unit testable without database

### 5. **CLI & HTTP Handlers**
- Thin presentation layer
- Delegates all logic to services
- Handles request/response serialization

## Adding New Features

To add a new feature (e.g., `internal/feature/myfeature/`):

1. **Define Entity & Ports** (`entity.go`):
```go
package myfeature

type MyEntity struct {
    ID    string    `json:"id" bson:"_id,omitempty"`
    Name  string    `json:"name" bson:"name"`
}

type Repository interface {
    Create(ctx context.Context, entity *MyEntity) error
    FindByID(ctx context.Context, id string) (*MyEntity, error)
}

type ExternalService interface {
    FetchData(ctx context.Context, id string) (string, error)
}
```

2. **Implement Service** (`service.go`):
```go
type Service struct {
    repo   Repository
    logger *zap.Logger
}

func NewService(repo Repository, logger *zap.Logger) *Service {
    return &Service{repo: repo, logger: logger}
}
```

3. **Create Repository Adapter** (`internal/infra/db/mongo/myfeature_repo.go`):
```go
type MyFeatureRepository struct {
    collection *mongo.Collection
}

func NewMyFeatureRepository(db *mongo.Database) myfeature.Repository {
    return &MyFeatureRepository{
        collection: db.Collection("myfeature"),
    }
}

func (r *MyFeatureRepository) Create(ctx context.Context, entity *myfeature.MyEntity) error {
    // Implementation
}
```

4. **Create HTTP Handler** (if needed):
```go
type MyFeatureHandler struct {
    service *myfeature.Service
}

func (h *MyFeatureHandler) Routes() chi.Router {
    r := chi.NewRouter()
    r.Post("/", h.Create)
    return r
}
```

5. **Wire in cmd or server**:
```go
repo := mongo.NewMyFeatureRepository(db)
service := myfeature.NewService(repo, logger)
handler := handlers.NewMyFeatureHandler(service)
r.Mount("/myfeature", handler.Routes())
```

## Testing Strategy

### Unit Tests
- Test services with mock repositories
- Test handlers with mock services
- Located in `*_test.go` files alongside code

### Integration Tests
- Use local MongoDB or Docker
- Test repository implementations
- Test adapter implementations

### Example Mock Repository
```go
type MockRepository struct {
    MockCreate func(ctx context.Context, a *Announcement) error
}

func (m *MockRepository) Create(ctx context.Context, a *Announcement) error {
    return m.MockCreate(ctx, a)
}
```

## Configuration

- Configuration values: `internal/config/config.go`
- Configuration files: `config/*.yml`
- Example config: `config/example_config.yml`

## Logging

- Structured logging using `uber/zap`
- Logger created in `internal/helper/logger.go`
- Injected into services and adapters
- Avoid `fmt.Printf` or `log.Printf` for application logs

## Summary

This Hexagonal DDD architecture provides:

✓ **Clear separation of concerns** - Each layer has single responsibility
✓ **High testability** - Mock ports, test in isolation
✓ **Framework independence** - Core logic is framework-agnostic
✓ **Easy maintenance** - Change implementations without affecting business logic
✓ **Scalability** - Easy to add new features following the pattern
✓ **Professional structure** - Industry-standard architecture pattern
