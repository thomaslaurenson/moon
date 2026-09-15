# Makefile conventions

Language-agnostic Makefile conventions. Language-specific targets live in the relevant language fragment.

- `make` is a task runner, not a build system (unless the project has no better option).
- A target name reads as an instruction: `build`, `test`, `format`, `check_mod`. Name the action, not the artefact it leaves behind. Where a short established name is clearer than a manufactured verb, use it (`ci`, `clean`, `help`, `snapshot`, `vuln`) rather than inventing `run_ci` or `scan_vulnerabilities`.
- A workflow step calls `make <target>` when the step is useful locally, appears in more than one workflow, or contains non-trivial logic. A one-line `gh release create` or `docker push` that only ever runs in CI stays in the workflow.
- Target names use underscores: `check_format`, `test_coverage`.
- Related targets share a prefix, so tab completion lists the family: `check_format`, `check_mod`, `check_cross`; `get_changelog`, `get_version`. A bare target that is also a family prefix is the family's everyday member and nothing else: `test` runs the unit layer beside `test_coverage`, and `configure` and `build` act on the everyday tree beside `configure_lint` and `build_fuzz`. Any other prefix stays a prefix, since completing it would otherwise stop at the bare target instead of offering the family.
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
- Mark each logical group with a `##@` section line (`##@ BUILD`, `##@ TEST`, `##@ LINT`, `##@ GENERATE`, `##@ GET`, `##@ CI`), which the help target prints as headings. Order groups by how often a developer reaches for them: BUILD, then TEST, then LINT, with GENERATE, GET and CI last. Omit empty sections.
- Include a `ci` target so a developer can run everything CI runs in one command before pushing, and a `clean` target after it. The language fragment defines both, since their recipes are language-specific. `ci` is prerequisites only, and it composes the same aggregate targets the workflows call rather than restating their contents, so the two cannot drift. It reproduces the checks rather than any matrix CI runs them under.
- All version and changelog extraction goes through `##@ GET` targets (`get_changelog`, `get_version`), so workflows never embed raw bash or awk.
- `##@ GENERATE` holds the targets that rewrite checked-in files from a source outside the repository: headers generated from a schema, a golden manifest regenerated from real data, a fixture rebuilt by a reference tool. Name them `gen_<what>`. A person runs them deliberately and commits the result; CI never does, because the committed copy is what the build and the tests read. The line against the neighbouring sections is what the target leaves behind: GENERATE writes tracked files, BUILD writes into `build/`, and GET only prints.

## get_changelog

`get_version` differs by language, since each keeps its version somewhere else. `get_changelog` does not: every release workflow prints the `CHANGELOG.md` entry for the tag it is publishing, git tags are `v`-prefixed (`v1.2.3`) while changelog headers are bare (`## 1.2.3 - ...`, see the changelog fragment), and the recipe is the same everywhere, so it is written down once here. It strips a leading `v` from `TAG` before matching, and exits non-zero when `TAG` is empty or no entry matches, so a release never publishes empty notes. Use it verbatim:

```makefile
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
```

The `END` block trims blank lines from both ends of the captured section, so the release body starts at the first heading rather than an empty line. Matching is anchored with `index($$0,"## "tag" ")==1` rather than a regex, so `1.2` never matches the `1.2.3` header. It prints the body without its `## X.Y.Z` header, because the release title already shows the version. It needs `SHELL := /bin/bash` for `[[`, which this fragment already requires.
