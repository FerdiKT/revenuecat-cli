BINARY_NAME ?= revenuecat
CMD_PATH ?= ./cmd/revenuecat
VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
DIST_DIR ?= dist
LDFLAGS = -X github.com/FerdiKT/revenuecat-cli/internal/buildinfo.Version=$(VERSION) -X github.com/FerdiKT/revenuecat-cli/internal/buildinfo.Commit=$(COMMIT) -X github.com/FerdiKT/revenuecat-cli/internal/buildinfo.Date=$(DATE)

.PHONY: build run tidy test dist clean

build:
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME) $(CMD_PATH)

run:
	go run $(CMD_PATH)

tidy:
	go mod tidy

test:
	go test ./...

dist: clean
	mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) $(CMD_PATH)
	tar -czf $(DIST_DIR)/$(BINARY_NAME)_$(VERSION)_darwin_amd64.tar.gz $(BINARY_NAME)
	rm $(BINARY_NAME)
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) $(CMD_PATH)
	tar -czf $(DIST_DIR)/$(BINARY_NAME)_$(VERSION)_darwin_arm64.tar.gz $(BINARY_NAME)
	rm $(BINARY_NAME)
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) $(CMD_PATH)
	tar -czf $(DIST_DIR)/$(BINARY_NAME)_$(VERSION)_linux_amd64.tar.gz $(BINARY_NAME)
	rm $(BINARY_NAME)
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) $(CMD_PATH)
	tar -czf $(DIST_DIR)/$(BINARY_NAME)_$(VERSION)_linux_arm64.tar.gz $(BINARY_NAME)
	rm $(BINARY_NAME)
	cd $(DIST_DIR) && shasum -a 256 *.tar.gz > checksums.txt

clean:
	rm -rf $(DIST_DIR) bin
