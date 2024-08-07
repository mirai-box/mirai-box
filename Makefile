PROJECT_NAME := miraibox

BIN_DIR := bin
BINARY := $(PROJECT_NAME)

# Go related variables.
GOBASE := $(shell pwd)
GOBIN := $(GOBASE)/$(BIN_DIR)

# Go commands
GOCMD      := go
GOBUILD    := $(GOCMD) build
GOCLEAN    := $(GOCMD) clean
GOTEST     := $(GOCMD) test
GOGET      := $(GOCMD) get
GORUN      := $(GOCMD) run

# Other tools
DOCKER := docker
MOCKERY := mockery

# Defining the path to the main Go file
MAIN_GO := cmd/service/main.go

.PHONY: all init test clean build/local run/local build/cgi run/docker deps test/update-mocks build/docker

all: build/local

init:
	$(GOCMD) mod download
	$(GOCMD) mod tidy

# Build the application
build/local:
	@echo "  >  Building binary for local environment..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(GOBIN)/$(BINARY) $(MAIN_GO)

# Run the application
run/bin:
	@echo "  >  Running application ..."
	$(GOBIN)/$(BINARY)

# Run the application
run/local:
	@echo "  >  Running application locally..."
	$(GORUN) $(MAIN_GO)

# Update mocks for tests
test/update-mocks:
	$(MOCKERY) --all 

test/unit:
	@echo "  >  Running unit tests..."
	$(GOTEST) -v  -coverprofile=coverage.out ./...

test/integration:
	@echo "  >  Running integration tests..."
	$(GOTEST) -timeout 60s -tags=integration ./...

clean:
	@echo "  >  Cleaning build cache"
	$(GOCLEAN)
	rm -rf $(GOBIN)/$(BINARY)

# Build Docker image
build/docker:
	@echo "  >  Building Docker image..."
	$(DOCKER) build -t $(PROJECT_NAME) .

# Run Docker container
run/docker: build/docker
	@echo "  >  Running Docker container..."
	$(DOCKER) run -p 8080:8080 $(PROJECT_NAME)

# Check and download dependencies
deps:
	@echo "  >  Checking for missing dependencies..."
	$(GOGET) -d ./...