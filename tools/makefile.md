# Makefile Conventions

Conventions for writing `Makefile` files across all projects.

## Contents

- [Design principles](#design-principles) — make as task runner, verb targets, underscores
- [Non-negotiable rules](#non-negotiable-rules) — help first, inline comments, per-target .PHONY
- [Global settings](#global-settings) — SHELL := /bin/bash
- [Variables](#variables) — naming, assignment operators, alignment
- [.PHONY](#phony) — declare per target, not in one block
- [Help target](#help-target) — canonical grep/awk implementation
- [Section separators](#section-separators) — comment-based grouping
- [ci target](#ci-target) — mirrors CI pipeline for local validation
- [Canonical target names](#compact-reference-targets) — build, lint, test, clean reference

## Design Principles

- `make` is a task runner - not a build system (unless the project has no better option)
- Every target is a verb: `build`, `test`, `lint`, not `binary`, `tests`, `linter`
- CI steps call `make <target>` - never raw commands directly in workflow files
- No release or tag targets - those are handled externally via shell functions
- Target names use underscores: `fmt_check`, `test_coverage`, `build_clean`
- Keep lines to 100 characters

## Non-Negotiable Rules

These are hard requirements. Any generated or edited Makefile that violates them is incorrect.

- `help` must be the first target
- Every user-facing target must use inline `##` help text on the target line
- Manual help blocks using `echo` lines are not allowed
- `.PHONY` must be declared directly above each target (not in one top-level list)

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
- Every user-facing target must include this inline comment format
- Targets without inline comments are treated as a spec violation unless intentionally hidden aliases
- Adjust the `%-18s` padding to fit the longest target name in the file
- Help output should fit in a terminal without wrapping (target + description ≤ 100 chars)

Anti-pattern (do not do this):

```makefile
help:
	@echo "build ..."
	@echo "test ..."
```

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

## Compact Reference Targets

Use these canonical target names where they fit your project:

- Build: `build`, `install`
- Lint/format: `fmt`, `fmt_check`, `lint`, `mod_check`, `vet`, `lint_fix`
- Test: `test`, `test_verbose`, `test_coverage`
- Aggregation: `ci`
- Cleanup: `clean`

Language-specific command choices are project-specific, but structure is stable:

- Go projects typically use `gofmt`, `go vet`, `go test -race -count=1`, and `go mod tidy`
- Python projects typically use `uv`, `ruff`, and `pytest`

The mandatory behaviour is defined by earlier sections in this file:

- Help generation from inline `##` comments
- Per-target `.PHONY` declarations
- Section separators
- CI workflows invoking `make <target>`
