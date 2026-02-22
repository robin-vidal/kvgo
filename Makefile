BINARY_NAME=kvgo-server
BUILD_DIR=bin
MAIN_PATH=./cmd/kvgo-server/
GO_FILES=$(shell find . -name '*.go')
COVERAGE_FILE=coverage.out

BLUE=\033[0;34m
NC=\033[0m

.PHONY: all build run test clean help

all: build

build:
	@echo "$(BLUE)Building $(BINARY_NAME)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

run: build
	@echo "$(BLUE)Running $(BINARY_NAME)...$(NC)"
	@./$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

test:
	@echo "$(BLUE)Running tests with -race...$(NC)"
	@go test -v -race ./...

coverage:
	@echo "$(BLUE)Running tests...$(NC)"
	@go test -race -coverprofile=$(COVERAGE_FILE) ./...
	@echo "$(BLUE)Generating HTML report...$(NC)"
	@go tool cover -html=$(COVERAGE_FILE) -o coverage.html
	@echo "$(BLUE)Global Statistics:$(NC)"
	@go tool cover -func=$(COVERAGE_FILE)

clean:
	@echo "$(BLUE)Cleaning...$(NC)"
	@rm -rf $(BUILD_DIR)
	@rm -f $(COVERAGE_FILE) coverage.html