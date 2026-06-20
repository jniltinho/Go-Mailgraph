## Variables for UPX
UPX_VERSION := 5.1.1
UPX_ARCHIVE  := upx-$(UPX_VERSION)-amd64_linux.tar.xz
UPX_DIR      := upx-$(UPX_VERSION)-amd64_linux
UPX_BIN      := /usr/local/bin/upx
UPX_URL      := https://github.com/upx/upx/releases/download/v$(UPX_VERSION)/$(UPX_ARCHIVE)

## Variables for Go application
APP        := mailgraph
BIN        := bin/$(APP)
PKG        := github.com/davidullrich/mailgraph/internal/buildinfo
VERSION    := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
IMAGE      ?= davidullrich/mailgraph:latest

LDFLAGS := -trimpath -ldflags "-s -w \
	-X $(PKG).Version=$(VERSION) \
	-X $(PKG).BuildDate=$(BUILD_TIME) \
	-X $(PKG).GitCommit=$(GIT_COMMIT)"

.PHONY: all build build-prod run clean deps tidy test install-upx build-docker certs help

all: clean build

build: clean
	@echo "Building $(APP)..."
	CGO_ENABLED=0 go build -o $(BIN) $(LDFLAGS) .

build-prod: clean
	@echo "Building $(APP) (UPX compressed)..."
	CGO_ENABLED=0 go build -o $(BIN) $(LDFLAGS) .
	upx --best --lzma $(BIN)

run: build
	@echo "Starting $(APP)..."
	./$(BIN) server \
		--logfile=/var/log/mail/mail.log \
		--daemon-rrd=./rrd \
		--listen=:8080

test:
	@echo "Running tests..."
	go test ./...

clean:
	@echo "Cleaning up..."
	rm -f $(BIN)
	rm -rf $(UPX_DIR) $(UPX_ARCHIVE)

tidy:
	@echo "Tidying go modules..."
	go mod tidy

deps:
	@echo "Installing dependencies..."
	go mod download

install-upx:
	@echo "Installing UPX binary..."
	curl -ksSL "$(UPX_URL)" -o "$(UPX_ARCHIVE)"
	tar -xf "$(UPX_ARCHIVE)"
	chmod +x "$(UPX_DIR)/upx"
	sudo mv "$(UPX_DIR)/upx" "$(UPX_BIN)"
	rm -rf "$(UPX_DIR)" "$(UPX_ARCHIVE)"

build-docker:
	@echo "Building Docker image $(IMAGE)..."
	docker build --no-cache --progress=plain \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		-t $(IMAGE) .

certs:
	@echo "Generating SSL certificates..."
	mkdir -p ssl
	openssl req -x509 -nodes -days 3650 -newkey rsa:2048 \
		-keyout ssl/server.key -out ssl/server.crt \
		-subj "/C=BR/ST=SP/L=Sao Paulo/O=Development/CN=localhost"

help:
	@echo "Makefile commands:"
	@echo "  build         - Build Go binary (trimpath + ldflags)"
	@echo "  build-prod    - Build + UPX compression (for releases)"
	@echo "  build-docker  - Multi-stage Docker image ($(IMAGE))"
	@echo "  run           - Build and start mailgraph locally"
	@echo "  test          - Run go test ./..."
	@echo "  clean         - Remove bin/$(APP)"
	@echo "  deps / tidy   - Go module download / cleanup"
	@echo "  install-upx   - Install UPX compressor"
	@echo "  certs         - Generate self-signed TLS certs in ssl/"
	@echo ""
	@echo "Examples:"
	@echo "  make build-prod"
	@echo "  make build-docker IMAGE=jniltinho/mailgraph:latest"
	@echo "  make certs"