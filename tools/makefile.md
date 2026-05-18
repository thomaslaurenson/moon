# Makefile Conventions

Conventions for writing `Makefile` files across all projects.

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
- Do **not** use `.SILENT:`; commands should echo to the terminal by default so failures are diagnosable without re-running
- Use `@` selectively to suppress noisy but unimportant lines (e.g. `@mkdir -p ...`, `@echo "done"`); never suppress the main command of a recipe

## Variables

```makefile
BINARY  := myapp
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

- Keeps the declaration co-located with the target; adding or removing a target is fully self-contained
- Eliminates the central list maintenance burden; the list can never drift out of sync
- The `help` target already serves as a table of contents for the file

## Help Target

`help` must always be the first target so it is the default when `make` is run with no arguments:

```makefile
help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  %-18s %s\n", $$1, $$2}'
```

- Target comment syntax: `target: ## Short description` - comment on the same line, after `##`
- Every user-facing target must include this inline comment format
- Targets without inline comments are treated as a spec violation unless intentionally hidden aliases
- Adjust the `%-18s` padding to fit the longest target name in the file
- Help output should fit in a terminal without wrapping (target + description <= 100 chars)

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
# GET
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
- Lint/format: `fmt`, `fmt_check`, `lint`, `mod_check`, `vet`, `fix`
- Test: `test`, `test_verbose`, `test_coverage`
- Aggregation: `ci`
- Cleanup: `clean`

Language-specific command choices are project-specific, but structure is stable:

- Go projects typically use `gofmt`, `go vet`, `go test -race -count=1`, and `go mod tidy`
- Python projects typically use `uv`, `ruff`, and `pytest`
- C++ projects typically use `clang-format-18`, `clang-tidy-18`, and `ctest`

The mandatory behaviour is defined by earlier sections in this file:

- Help generation from inline `##` comments
- Per-target `.PHONY` declarations
- Section separators

---

## GET Section

All project Makefiles must include a `# GET` section. GitHub Actions workflows call these targets instead of embedding raw bash or awk scripts in workflow files.

### Python projects

```makefile
# GET

.PHONY: get_python_project_version
get_python_project_version: ## Print the project version from pyproject.toml
	python3 -c "import tomllib, pathlib; print(tomllib.loads(pathlib.Path('pyproject.toml').read_text())['project']['version'])"

.PHONY: get_python_required_version
get_python_required_version: ## Print the required Python version from pyproject.toml
	grep -oP 'requires-python.*>=\K[0-9.]+' pyproject.toml

.PHONY: get_changelog_entry
get_changelog_entry: ## Extract the changelog entry for the current version to /tmp/release_notes.md
	@VERSION=$$(python3 -c "import tomllib, pathlib; print(tomllib.loads(pathlib.Path('pyproject.toml').read_text())['project']['version'])"); \
	awk -v ver="$$VERSION" ' \
	  /^## / { if (found) exit; if (index($$0, "## " ver " ") || $$0 == "## " ver) { found=1 } next } \
	  found { lines[n++] = $$0 } \
	  END { \
	    s=0; while (s < n && lines[s] ~ /^[[:space:]]*$$/) s++; \
	    e=n-1; while (e >= s && lines[e] ~ /^[[:space:]]*$$/) e--; \
	    for (i=s; i<=e; i++) print lines[i] \
	  } \
	' CHANGELOG.md > /tmp/release_notes.md; \
	test -s /tmp/release_notes.md || { echo "No CHANGELOG entry found for $$VERSION" >&2; exit 1; }
```

### C++ projects

C++ project version is declared in the root `CMakeLists.txt` via the `project()` `VERSION` parameter. The `get_changelog_entry` target reads the version directly from `CMakeLists.txt`:

```makefile
# GET

.PHONY: get_project_version
get_project_version: ## Print the project version from CMakeLists.txt
	@grep -oP 'project\([^)]*VERSION \K[0-9]+\.[0-9]+\.[0-9]+' CMakeLists.txt

.PHONY: get_changelog_entry
get_changelog_entry: ## Extract the changelog entry for the current version to /tmp/release_notes.md
	@VERSION=$$(grep -oP 'project\([^)]*VERSION \K[0-9]+\.[0-9]+\.[0-9]+' CMakeLists.txt); \
	awk -v ver="$$VERSION" ' \
	  /^## / { if (found) exit; if (index($$0, "## " ver " ") || $$0 == "## " ver) { found=1 } next } \
	  found { lines[n++] = $$0 } \
	  END { \
	    s=0; while (s < n && lines[s] ~ /^[[:space:]]*$$/) s++; \
	    e=n-1; while (e >= s && lines[e] ~ /^[[:space:]]*$$/) e--; \
	    for (i=s; i<=e; i++) print lines[i] \
	  } \
	' CHANGELOG.md > /tmp/release_notes.md; \
	test -s /tmp/release_notes.md || { echo "No CHANGELOG entry found for $$VERSION" >&2; exit 1; }
```

- The version regex matches the `project(MyApp VERSION 1.2.3)` pattern in `CMakeLists.txt`. If the `project()` call spans multiple lines, ensure `VERSION` appears on the same line as the version number.
- The awk logic is identical across Python and C++; only the version source differs. This ensures changelog extraction behaviour is consistent.
- The target exits non-zero if no matching entry is found, which causes CI to fail fast before attempting a release with an empty changelog.

### Go projects

Go project version comes from the git tag set by goreleaser. The `get_changelog_entry` target accepts `TAG` on the command line and outputs to stdout. The calling workflow redirects to a temp file.

```makefile
TAG ?= $(shell git describe --tags --abbrev=0 2>/dev/null)

# GET

.PHONY: get_changelog_entry
get_changelog_entry: ## Print release notes for TAG to stdout (override with TAG=v1.0.0)
	@tag="$(TAG)"; tag="$${tag#v}"; \
	if [[ -z "$$tag" ]]; then \
	  printf 'get_changelog_entry: TAG is empty; pass TAG=v1.0.0 or create a git tag\n' >&2; \
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
	  printf 'get_changelog_entry: no CHANGELOG entry for %s\n' "$$tag" >&2; \
	  exit 1; \
	fi; \
	printf '%s\n' "$$notes"
```

- Same pattern as Bash/Shell projects: outputs to stdout and accepts `TAG` on the command line.
- The calling release workflow redirects: `make get_changelog_entry TAG=${GITHUB_REF_NAME} > /tmp/release-notes.md`.
- Strip the `v` prefix in the recipe (`$${tag#v}`) because git tags use `v1.0.0` but CHANGELOG entries use bare versions (`1.0.0`).

### Bash/Shell projects

Bash/shell projects version from a source file (e.g. `VERSION="0.2.3"` in `src/app.bash`). The `get_changelog_entry` target prints to stdout rather than a temp file, making it composable: callers can pipe or redirect as needed.

The target accepts `TAG` on the command line (`make get_changelog_entry TAG=v1.0.0`) and defaults to the latest git tag. The `v` prefix is stripped before looking up the changelog entry.

```makefile
TAG ?= $(shell git describe --tags --abbrev=0 2>/dev/null)

# GET

.PHONY: get_version
get_version: ## Print the project version from src/app.bash
	@grep -oE 'VERSION="[0-9]+\.[0-9]+\.[0-9]+"' src/app.bash | sed 's/VERSION="//;s/"//'

.PHONY: get_changelog_entry
get_changelog_entry: ## Print release notes for TAG to stdout (override with TAG=v1.0.0)
	@tag="$(TAG)"; tag="$${tag#v}"; \
	if [[ -z "$$tag" ]]; then \
	  printf 'get_changelog_entry: TAG is empty; pass TAG=v1.0.0 or create a git tag\n' >&2; \
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
	  printf 'get_changelog_entry: no CHANGELOG entry for %s\n' "$$tag" >&2; \
	  exit 1; \
	fi; \
	printf '%s\n' "$$notes"
```

The key differences from the Python/C++ pattern:

- Outputs to stdout, not `/tmp/release_notes.md`. The calling workflow redirects as needed: `make get_changelog_entry TAG=v1.0.0 > /tmp/release-notes.md`.
- `TAG ?=` defaults to the latest git tag via `git describe`. Use `$(TAG)` (not `$${TAG}`) in the recipe so Make's variable expansion is used; shell does not inherit Make variables set via `?=`.
- Strip the `v` prefix in the recipe (`$${tag#v}`) because git tags use `v1.0.0` but CHANGELOG entries use bare versions (`1.0.0`).
- The target exits non-zero if TAG is empty or no matching entry is found.
- Release artifact steps (tarball creation, version patching, checksum generation) are CI-only and do not belong in the Makefile. Keep them inline in the release workflow.
- Avoid `grep -oP` (GNU-only, unavailable on macOS). Use `grep -oE` or `sed -n` instead. When matching a prefix like `readonly VERSION=`, anchor with `.*VERSION=` rather than `^VERSION=`.
