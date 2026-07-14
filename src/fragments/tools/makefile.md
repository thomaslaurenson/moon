# Makefile conventions

Language-agnostic Makefile conventions. Language-specific targets live in the relevant language fragment.

- `make` is a task runner, not a build system (unless the project has no better option).
- Every target is a verb (`build`, `test`, `lint`), never a noun.
- CI steps call `make <target>`, never raw commands.
- Target names use underscores: `fmt_check`, `test_coverage`.
- Keep lines to 100 characters.

Non-negotiable:

- `help` must be the first target, and the default when `make` runs with no arguments.
- Every user-facing target uses inline `##` help text on the target line. No manual `echo` help blocks.
- `.PHONY` is declared directly above each target, never as one top-level list.

```makefile
SHELL := /bin/bash

# BUILD

.PHONY: help
help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  %-18s %s\n", $$1, $$2}'
```

- Set `SHELL := /bin/bash` at the top, before variables. Do not use `.SILENT:`; use `@` selectively.
- Declare variables in `UPPER_SNAKE_CASE` at the top, `:=` by default, `?=` only for command-line overrides.
- Use a comment separator before each logical group (`# BUILD`, `# LINT`, `# TEST`, `# GET`). Omit empty sections.
- Include a `ci` target mirroring what the lint and test workflows run, placed just before `clean`.
- All version and changelog extraction goes through `# GET` targets (`get_changelog`, `get_version`), so workflows never embed raw bash or awk.
