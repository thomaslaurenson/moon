# Bash Makefile targets

Targets common to every maintained Bash project; see the Makefile conventions fragment for the structure they sit in and for `get_changelog`, which is the same in every language. The testing fragment owns the rules the check targets implement and the workflows fragment owns the versioning rule. This fragment is where every target is written down, so a workflow calling `make <target>` finds it defined in exactly one place.

## Variables

```makefile
SHELL := /bin/bash

MAIN    := src/app.bash
SCRIPTS := $(MAIN) src/lib/init.sh test/helpers/mock_cmd
BATS    := test/extern/bats/bin/bats
```

- `SCRIPTS` names every file `bash -n` and `shellcheck` cover. A `find` over `src/` would catch data files, and the extensionless files a Bash project ships (completions, mock helpers) are exactly the ones an extension glob misses. Naming them is what keeps the list honest, and adding a script means adding it here.
- `MAIN` is the one file that embeds `VERSION`; see the workflows fragment.

## TEST

```makefile
##@ TEST

.PHONY: test
test: ## Run the bats suite
	$(BATS) test/
```

## LINT

```makefile
##@ LINT

.PHONY: check_syntax
check_syntax: ## Fail if any script does not parse
	@for f in $(SCRIPTS); do bash -n "$$f" || exit 1; done

.PHONY: check_lint
check_lint: ## Run shellcheck over every script
	shellcheck $(SCRIPTS)

.PHONY: check_version
check_version: ## Fail if the newest changelog entry disagrees with the embedded version
	@v="$$($(MAKE) -s get_version)"; c="$$(awk '/^## /{print $$2; exit}' CHANGELOG.md)"; \
	  [[ "$$v" == "$$c" ]] || \
	  { printf '[!] version %s but the changelog says %s\n' "$$v" "$$c" >&2; exit 1; }

.PHONY: check_version_tag
check_version_tag: ## Fail unless the embedded version matches TAG (TAG=vX.Y.Z)
	@test -n "$(TAG)" || { printf 'check_version_tag: TAG is required\n' >&2; exit 2; }
	@v="$$($(MAKE) -s get_version)"; tag="$(TAG)"; tag="$${tag#v}"; \
	  [[ "$$v" == "$$tag" ]] || \
	  { printf '[!] version %s but tag %s\n' "$$v" "$$tag" >&2; exit 1; }

.PHONY: bump_bats
bump_bats: ## Move the bats submodule to a release tag (TAG=v1.11.0)
	@test -n "$(TAG)" || { printf 'bump_bats: TAG is required\n' >&2; exit 2; }
	git -C test/extern/bats fetch --tags && git -C test/extern/bats checkout "$(TAG)"
```

- `check_syntax` runs first in `check_all` because it is fast and a parse error makes every later finding noise. `shellcheck` reads the shebang for the dialect, so the list needs no `-s bash` (see the testing fragment).
- `check_version` compares the embedded version with the newest changelog heading. A project whose version appears elsewhere as well, in a man page for instance, adds one comparison per file to the same target.
- `check_version_tag` is not in `check_all`: it needs a `TAG` that only the release workflow has, which runs it before building anything (see the workflows fragment).
- `bump_bats` is how the submodule moves, never Dependabot: the `gitsubmodule` ecosystem follows branch commits, not releases (see the testing fragment).

## GET

```makefile
##@ GET

.PHONY: get_version
get_version: ## Print the version embedded in the main script (fails if absent)
	@v="$$(sed -n 's/^readonly VERSION="\(.*\)"$$/\1/p' $(MAIN))"; \
	  test -n "$$v" && printf '%s\n' "$$v"
```

- `get_version` reads the one `readonly VERSION="X.Y.Z"` line the workflows fragment requires, and fails when it finds none rather than handing a release an empty version to tag.

## CI

```makefile
##@ CI

.PHONY: check_all
check_all: check_syntax check_lint check_version ## Run every static check

.PHONY: ci
ci: check_all test ## Run the checks CI runs

.PHONY: clean
clean: ## Remove release artefacts
	rm -rf dist
```

- `check_all` is the only place the static checks are listed: `lint.yml` calls it and `ci` composes it, so there is no second copy to fall out of step.
- `clean` removes `dist/`, where the release targets assemble their tarballs and checksums (see the workflows fragment). A project that does not release has nothing to remove and omits the target.
