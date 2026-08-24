.PHONY: all build clean test vet help \
        scraper briefing scrape-news announcement issuer downloader parse-xbrl \
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
	@echo "Scraping & Intelligence Commands:"
	@echo "  make briefing       - Run 7 AM GMT+8 multi-channel scraper & daily briefing"
	@echo "  make scrape-news    - Alias for make briefing"
	@echo "  make parse-xbrl     - Parse XBRL statements & compute forensic ratios"
	@echo "  make downloader     - Run Financial report (XBRL/Excel) downloader"
	@echo "  make announcement   - Run official IDX disclosures & announcements checker"
	@echo "  make issuer         - Run IDX issuer list updater"
	@echo ""
	@echo "Web UI Commands:"
	@echo "  make web            - Start Nuxt 4 Web UI in development mode (localhost:3000)"
	@echo "  make web-dev        - Alias for make web"
	@echo "  make web-build      - Build Nuxt 4 Web UI for production"
	@echo "  make web-preview    - Preview built production server"
	@echo "  make web-install    - Install Web UI npm dependencies"
	@echo ""
	@echo "Build & Testing Commands:"
	@echo "  make build          - Build all 5 Go binaries into bin/"
	@echo "  make test           - Run all Go test suites"
	@echo "  make vet            - Run Go static analysis (go vet)"
	@echo "  make clean          - Remove built binaries and temporary artifacts"
	@echo ""
	@echo "Options:"
	@echo "  CONFIG_PATH=<path>  - Specify config file (current: $(CONFIG_PATH))"
	@echo "  FLAGS=\"<flags>\"     - Pass extra CLI flags (e.g. FLAGS=\"--no-headless\")"
	@echo "  IS_LINUX=1          - Cross-compile for Linux amd64"
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
	@echo "Build complete. Binaries saved to bin/"

# -----------------------------------------------------------------------------
# Scraping & Ingestion Runners
# -----------------------------------------------------------------------------
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
