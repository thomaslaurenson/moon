# Makefile conventions

Language-agnostic Makefile conventions. Language-specific targets live in the relevant language fragment.

- `make` is a task runner, not a build system (unless the project has no better option).
- A target name reads as an instruction: `build`, `test`, `format`, `check_mod`. Name the action, not the artefact it leaves behind. Where a short established name is clearer than a manufactured verb, use it (`ci`, `clean`, `help`, `snapshot`, `vuln`) rather than inventing `run_ci` or `scan_vulnerabilities`.
- A workflow step calls `make <target>` when the step is useful locally, appears in more than one workflow, or contains non-trivial logic. A one-line `gh release create` or `docker push` that only ever runs in CI stays in the workflow.
- Target names use underscores: `check_format`, `test_coverage`.
- Related targets share a prefix, so tab completion lists the family: `check_format`, `check_mod`, `check_cross`; `get_changelog`, `get_version`. A family prefix is never also a bare target, since completing it would stop at the bare one instead of offering the family. `test` is the one deliberate exception, because `make test` is a convention across every ecosystem and worth more than the ambiguity.
- Keep lines to 100 characters.

Non-negotiable:

- `help` must be the first target, and the default when `make` runs with no arguments.
- Every user-facing target uses inline `##` help text on the target line. No manual `echo` help blocks.
- `.PHONY` is declared directly above each target, never as one top-level list.

```makefile
SHELL := /bin/bash

##@ BUILD

.PHONY: help
help: ## Show this help message
	@awk 'BEGIN {FS = ":.*?## "} /^##@ / {printf "\n%s\n", substr($$0, 5)} \
		/^[a-zA-Z_-]+:.*## / {printf "  %-18s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
```

The help target renders the `##@` section markers as menu headings, so the grouping a reader sees in the file is the grouping a user sees from `make help`.

- Set `SHELL := /bin/bash` at the top, before variables. Do not use `.SILENT:`; use `@` selectively.
- Declare variables in `UPPER_SNAKE_CASE` at the top, `:=` by default, `?=` only for command-line overrides.
- A shell expansion in a recipe needs `$$`: make consumes a single `$` before the shell ever sees it. `out="$$(cmd)"; test -z "$$out"` is a shell command substitution; `out="$(cmd)"` is make expanding a variable or function named `cmd`, which yields the empty string and a test that no longer means anything. A `$` intended for make, such as `$(MAKEFILE_LIST)` or a `$(VERSION)` declared at the top, stays single.
- Mark each logical group with a `##@` section line (`##@ BUILD`, `##@ TEST`, `##@ LINT`, `##@ GET`, `##@ CI`), which the help target prints as headings. Order groups by how often a developer reaches for them: BUILD, then TEST, then LINT, with GET and CI last. Omit empty sections.
- Include a `ci` target so a developer can run everything CI runs in one command before pushing, and a `clean` target after it. The language fragment defines both, since their recipes are language-specific. `ci` is prerequisites only, and it composes the same aggregate targets the workflows call rather than restating their contents, so the two cannot drift. It reproduces the checks rather than any matrix CI runs them under.
- All version and changelog extraction goes through `##@ GET` targets (`get_changelog`, `get_version`), so workflows never embed raw bash or awk.
