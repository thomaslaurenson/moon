# Makefile Conventions

Conventions for writing `Makefile` files across all projects.

## Design Principles

- `make` is a task runner — not a build system (unless the project has no better option)
- Every target is a verb: `build`, `test`, `lint`, not `binary`, `tests`, `linter`
- CI steps call `make <target>` — never raw commands directly in workflow files
- No release or tag targets — those are handled externally via shell functions
- Target names use underscores: `fmt_check`, `test_coverage`, `build_clean`
- Keep lines to 100 characters

## Global Settings

Always set these at the top of every Makefile, before any variables:

```makefile
SHELL := /bin/bash
.SILENT:
```

- `SHELL := /bin/bash` — ensures consistent shell behaviour regardless of the invoking environment
- `.SILENT:` (no prerequisites) — globally suppresses recipe echoing; equivalent to `@` on every line; use `printf` or `echo` explicitly when output is needed

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

Declare one `.PHONY` block at the top of the file, listing every target, in the same order they appear in the file, with line breaks between sections:

```makefile
.PHONY: help \
	build install \
	fmt fmt_check mod_check vet \
	test test_verbose test_coverage \
	ci \
	clean
```

- Always declare every target — eliminates silent no-ops when a directory with the same name exists
- Order matches file order
- Use `\` continuation with one tab of indent per group

## Help Target

`help` must always be the first target so it is the default when `make` is run with no arguments:

```makefile
help: ## Show this help message
	grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  %-18s %s\n", $$1, $$2}'
```

- Target comment syntax: `target: ## Short description` — comment on the same line, after `##`
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
- Only include sections that have targets — omit empty sections

## ci Target

Include a `ci` target that mirrors what the `lint.yml` and `test.yml` workflows run. Use it to validate locally before pushing:

```makefile
ci: lint test ## Run all CI checks locally
```

- The exact dependencies vary per project — match what the CI workflows actually run
- Place `ci` just before `clean`, after the test section

## Example (Go)

```makefile
SHELL := /bin/bash
.SILENT:

BINARY  := myapp
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X github.com/example/myapp/cmd.Version=$(VERSION)

.PHONY: help \
	build install \
	fmt fmt_check mod_check vet \
	test test_verbose test_coverage \
	ci \
	clean

help: ## Show this help message
	grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  %-18s %s\n", $$1, $$2}'

# BUILD
build: ## Build the binary
	go build -ldflags="$(LDFLAGS)" -o bin/$(BINARY) .

install: ## Install to GOPATH/bin
	go install -ldflags="$(LDFLAGS)" .

# LINT
fmt: ## Format all Go source files
	gofmt -w .

fmt_check: ## Check formatting without writing
	unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		printf 'Unformatted files:\n%s\n' "$$unformatted"; \
		exit 1; \
	fi

mod_check: ## Check go.mod and go.sum are tidy
	go mod tidy
	git diff --exit-code go.mod go.sum

vet: ## Run go vet
	go vet ./...

# TEST
test: ## Run all tests
	go test -race -count=1 ./...

test_verbose: ## Run all tests with verbose output
	go test -race -count=1 -v ./...

test_coverage: ## Run tests and print coverage
	go test -race -count=1 -coverpkg=./internal/... -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	rm coverage.out

ci: fmt_check mod_check vet test ## Run all CI checks locally

# TASKS
clean: ## Remove build artifacts
	rm -rf bin/ dist/
```
