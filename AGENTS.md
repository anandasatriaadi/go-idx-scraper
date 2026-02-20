# Agent Guide: go-idx-scraper

High-level project context, development standards, and actionable guidance for AI agents working on this repository.

## Project Context
`go-idx-scraper` is a Go-based toolset for fetching and processing financial data from the Indonesia Stock Exchange (IDX). It automates the collection of stock issuer lists, financial reports, and news for analysis.

## Architectural Pattern: Hexagonal DDD
The project follows a Hexagonal (Ports and Adapters) architecture with Domain-Driven Design (DDD) principles. 

**Note on Nomenclature:** Business logic and entities are organized into "features" (formerly `domain`). All new business domains must be placed under `internal/feature/`.

### Layered Structure
- **Core (Feature):** `internal/feature/` - Contains the business logic, entities, and repository interfaces (Ports). This layer is independent of any external libraries or databases.
- **Infrastructure:** `internal/infra/` - External implementations (Adapters) for DB, scrapers, and external APIs. This layer interacts with the outside world.
- **Presentation:** `internal/presentation/` - Entry points for the application, such as HTTP handlers or CLI commands.
- **Helper:** `internal/helper/` - Pure utility functions without dependencies on core business logic.

## Idiomatic Go Standards
Adhere to standard Go practices as defined in "Effective Go" and the Go Wiki's "Code Review Comments".

### Naming Conventions
Follow standard library patterns for naming:
- **Variables:** Use `camelCase`. Keep names short but descriptive (e.g., `srv` for service, `repo` for repository, `ctx` for context).
- **Functions:** Use `PascalCase` for exported and `camelCase` for unexported. Names should be descriptive verbs (e.g., `FetchNews`, `parseHTML`).
- **Interfaces:** Usually end in `-er` if they define a single method (e.g., `Scraper`). For repository ports, naming it `Repository` within the feature package is standard (e.g., `announcement.Repository`).
- **Receiver Names:** Use 1-3 letter abbreviations of the struct type (e.g., `func (s *Service) ...` or `func (h *AnnouncementHandler) ...`). Never use `this` or `self`.
- **Package Names:** Short, concise, all lowercase, and single words. Avoid underscores or mixed caps.

### Error Handling
- **Propagate Errors:** Wrap errors only when adding significant context. Use `%w` with `fmt.Errorf`.
- **Nil Checks:** Always check for errors immediately after a function call.
- **Custom Errors:** Define specific error types or variables for sentinel errors in the feature package if they need to be handled specifically by callers.

## Feature Implementation Details (`internal/feature/`)
Each feature folder (e.g., `internal/feature/announcement`) should typically contain:
- `entity.go`: The core business models (Structs).
- `service.go`: Business logic implementations.
- `repository.go` (or interface in `entity.go`): Interface definitions for data persistence.

### Entity Standards
- Use pointers for optional fields to distinguish between zero-values and missing data in MongoDB.
- Include `json` and `bson` tags (snake_case) for all fields.
- Example:
  ```go
  type Announcement struct {
      CreatedAt         time.Time           `json:"created_at" bson:"created_at"`
      UpdatedAt         time.Time           `json:"updated_at" bson:"updated_at"`
      ID                string              `json:"id" bson:"_id,omitempty"`
      Title             *string             `json:"title" bson:"title"`
  }
  ```

## Infrastructure & Adapters (`internal/infra/`)
Adapters implement the interfaces defined in the feature layer.

### Data Persistence (MongoDB)
- Use `go.mongodb.org/mongo-driver/v2`.
- Repositories should be located in `internal/infra/db/mongo/`.
- **Repository Generation:** Use the custom tool for boilerplate: `go generate ./internal/feature/...`.

### Scraping Logic
- Scrapers should be modular and located in `internal/infra/scraper/`.
- Use `selenium` for JavaScript-heavy sites.
- Use `goquery` for static HTML parsing.
- Always include a `NewScraper` constructor that accepts a `*zap.Logger` and a `Browser` interface.

## Presentation Layer (`internal/presentation/`)
### HTTP (Chi)
- Handlers should be structs that take services as dependencies.
- Group routes using `chi.Router`.
- Keep handlers thin; delegate all logic to services.
- Handlers should return standard JSON responses and handle errors with appropriate HTTP status codes.

### CLI
- Entry points are located in `cmd/<command>/main.go`.
- Use standard `flag` package for configuration like `configPath`.
- Each major action has its own dedicated main file.

## Logging & Concurrency
### Logging (Zap)
- Inject `*zap.Logger` into structs.
- Use structured logging: `logger.Error("failed to fetch", zap.Error(err), zap.String("url", url))`.
- Avoid using `fmt.Printf` or `log.Printf` for application-level logging.

### Concurrency & Context
- Always propagate `context.Context` through all layers.
- Use context for timeouts and cancellation, especially in network and DB calls.
- Avoid using `context.Background()` deep in the call stack; pass it from the entry point (main/handler).
- When using Goroutines, ensure they are properly managed and can be cancelled via context.

## Testing Guidance
### Unit Testing
- Place `_test.go` files in the same package as the code being tested.
- Use table-driven tests for complex logic to cover multiple edge cases.
- Use `zap.NewNop()` to silence logs during tests unless you are specifically testing log output.

### Mocking & Integration
- For scrapers, use local HTML files or strings to test parsers without making network requests.
- Ensure tests are independent and clean up after themselves (e.g., closing DB connections).
- Use interfaces to allow for easy mocking of repositories in service tests.

## Development Environment Tips
- **Runtime:** Go 1.24.0+
- **Config:** Use `config/config.yml`. Copy from `example_config.yml` and adjust for local environment.
- **Tooling:** Ensure `chromedriver` is installed for `selenium` tasks.
- **Database:** A local MongoDB instance or a Docker container is recommended for development.

## Common Agent Workflows
1. **Adding a New Feature:**
   - Create `internal/feature/<name>/`.
   - Define entity and repository interface in `entity.go`.
   - Implement business logic in `service.go`.
   - Implement repository in `internal/infra/db/mongo/`.
   - Register the feature in `cmd/` or `internal/presentation/http/`.
2. **Updating Scrapers:**
   - Modify the implementation in `internal/infra/scraper/`.
   - Update tests in the same directory with new sample HTML if the site structure changed.
   - Verify the scraper by running the corresponding command in `cmd/`.
3. **Configuration Changes:**
   - Update `internal/config/config.go` struct.
   - Update `config/example_config.yml` to reflect changes for other developers.

## Automation Commands
- **Generate Mocks/Repos:** `go generate ./...`
- **Run All Tests:** `go test ./...`
- **Build CLI:** `go build -o bin/ ./cmd/...`
- **Run Server:** `go run cmd/server/main.go`
- **Check Linting:** `go vet ./...` (or use a configured linter like golangci-lint)

## Project Specific Details
- **Emailing:** Uses `gomail.v2` and templates in `internal/helper/email.go`.
- **Excel:** Uses `xuri/excelize/v2` for processing and generating financial report workbooks.
- **OpenRouter:** Integrated for AI-assisted processing of scraped content, such as summarization or sentiment analysis.
- **Browser Automation:** `selenium` is used for scraping dynamic content. Ensure the `chromedriver_path` is correctly configured in `config.yml`.

## Repository Management
- **Git:** Follow standard Git workflows. Create feature branches and submit PRs.
- **Commits:** Write clear, concise commit messages following the project's history.
- **Documentation:** Keep this `AGENTS.md` and the root `README.md` updated as the project evolves.

## Glossary
- **Issuer:** A company that listed its stocks on the IDX.
- **Announcement:** Official disclosure from an issuer or IDX.
- **Financial Report:** Periodic financial statement (Quarterly/Annual) from an issuer.
- **Kontan:** A popular financial news source in Indonesia used for scraping.
