# Makefile Conventions

Conventions for writing `Makefile` files across all projects.

## Design Principles

- `make` is a task runner - not a build system (unless the project has no better option)
- Every target is a verb: `build`, `test`, `lint`, not `binary`, `tests`, `linter`
- CI steps call `make <target>` - never raw commands directly in workflow files
- No release or tag targets - those are handled externally via shell functions
- Target names use underscores: `fmt_check`, `test_coverage`, `build_clean`
- Keep lines to 100 characters

## Global Settings

Always set these at the top of every Makefile, before any variables:

```makefile
SHELL := /bin/bash
```

- `SHELL := /bin/bash` - ensures consistent shell behaviour regardless of the invoking environment
- Do **not** use `.SILENT:` — commands should echo to the terminal by default so failures are diagnosable without re-running
- Use `@` selectively to suppress noisy but unimportant lines (e.g. `@mkdir -p ...`, `@echo "done"`) — never suppress the main command of a recipe

## Variables

```makefile
BINARY  := narc
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X github.com/example/repo/cmd.Version=$(VERSION)

INSTALL_DIR ?= /usr/local/bin
```

- `UPPER_SNAKE_CASE` for all variable names
- Declare all variables at the top of the file, before `.PHONY` and targets
- Use `:=` (immediate assignment) by default
- Use `?=` only when the variable is intentionally overridable from the command line (e.g. `make install INSTALL_DIR=/custom`)
- Align `:=` and `?=` operators with spaces for readability

## .PHONY

Declare `.PHONY` immediately before each target, not in a single block at the top:

```makefile
.PHONY: help
help: ## Show this help message
	...

.PHONY: build
build: ## Build the binary
	...
```

- Keeps the declaration co-located with the target — adding or removing a target is fully self-contained
- Eliminates the central list maintenance burden; the list can never drift out of sync
- The `help` target already serves as a table of contents for the file

## Help Target

`help` must always be the first target so it is the default when `make` is run with no arguments:

```makefile
help: ## Show this help message
	grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  %-18s %s\n", $$1, $$2}'
```

- Target comment syntax: `target: ## Short description` - comment on the same line, after `##`
- Adjust the `%-18s` padding to fit the longest target name in the file
- Help output should fit in a terminal without wrapping (target + description ≤ 100 chars)

## Section Separators

Use a comment separator before each logical group of targets. Common sections:

```makefile
# BUILD
# LINT
# TEST
# DOCS
# TASKS
```

- One blank line before the separator, no blank line after
- Only include sections that have targets - omit empty sections

## ci Target

Include a `ci` target that mirrors what the `lint.yml` and `test.yml` workflows run. Use it to validate locally before pushing:

```makefile
ci: lint test ## Run all CI checks locally
```

- The exact dependencies vary per project - match what the CI workflows actually run
- Place `ci` just before `clean`, after the test section

## Example (Go)

```makefile
SHELL := /bin/bash

BINARY  := myapp
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X github.com/example/myapp/cmd.Version=$(VERSION)

.PHONY: help
help: ## Show this help message
	grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  %-18s %s\n", $$1, $$2}'

# BUILD
.PHONY: build
build: ## Build the binary
	go build -ldflags="$(LDFLAGS)" -o bin/$(BINARY) .

.PHONY: install
install: ## Install to GOPATH/bin
	go install -ldflags="$(LDFLAGS)" .

# LINT
.PHONY: fmt
fmt: ## Format all Go source files
	gofmt -w .

.PHONY: fmt_check
fmt_check: ## Check formatting without writing
	unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		printf 'Unformatted files:\n%s\n' "$$unformatted"; \
		exit 1; \
	fi

.PHONY: mod_check
mod_check: ## Check go.mod and go.sum are tidy
	go mod tidy
	git diff --exit-code go.mod go.sum

.PHONY: vet
vet: ## Run go vet
	go vet ./...

# TEST
.PHONY: test
test: ## Run all tests
	go test -race -count=1 ./...

.PHONY: test_verbose
test_verbose: ## Run all tests with verbose output
	go test -race -count=1 -v ./...

.PHONY: test_coverage
test_coverage: ## Run tests and print coverage
	go test -race -count=1 -coverpkg=./internal/... -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	rm coverage.out

.PHONY: ci
ci: fmt_check mod_check vet test ## Run all CI checks locally

# TASKS
.PHONY: clean
clean: ## Remove build artifacts
	rm -rf bin/ dist/
```

## Example (Python / uv)

```makefile
SHELL := /bin/bash

.PHONY: help
help: ## Show this help message
	grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  %-18s %s\n", $$1, $$2}'

# DEV
.PHONY: install
install: ## Install dependencies
	uv sync

.PHONY: update
update: ## Upgrade all locked dependencies
	uv lock --upgrade
	uv sync

# LINT
.PHONY: lint
lint: ## Check code with ruff
	uv run ruff check .

.PHONY: lint_fix
lint_fix: ## Auto-fix lint issues
	uv run ruff check --fix .

.PHONY: format
format: ## Format code with ruff
	uv run ruff format .

.PHONY: format_check
format_check: ## Check formatting without writing changes
	uv run ruff format --check .

# TEST
.PHONY: test
test: ## Run unit tests
	uv run pytest

.PHONY: test_coverage
test_coverage: ## Run unit tests with coverage report
	uv run coverage run -m pytest
	uv run coverage report

.PHONY: ci
ci: lint format_check test ## Run all CI checks locally

# TASKS
.PHONY: clean
clean: ## Remove venv, caches, and build artifacts
	rm -rf .venv dist .pytest_cache .coverage
	find . -type d -name '__pycache__' -exec rm -rf '{}' +
```
