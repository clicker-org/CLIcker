BINARY_NAME := clicker
MODULE := github.com/clicker-org/clicker
CMD_PATH := ./cmd/clicker
BIN_DIR := bin
DIST_DIR := dist

.PHONY: build run test lint clean build-all release

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY_NAME) $(CMD_PATH)

run: build
	./$(BIN_DIR)/$(BINARY_NAME)

test:
	go test ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf $(BIN_DIR) $(DIST_DIR)

purge: clean
	rm -f "$${XDG_CONFIG_HOME:-$$HOME/.config}/clicker/save.json"

build-all:
	@mkdir -p $(DIST_DIR)
	GOOS=linux  GOARCH=amd64  go build -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64   $(CMD_PATH)
	GOOS=linux  GOARCH=arm64  go build -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64   $(CMD_PATH)
	GOOS=darwin GOARCH=amd64  go build -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64  $(CMD_PATH)
	GOOS=darwin GOARCH=arm64  go build -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64  $(CMD_PATH)

release: build-all
	@echo "Release binaries written to $(DIST_DIR)/"
