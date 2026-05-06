# Repository Refactoring Specification for VS Code Copilot

**Context for Copilot:** You are tasked with refactoring the AI guideline documentation in this repository. Follow these exact instructions to trim token bloat, remove contradictions, and improve the logic for agentic coding. Execute these changes file by file.

## Phase 1: Global Document Cleanup (All `.md` files)
1. **Remove Tables of Contents:** Scan all markdown files in the repository. Delete any `## Contents` sections and their associated link lists.
2. **Remove Cross-References:** Delete any explicit markdown cross-references or "Routing" instructions (e.g., "See `python/testing.md` for details") from all files.
3. **Keep Existing Structure:** Do not delete or consolidate files unless explicitly instructed in Phase 2. Keep standard language idioms and duplicate rules intact exactly where they are.
4. **Update Comment Hygiene:** In any file that outlines comment style (e.g., style guides), replace mentions of "Protect human comments" with: 
   > "Preserve comments that explain the 'why' (business logic, architecture). Aggressively delete and refactor comments that narrate the 'what' (step-by-step code narration)."

## Phase 2: Git Rules Consolidation (`github/`)
1. **Merge & Delete:** - Read `github/commits.md` and `github/prs.md`.
   - Migrate all unique rules regarding branching, committing (tense/mood), and pull requests into `github/conventions.md`.
   - Delete `github/commits.md` and `github/prs.md`.

## Phase 3: Python Architecture Overhaul (`python/`)
### 1. Scaffolding Split (`python/scaffolding.md`)
Rewrite the scaffolding rules to clearly separate "Libraries" vs. "Apps/Scripts":
* **Libraries:**
  - **Testing:** Must include `pytest-cov` and `coverage`. Makefile `test` runs with coverage tracking (`pytest --cov=src`).
  - **Type Checking:** Must include `pyright`.
  - **Build System:** Must include `hatchling` (or specified build system) in `pyproject.toml`. Makefile requires a `build` target.
  - **Logging:** Strictly use standard library `logging`. Never use third-party loggers like `structlog`.
  - **Badges:** Mandate the following README badges: Build state, Release state, Release version, Release downloads, Python version (dynamically extracted from `pyproject.toml`), and Test coverage.
* **Apps/Scripts:**
  - **Testing:** Use base `pytest` only. Omit coverage dependencies and tracking.
  - **Type Checking:** Omit `pyright` entirely.
  - **Build System:** Omit build system configurations entirely (not installable).
  - **Logging:** Require structured logging using `structlog`.
  - **Badges:** Omit coverage and release badges.

### 2. Pyright Hardcoding Fix (`python/scaffolding.md`)
- Add a strict rule: *"Never hardcode `pythonVersion` in the `[tool.pyright]` config block. It must infer the version automatically from `requires-python`."*
- Ensure any example `[tool.pyright]` blocks in the documentation reflect this by removing the hardcoded `pythonVersion` key.

### 3. Testing Overhaul (`python/testing.md`)
- **Delete the Hook:** Remove any mention or code block demonstrating the `conftest.py` / `pytest_collection_modifyitems` hook used for skipping integration tests.
- **New Rule:** Explicitly state that integration tests are skipped natively via the Makefile using the flag: `pytest -m "not integration"`.

### 4. Docs Fix (`python/docs.md`)
- **Remove Fallback:** Delete the `autodoc_mock_imports` fallback entirely.
- **New Rule:** If a docs build fails due to a missing import, you must fix it by adding the required dependency to the `docs` extra group in `pyproject.toml`, never by mocking the import in Sphinx.

## Phase 4: Go Rules Update (`golang/`)
### 1. Tooling Clarification (`golang/scaffolding.md` & `golang/goreleaser.md`)
- Update the Go "Tools" or "Linters" section to strictly enforce:
  > "No third-party linters or formatters are permitted. Specifically, DO NOT use `golangci-lint` or `govulncheck` under any circumstances. However, third-party release tools like `goreleaser` and `cosign` are explicitly permitted and required."

## Phase 5: Makefiles and CI/CD Refactor
### 1. Makefile Master Structure (`tools/makefile.md`)
- Establish this file as the *single source of truth* for global Makefile structure.
- Define standard section headers: `# BUILD`, `# TEST`, `# GET`.
- Define standard global boilerplate targets (e.g., `clean`, self-documenting `help`).
- **Add the `# GET` section:** Mandate the inclusion of the following specific targets to replace CI bash scripts:
  - `get_python_project_version`: Extracts the version from `pyproject.toml`.
  - `get_python_required_version`: Extracts the required Python version from `pyproject.toml`.
  - `get_changelog_entry`: Extracts the latest changelog entry from `CHANGELOG.md`.

### 2. De-duplicate Language Makefiles (`python/scaffolding.md`, `golang/scaffolding.md`)
- Remove the massive, raw Makefile boilerplate code blocks from both of these files.
- Replace them with a simple "Command Map" instruction. Example format:
  > "Adhere to the global Makefile structure established in `tools/makefile.md`. Use the following commands for your standard targets:
  > * `build`: `uv build`
  > * `test`: `uv run pytest`"

### 3. GitHub Actions Context Relief (`github/actions.md`)
- Delete all raw, multi-line `awk` and bash script blocks used for extracting versions or changelogs.
- **New Rule:** "Never write raw bash or awk scripts in workflows to extract versions or changelogs. You must execute the repository's Makefile targets (e.g., `make get_python_project_version`, `make get_changelog_entry`) and pass their standard output to the respective GitHub Action steps."

---
**End of Specification.**
