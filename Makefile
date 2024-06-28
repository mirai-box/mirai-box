PROJECT_NAME := miraibox

BIN_DIR := bin
BINARY := app

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
MAIN_GO := cmd/http/main.go
MAIN_CGI_GO := cmd/cgi/main.go

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

build/cgi:
	@echo "  >  Building binary for CGI ..."
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=linux $(GOBUILD) -o $(GOBIN)/$(BINARY) $(MAIN_CGI_GO)

# Run the application
run/local:
	@echo "  >  Running application locally..."
	$(GORUN) $(MAIN_GO)

# Update mocks for tests
test/update-mocks:
	$(MOCKERY) --all 

# Run tests across the project
test:
	@echo "  >  Running tests..."
	$(GOTEST) -v ./...

# Clean up binaries and cached data
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