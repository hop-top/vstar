# vstar developer build targets.
# Run from the repo root. Requires GNU make.

COVER_PROFILE := coverage.out
COVER_HTML    := coverage.html

.PHONY: help build test vet fmt fmt-check lint cover tidy fixtures-verify ci

help: ## Show available targets
	@echo "vstar targets:"
	@echo "  build            Compile all packages"
	@echo "  test             Run all tests with race detector"
	@echo "  vet              Run go vet"
	@echo "  fmt              Format with gofumpt (idempotent)"
	@echo "  fmt-check        Verify formatting without writing"
	@echo "  lint             Run golangci-lint"
	@echo "  cover            Produce coverage profile + HTML report"
	@echo "  tidy             Run go mod tidy"
	@echo "  fixtures-verify  Regenerate fixtures + fail on drift"
	@echo "  ci               Aggregate CI gate (vet + fmt-check + lint + cover + fixtures-verify)"

build: ## Compile all packages
	go build ./...

test: ## Run all tests with race detector
	go test -race ./...

vet: ## Run go vet
	go vet ./...

fmt: ## Format with gofumpt (gofmt + extras); idempotent
	gofumpt -w .

fmt-check: ## Verify formatting without writing
	@diff_out=$$(gofumpt -l .); \
	if [ -n "$$diff_out" ]; then \
		echo "gofumpt: files need formatting:" >&2; \
		echo "$$diff_out" >&2; \
		exit 1; \
	fi

lint: ## Run golangci-lint
	golangci-lint run ./...

cover: ## Produce coverage profile + HTML report
	go test -race -coverprofile=$(COVER_PROFILE) -covermode=atomic ./...
	go tool cover -html=$(COVER_PROFILE) -o $(COVER_HTML)

tidy: ## Run go mod tidy
	go mod tidy

fixtures-verify: ## Regenerate testdata canonical/hash + sync fuzz seeds; fail on drift
	go run ./cmd/fixtures-verify
	@if ! git diff --quiet -- ./testdata ./codec/rfc5545/testdata/fuzz ./codec/rfc6350/testdata/fuzz; then \
		echo "fixtures-verify: drift detected against committed corpus." >&2; \
		echo "Run 'make fixtures-verify' locally and commit the regenerated files," >&2; \
		echo "or revert the implementation change that caused the drift." >&2; \
		git --no-pager diff --stat -- ./testdata ./codec/rfc5545/testdata/fuzz ./codec/rfc6350/testdata/fuzz >&2; \
		exit 1; \
	fi

ci: vet fmt-check lint cover fixtures-verify ## Aggregate CI gate
