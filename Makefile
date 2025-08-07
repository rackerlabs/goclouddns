# default target is build
.DEFAULT_GOAL := help

.PHONY: help
help: ## Displays this help message
	@echo "$$(grep -hE '^\S+:.*##' $(MAKEFILE_LIST) | sed -e 's/:.*##\s*/|/' -e 's/^\(.\+\):\(.*\)/\\x1b[36m\1\\x1b[m:\2/' | column -c2 -t -s'|' | sort)"

.PHONY: lint
lint: ## Run golangci-lint
	golangci-lint run

.PHONY: fmt
fmt: ## Ensure consistent code style
	@go mod tidy
	@go fmt ./...
	@golangci-lint run --fix

.PHONY: run
run: ## Run the test binary locally
	@go run cmd/clouddns/main.go

.PHONY: build
build: ## Build the binary locally
	@go build -o bin/clouddns ./cmd/clouddns

.PHONY: release-test
release-test: ## Test goreleaser configuration without publishing
	goreleaser release --snapshot --clean

.PHONY: test
test: ## Run tests
	go test -race ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
