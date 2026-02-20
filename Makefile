.PHONY: build clean test help scraper announcement issuer downloader server

# Default config path
CONFIG_PATH ?= config/config.yml
FLAGS ?= 

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

build:
	@echo "Building binaries..."
	@mkdir -p bin
	go build -o bin/scraper ./cmd/scraper/main.go
	go build -o bin/announcement ./cmd/announcement/main.go
	go build -o bin/issuer ./cmd/issuer/main.go
	go build -o bin/downloader ./cmd/downloader/main.go
	go build -o bin/server ./cmd/server/main.go

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
