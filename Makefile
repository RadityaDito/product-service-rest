# Makefile for Golang PostgreSQL CRUD Project

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

# Binary name
BINARY_NAME=product-service
BINARY_UNIX=$(BINARY_NAME)_unix

# Docker parameters
DOCKER_COMPOSE=docker-compose
DOCKER_BUILD=docker build
DOCKER_RUN=docker run

# Local run
.PHONY: run
run:
	$(GOCMD) run cmd/main.go

# Build
.PHONY: build
build:
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/main.go

# Clean
.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)

# Test
.PHONY: test
test:
	$(GOTEST) -v ./...

# Docker build
.PHONY: docker-build
docker-build:
	$(DOCKER_BUILD) -t product-service .

# Docker compose up for staging
.PHONY: docker-up-stg
docker-up-stg:
	$(DOCKER_COMPOSE) -f docker-compose.stg.yml up --build

# Docker compose down for staging
.PHONY: docker-down-stg
docker-down-stg:
	$(DOCKER_COMPOSE) -f docker-compose.stg.yml down

# Docker compose up for production
.PHONY: docker-up-prod
docker-up-prod:
	$(DOCKER_COMPOSE) -f docker-compose.prod.yml up --build

# Docker compose down for production with volumes and network cleanup
.PHONY: docker-down-prod
docker-down-prod:
	$(DOCKER_COMPOSE) -f docker-compose.prod.yml down --volumes --remove-orphans

# Install dependencies
.PHONY: deps
deps:
	$(GOGET) -v ./...

# Cross compilation
.PHONY: build-linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v ./cmd/main.go