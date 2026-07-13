SHELL := /bin/bash

BINARY  := moon
VERSION := $(shell git describe --tags --always 2>/dev/null || echo dev)
LDFLAGS := -s -w -X github.com/thomaslaurenson/moon/cmd.Version=$(VERSION)

# BUILD

.PHONY: help
help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  %-16s %s\n", $$1, $$2}'

.PHONY: build
build: ## Build the moon binary into dist/ (embeds src/ and bundles/)
	@go build -ldflags="$(LDFLAGS)" -o dist/$(BINARY) .
	@printf '[*] built dist/%s (version %s)\n' "$(BINARY)" "$(VERSION)"

.PHONY: run
run: ## Run without building, e.g. make run ARGS="show python-lib"
	@go run . $(ARGS)

.PHONY: snapshot
snapshot: ## Build a full local snapshot release with goreleaser (no publish)
	@goreleaser release --snapshot --clean

# LINT

.PHONY: fmt
fmt: ## Format all Go source
	@gofmt -w .

.PHONY: fmt_check
fmt_check: ## Fail if any file is not gofmt-clean
	@out="$$(gofmt -l .)"; test -z "$$out" || { printf 'not formatted:\n%s\n' "$$out"; exit 1; }

.PHONY: mod_check
mod_check: ## Fail if go.mod/go.sum are not tidy
	@go mod tidy
	@git diff --exit-code -- go.mod go.sum || { printf 'go.mod/go.sum not tidy; commit the diff\n' >&2; exit 1; }

.PHONY: vet
vet: ## Run go vet
	@go vet ./...

# TEST

.PHONY: test
test: ## Run tests with the race detector
	@go test -race -count=1 ./...

.PHONY: test_coverage
test_coverage: ## Run tests with a coverage report (internal/ only; cmd/ is wiring)
	@go test -race -count=1 -coverpkg=./internal/... -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out | tail -1

.PHONY: check
check: ## Validate every recipe: missing fragments, include cycles, orphans
	@go run . check

# GET

.PHONY: get_changelog
get_changelog: ## Print release notes for a tag: make get_changelog TAG=v1.2.3
	@[[ -n "$(TAG)" ]] || { printf 'Usage: make get_changelog TAG=v1.2.3\n' >&2; exit 1; }
	@v="$${TAG#v}"; \
	notes="$$(awk -v ver="$$v" ' \
		$$0 ~ "^## " ver "( |$$)" { found=1; next } \
		found && /^## / { exit } \
		found { print } \
	' CHANGELOG.md)"; \
	[[ -n "$$(printf '%s' "$$notes" | tr -d '[:space:]')" ]] \
		|| { printf 'no changelog entry for %s\n' "$$TAG" >&2; exit 1; }; \
	printf '%s\n' "$$notes"

.PHONY: get_version
get_version: ## Print the version that would be baked into the binary
	@echo "$(VERSION)"

# CI

.PHONY: ci
ci: fmt_check mod_check vet test check ## Run all CI checks

.PHONY: clean
clean: ## Remove the binary and generated bundles
	@rm -rf dist coverage.out
	@printf '[*] cleaned\n'
