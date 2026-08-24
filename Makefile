SHELL := /bin/bash

BINARY  := moon
VERSION := $(shell git describe --tags --always --dirty --match 'v*' 2>/dev/null || echo dev)
LDFLAGS := -s -w -X github.com/thomaslaurenson/moon/cmd.Version=$(VERSION)

# BUILD
.PHONY: help
help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  %-16s %s\n", $$1, $$2}'

.PHONY: build
build: ## Build the moon binary into dist/ (embeds src/fragments and src/bundles)
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

.PHONY: vuln
vuln: ## Scan for known vulnerabilities reachable from this code
	@go run golang.org/x/vuln/cmd/govulncheck@latest ./...

# TEST
.PHONY: test
test: ## Run tests with the race detector
	@go test -race -count=1 ./...

.PHONY: test_coverage
test_coverage: ## Run tests with a coverage report (internal/ only; cmd/ is wiring)
	@go test -race -count=1 -coverpkg=./internal/... -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out
	@rm coverage.out

.PHONY: check
check: ## Validate every bundle: missing fragments, include cycles, orphans
	@go run . check

# GET
.PHONY: get_changelog
get_changelog: ## Print release notes for TAG to stdout (TAG=v1.0.0)
	@tag="$(TAG)"; tag="$${tag#v}"; \
	if [[ -z "$$tag" ]]; then \
	  printf 'get_changelog: TAG is empty; pass TAG=v1.0.0\n' >&2; \
	  exit 1; \
	fi; \
	notes="$$(awk -v tag="$$tag" ' \
	  /^## / { if (found) exit; if (index($$0,"## "tag" ")==1 || $$0=="## "tag) found=1; next } \
	  found { lines[n++]=$$0 } \
	  END { \
	    s=0; while (s<n && lines[s]~/^[[:space:]]*$$/) s++; \
	    e=n-1; while (e>=s && lines[e]~/^[[:space:]]*$$/) e--; \
	    for (i=s;i<=e;i++) print lines[i] \
	  }' CHANGELOG.md)"; \
	if [[ -z "$$notes" ]]; then \
	  printf 'get_changelog: no CHANGELOG entry for %s\n' "$$tag" >&2; \
	  exit 1; \
	fi; \
	printf '%s\n' "$$notes"

.PHONY: get_version
get_version: ## Print the version that would be baked into the binary
	@echo "$(VERSION)"

# CI
.PHONY: ci
ci: fmt_check mod_check vet test check ## Run all CI checks

.PHONY: clean
clean: ## Remove the binary and generated bundles
	@rm -rf dist install.sh install.ps1 checksums.txt checksums.txt.sigstore.json coverage.out
	@printf '[*] cleaned\n'
