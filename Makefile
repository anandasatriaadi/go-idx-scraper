.PHONY: all build clean test vet help \
        scraper briefing scrape-news announcement issuer downloader parse-xbrl \
        update-prices seed-ticker reset-db \
        web-dev web-build web-preview web-install web

# Detect default OS and Architecture
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

# Auto-detect config path (fallback to config-mac.yml if present on macOS)
ifeq ($(GOOS),darwin)
	DEFAULT_CONFIG = config/config-mac.yml
else
	DEFAULT_CONFIG = config/config.yml
endif

CONFIG_PATH ?= $(if $(wildcard $(DEFAULT_CONFIG)),$(DEFAULT_CONFIG),config/config.yml)
TICKER ?= BBRI
YEARS ?= 5
PERIODS ?= TW1,TW2,TW3,Audit
RANGE ?= 5d
FLAGS ?=

ifeq ($(IS_LINUX),1)
	GOOS = linux
	GOARCH = amd64
endif

all: build

help:
	@echo "================================================================="
	@echo "             IDX Intelligence Terminal & Scraper                 "
	@echo "================================================================="
	@echo "Single-Ticker & Data Seeding:"
	@echo "  make seed-ticker [TICKER=BBRI] [YEARS=5]  - 5-year seeding (XBRL + Yahoo prices + Graham/Piotroski valuation)"
	@echo "  make reset-db                             - Wipe & re-index MongoDB collections (with prompt or FLAGS=\"-force\")"
	@echo "  make update-prices [TICKER=BBRI]          - Fetch & sync Yahoo Finance daily price candles into MongoDB"
	@echo ""
	@echo "Scraping & Ingestion Pipelines:"
	@echo "  make briefing                             - Run 7 AM GMT+8 multi-channel scraper & daily briefing"
	@echo "  make scrape-news                          - Alias for make briefing"
	@echo "  make parse-xbrl [FLAGS=\"-ticker=BBRI\"]    - Stream parse XBRL filings into MongoDB & compute ratios"
	@echo "  make downloader [FLAGS=\"-ticker=BBRI\"]    - Download Financial reports (XBRL/Excel) via Selenium"
	@echo "  make announcement                         - Scrape official IDX disclosures & corporate announcements"
	@echo "  make issuer                               - Update listed stock issuer registry (stock-list.json)"
	@echo ""
	@echo "Web UI Commands (idx-web Nuxt 4):"
	@echo "  make web                                  - Start Nuxt 4 Web UI in development mode (localhost:3000)"
	@echo "  make web-dev                              - Alias for make web"
	@echo "  make web-build                            - Build Nuxt 4 Web UI for production"
	@echo "  make web-preview                          - Preview built production server"
	@echo "  make web-install                          - Install Web UI npm dependencies"
	@echo ""
	@echo "Build & Testing Commands:"
	@echo "  make build                                - Build all 8 Go binaries into bin/ (cmd/* & tools/*)"
	@echo "  make test                                 - Run all Go test suites (go test ./...)"
	@echo "  make vet                                  - Run Go static analysis (go vet ./...)"
	@echo "  make clean                                - Remove built binaries and web build artifacts"
	@echo ""
	@echo "Options & Parameters:"
	@echo "  TICKER=<ticker>                           - Target ticker symbol (default: BBRI)"
	@echo "  YEARS=<years>                             - Historical years count or range (default: 5)"
	@echo "  RANGE=<range>                             - Yahoo price range: 5d, 1mo, 1y, 5y, max (default: 5d)"
	@echo "  CONFIG_PATH=<path>                        - Specify config file (current: $(CONFIG_PATH))"
	@echo "  FLAGS=\"<flags>\"                           - Pass extra CLI flags (e.g. FLAGS=\"-no-headless -clean-db\")"
	@echo "  IS_LINUX=1                                - Cross-compile for Linux amd64"
	@echo "================================================================="

# -----------------------------------------------------------------------------
# Go Build Targets
# -----------------------------------------------------------------------------
build:
	@echo "Building Go binaries for $(GOOS)/$(GOARCH)..."
	@mkdir -p bin
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="-s -w" -o bin/scraper ./cmd/scraper/main.go
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="-s -w" -o bin/announcement ./cmd/announcement/main.go
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="-s -w" -o bin/issuer ./cmd/issuer/main.go
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="-s -w" -o bin/downloader ./cmd/downloader/main.go
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="-s -w" -o bin/xbrl_parser ./cmd/xbrl_parser/main.go
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="-s -w" -o bin/price_updater ./cmd/price_updater/main.go
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="-s -w" -o bin/seed_ticker ./tools/seed_ticker/main.go
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="-s -w" -o bin/reset_db ./tools/reset_db/main.go
	@echo "Build complete. Binaries saved to bin/"

# -----------------------------------------------------------------------------
# Pipeline Runners & Seeder Tools
# -----------------------------------------------------------------------------
seed-ticker:
	go run ./tools/seed_ticker/main.go -ticker=$(TICKER) -years=$(YEARS) --config $(CONFIG_PATH) $(FLAGS)

reset-db:
	go run ./tools/reset_db/main.go --config $(CONFIG_PATH) $(FLAGS)

update-prices:
	go run ./cmd/price_updater/main.go $(if $(TICKER),-ticker=$(TICKER),) $(if $(RANGE),-range=$(RANGE),) --config $(CONFIG_PATH) $(FLAGS)

briefing:
	go run ./cmd/scraper/main.go --config $(CONFIG_PATH) $(FLAGS)

scrape-news: briefing

parse-xbrl:
	go run ./cmd/xbrl_parser/main.go --config $(CONFIG_PATH) $(FLAGS)

announcement:
	go run ./cmd/announcement/main.go --config $(CONFIG_PATH) $(FLAGS)

issuer:
	go run ./cmd/issuer/main.go --config $(CONFIG_PATH) $(FLAGS)

downloader:
	go run ./cmd/downloader/main.go --config $(CONFIG_PATH) $(FLAGS)

# -----------------------------------------------------------------------------
# Web UI Commands
# -----------------------------------------------------------------------------
web: web-dev

web-dev:
	@echo "Starting Nuxt 4 Web UI on http://localhost:3000..."
	npm --prefix idx-web run dev

web-build:
	@echo "Building Nuxt 4 Web UI for production..."
	npm --prefix idx-web run build

web-preview:
	@echo "Starting production preview..."
	npm --prefix idx-web run preview

web-install:
	@echo "Installing Web UI dependencies..."
	npm --prefix idx-web install

# -----------------------------------------------------------------------------
# Quality & Maintenance
# -----------------------------------------------------------------------------
test:
	go test -v ./...

vet:
	go vet ./...

clean:
	rm -rf bin/
	rm -rf idx-web/.nuxt idx-web/.output
