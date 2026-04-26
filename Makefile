# Plonk development tasks

.PHONY: help build install test lint test-bats test-coverage clean dev-setup find-dead-code docker-build docker-test docker-test-smoke docker-test-file docker-verify docker-shell docker-clean docker-test-all

# Show available targets
help:
	@echo "Available targets:"
	@echo "  build            - Build the plonk binary with version information"
	@echo "  install          - Install plonk to GOPATH/bin"
	@echo "  test             - Run all unit tests"
	@echo "  lint             - Run golangci-lint"
	@echo "  test-bats        - Run BATS behavioral tests locally"
	@echo "  test-coverage    - Run tests with coverage"
	@echo "  clean            - Clean build artifacts and test cache"
	@echo "  dev-setup        - Setup development environment"
	@echo "  find-dead-code   - Find dead code"
	@echo "  docker-build     - Build the Docker test image"
	@echo "  docker-test      - Run all BATS tests in Docker"
	@echo "  docker-test-smoke - Run smoke tests only in Docker"
	@echo "  docker-test-file - Run one BATS test file in Docker (file=path)"
	@echo "  docker-verify    - Verify package managers in Docker"
	@echo "  docker-shell     - Start interactive shell in Docker"
	@echo "  docker-clean     - Clean Docker images and containers"
	@echo "  docker-test-all  - Build image and run Docker tests"

# Build the plonk binary with version information
build:
	@echo "Building plonk with version information..."
	@mkdir -p bin
	@VERSION=$$(git describe --tags --always --dirty 2>/dev/null || echo "dev"); \
	COMMIT=$$(git rev-parse --short HEAD 2>/dev/null || echo "none"); \
	DATE=$$(date -u +"%Y-%m-%dT%H:%M:%SZ"); \
	if ! go build -ldflags "-X main.version=$$VERSION -X main.commit=$$COMMIT -X main.date=$$DATE" -o bin/plonk ./cmd/plonk; then \
		echo "Build failed"; \
		exit 1; \
	fi; \
	echo "Built versioned plonk binary to bin/ (version: $$VERSION)"

# Install plonk to GOPATH/bin
install:
	@echo "Installing plonk..."
	@go install ./cmd/plonk
	@echo "Installed plonk to $$(go env GOPATH)/bin"

# Run all unit tests
test:
	@echo "Running unit tests..."
	@go test ./...
	@echo "Unit tests passed!"

# Run golangci-lint
lint:
	@echo "Running linter..."
	@go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint run --timeout=10m
	@echo "Lint checks passed!"

# Run BATS behavioral tests locally
# WARNING: Prefer 'make docker-test' - local execution modifies your system!
# These tests install real packages and create real files. Only use this if you
# understand the risks. See tests/bats/README.md for details.
test-bats:
	@echo "WARNING: Running BATS tests LOCALLY - this will modify your system!"
	@echo "   Prefer 'make docker-test' for isolated execution."
	@echo ""
	@if ! command -v bats >/dev/null 2>&1; then \
		echo "BATS not found. Install with: brew install bats-core"; \
		exit 1; \
	fi
	@cd tests/bats && PLONK_TEST_CLEANUP_PACKAGES=1 PLONK_TEST_CLEANUP_DOTFILES=1 bats behavioral/
	@echo "BATS tests completed!"

# Run tests with coverage
test-coverage:
	@echo "Running unit tests with normalized coverage..."
	@go clean -cache -testcache
	@COVER_EXCLUDE_REGEX=$${COVER_EXCLUDE_REGEX:-'/(internal/testutil|tools|cmd/.*)$$'}; \
	PKGS=$$(go list ./... | grep -Ev "$$COVER_EXCLUDE_REGEX"); \
	if [ -z "$$PKGS" ]; then \
		echo "No packages selected for coverage after filtering. Check COVER_EXCLUDE_REGEX." >&2; \
		exit 2; \
	fi; \
	COVERPKG=$$(printf '%s\n' $$PKGS | paste -sd, -); \
	go test -coverpkg="$$COVERPKG" -covermode=atomic -coverprofile=coverage.out $$PKGS
	@go tool cover -func=coverage.out > coverage.txt
	@awk 'END{printf "Total coverage (normalized): %s\n", $$3}' coverage.txt
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Unit tests passed! Coverage report: coverage.html (see also coverage.txt)"

# Clean build artifacts and test cache
clean:
	@echo "Cleaning build artifacts and caches..."
	@rm -rf bin dist
	@rm -f coverage.out coverage.html coverage.txt
	@go clean
	@go clean -testcache
	@echo "Build artifacts and test cache cleaned"

# Setup development environment for new contributors
dev-setup:
	@echo "Setting up development environment..."
	@echo "  - Downloading Go dependencies..."
	@go mod download
	@echo "  - Installing pre-commit hooks..."
	@if command -v pre-commit >/dev/null 2>&1; then \
		pre-commit install; \
	else \
		echo "pre-commit not found. Install with: brew install pre-commit"; \
		exit 1; \
	fi
	@echo "Development environment ready!"
	@echo ""
	@echo "Next steps:"
	@echo "  - Run 'make' to see available commands"
	@echo "  - Run 'make build' to build the binary"
	@echo "  - Run 'make test' to run tests"

# Find dead code (unreachable functions)
find-dead-code:
	@echo "Finding dead code..."
	@if ! command -v deadcode >/dev/null 2>&1; then \
		echo "Installing deadcode tool..."; \
		go install golang.org/x/tools/cmd/deadcode@latest; \
	fi
	@deadcode -test ./...

# Docker Commands - Run BATS tests in containerized environment

# Build the Docker test image
docker-build:
	@echo "Building plonk-test Docker image..."
	@docker build -t plonk-test:latest .
	@echo "Docker image built successfully!"

# Run all BATS tests in Docker
docker-test:
	@echo "Running BATS tests in Docker..."
	@docker compose run --rm test
	@echo "Docker tests completed!"

# Run smoke tests only in Docker (fast verification)
docker-test-smoke:
	@echo "Running smoke tests in Docker..."
	@docker compose run --rm smoke
	@echo "Smoke tests completed!"

# Run specific BATS test file in Docker
docker-test-file:
	@if [ -z "$(file)" ]; then \
		echo "Usage: make docker-test-file file=tests/bats/behavioral/02-package-install.bats"; \
		exit 2; \
	fi
	@echo "Running $(file) in Docker..."
	@docker compose run --rm test bats "$(file)"

# Verify all package managers are available in Docker
docker-verify:
	@echo "Verifying package managers in Docker..."
	@docker compose run --rm test verify

# Start interactive shell in Docker container
docker-shell:
	@echo "Starting interactive shell in Docker..."
	@docker compose run --rm shell

# Clean Docker images and containers
docker-clean:
	@echo "Cleaning Docker resources..."
	@docker compose down --rmi local --volumes --remove-orphans 2>/dev/null || true
	@docker rmi plonk-test:latest 2>/dev/null || true
	@echo "Docker resources cleaned!"

# Build and run tests in Docker (one command)
docker-test-all: docker-build docker-test
