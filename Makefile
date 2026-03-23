BINARY_NAME=kvgo-server
BUILD_DIR=bin
MAIN_PATH=./cmd/kvgo-server/
GO_FILES=$(shell find . -name '*.go')
COVERAGE_FILE=coverage.out

DOCKER_IMAGE=kvgo
DOCKER_TAG=latest

BLUE=\033[0;34m
NC=\033[0m

.PHONY: all build run test clean help docker-build docker-up docker-down docker-status docker-logs

all: build

build:
	@echo -e "$(BLUE)Building $(BINARY_NAME)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

run: build
	@echo -e "$(BLUE)Running $(BINARY_NAME)...$(NC)"
	@./$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

test:
	@echo -e "$(BLUE)Running tests with -race...$(NC)"
	@go test -v -race ./...

coverage:
	@echo -e "$(BLUE)Running tests...$(NC)"
	@go test -race -coverprofile=$(COVERAGE_FILE) ./...
	@echo -e "$(BLUE)Generating HTML report...$(NC)"
	@go tool cover -html=$(COVERAGE_FILE) -o coverage.html
	@echo -e "$(BLUE)Global Statistics:$(NC)"
	@go tool cover -func=$(COVERAGE_FILE)

clean:
	@echo -e "$(BLUE)Cleaning...$(NC)"
	@rm -rf $(BUILD_DIR)
	@rm -f $(COVERAGE_FILE) coverage.html

docker-build:
	@echo -e "$(BLUE)Building Docker image $(DOCKER_IMAGE):$(DOCKER_TAG)...$(NC)"
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-up: docker-build
	@echo -e "$(BLUE)Starting services with docker-compose...$(NC)"
	@docker-compose up -d

docker-down:
	@echo -e "$(BLUE)Stopping services...$(NC)"
	@docker-compose down

docker-status:
	@echo -e "$(BLUE)Checking services status...$(NC)"
	@docker-compose ps

docker-logs:
	@echo -e "$(BLUE)Showing logs for all services...$(NC)"
	@docker-compose logs -f