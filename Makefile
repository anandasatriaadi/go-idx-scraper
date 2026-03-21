.PHONY: build clean test help scraper announcement issuer downloader server

# Default config path
CONFIG_PATH ?= config/config.yml
FLAGS ?= 

# Build target OS/Arch
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

ifeq ($(IS_LINUX),1)
	GOOS = linux
	GOARCH = amd64
endif

all: build

help:
	@echo "Available commands:"
	@echo "  make build         - Build all binaries"
	@echo "  make scraper       - Run Kontan scraper"
	@echo "  make announcement  - Run IDX announcement checker"
	@echo "  make issuer        - Run Issuer list updater"
	@echo "  make downloader    - Run Financial report downloader"
	@echo "  make server        - Run REST API server"
	@echo "  make test          - Run all unit tests"
	@echo "  make clean         - Remove built binaries"
	@echo ""
	@echo "Options:"
	@echo "  CONFIG_PATH=<path>  - Specify config file path (default: config/config.yml)"
	@echo "  FLAGS=\"<flags>\"    - Pass additional flags to the command (e.g., FLAGS=\"--no-headless\")"
	@echo "  IS_LINUX=1         - Build for Linux x86_64"

build:
	@echo "Building binaries for $(GOOS)/$(GOARCH)..."
	@mkdir -p bin
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="-s -w" -o bin/scraper ./cmd/scraper/main.go
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="-s -w" -o bin/announcement ./cmd/announcement/main.go
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="-s -w" -o bin/issuer ./cmd/issuer/main.go
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="-s -w" -o bin/downloader ./cmd/downloader/main.go
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="-s -w" -o bin/server ./cmd/server/main.go

scraper:
	go run ./cmd/scraper/main.go --config $(CONFIG_PATH) $(FLAGS)

announcement:
	go run ./cmd/announcement/main.go --config $(CONFIG_PATH) $(FLAGS)

issuer:
	go run ./cmd/issuer/main.go --config $(CONFIG_PATH) $(FLAGS)

downloader:
	go run ./cmd/downloader/main.go --config $(CONFIG_PATH) $(FLAGS)

server:
	go run ./cmd/server/main.go --config $(CONFIG_PATH) $(FLAGS)

test:
	go test -v ./...

clean:
	rm -rf bin/
