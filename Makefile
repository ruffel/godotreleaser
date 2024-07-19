#===============================================================================
##@ CORE COMMANDS
#===============================================================================

.PHONY: all
all: ## Default 'make' target. Lints and builds the binaries.
	@$(MAKE) build
	@$(MAKE) lint
	@$(MAKE) test

#===============================================================================
##@ ADDITIONAL COMMANDS
#===============================================================================

.PHONY: build
build: ## Compile go packages, dependencies and binaries.
	@CGO_ENABLED=0 go build -ldflags $(LDFLAGS) -o ./bin/godotreleaser ./cmd/godotreleaser


.PHONY: lint
lint: ## Run Golang linter to flag programming errors.
	@golangci-lint run --timeout 10m0s ./... --max-same-issues 0

.PHONY: test
test: ## Run unit tests found in this repository and report results.
	@$(MAKE) test-packages
	@$(MAKE) test-coverage

.PHONY: test-packages
test-packages: ## Run unit tests and display results grouped by package.
	@gotestsum --format pkgname-and-test-fails --jsonfile ./bin/test.log -- -race -cover -count=1 -coverprofile=./bin/coverage.out ./...

.PHONY: test-coverage
test-coverage: ## Display code coverage statistics of unit tests.
	@go tool cover -func=./bin/coverage.out | tail -n 1 | awk '{$$1=$$1;print}'
	@go tool cover -html=./bin/coverage.out -o ./bin/coverage.html

.PHONY: help
help:  ## Show help for command.
	@awk 'BEGIN {FS = ":.*##"; printf "\033[1m\033[37mUSAGE\033[0m\n  make <command>\033[36m\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

%::
	@printf "make: *** No rule to make target '$@'.  Stop.\n\n" >&2
	@$(MAKE) help >&2

#-------------------------------------------------------------------------------
# Environment flags
#-------------------------------------------------------------------------------
.ONESHELL:
SHELL = /bin/bash
.SHELLFLAGS = -cEeuo pipefail
MAKEFLAGS += --no-print-directory

#-------------------------------------------------------------------------------
# Makefile variables.
#-------------------------------------------------------------------------------
VERSION=$(shell git describe --tags --always --dirty)

LDFLAGS = $(eval LDFLAGS := "\
-X 'github.com/ruffel/godotreleaser/internal/cmd/version.version=${VERSION}'\
")$(LDFLAGS)
