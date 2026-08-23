# Go Makefile targets

Targets common to every Go project (see the Makefile conventions fragment for structure).

## VERSION

`build` stamps `VERSION` into the binary, so it is declared at the top with the other variables:

```make
VERSION := $(shell git describe --tags --always --dirty --match 'v*' 2>/dev/null || echo dev)
```

Declaring it at all is the first half. Make expands an undefined `VERSION` to the empty string rather than complaining, so `-X <module>/cmd.Version=` stamps an empty version. That is not the `"dev"` the version block falls back from, so it wins outright and the build-info recovery never runs (see the scaffolding fragment). Cobra then registers no `--version` flag, because it only adds one when `Version` is non-empty:

```
$ ./mytool version
                                     # empty
$ ./mytool --version
mytool: unknown flag: --version
```

`make build` succeeds in both cases, so nothing reports it.

Each flag is doing a job. `--always` falls back to a short commit hash where no tag is reachable and `|| echo dev` covers a tree that is not a git checkout at all, which together are what keep the value from ever being empty. `--dirty` marks a build made from uncommitted changes.

`--match 'v*'` is the one that is easy to leave off and wrong to. The prerelease process publishes its channel under a moving `dev` tag (see the release fragment), so as soon as a developer has fetched tags, an unmatched `git describe` names that tag instead of the last release:

```
$ git describe --tags --always --dirty              # dev tag fetched
dev-dirty
$ git describe --tags --always --dirty --match 'v*'
v1.2.3-1-ga2ee320-dirty
```

This is the same trap `git.ignore_tags` covers on the goreleaser side, and it has to be closed in both places.

- `format`: `$(GOIMPORTS) -w .`, where `GOIMPORTS := go run golang.org/x/tools/cmd/goimports@latest -local <module>`. It formats exactly as `gofmt` does and groups imports as well (see the style fragment).
- `check_format`: capture `$(GOIMPORTS) -l .` and fail if non-empty: `out="$$($(GOIMPORTS) -l .)"; test -z "$$out"`. Do not write `... -l . && git diff --exit-code`: `-l` never changes files and exits 0 whatever it finds, so that form can never fail.
- `check_mod`: `go mod tidy && git diff --exit-code go.mod go.sum`
- `vet`: `go vet ./...` followed by `go vet -tags=integration ./...`. The second run is what compiles the integration files; without it a tagged test can stop building and nothing notices (see the integration testing fragment).
- `check_cross`: `GOOS=windows go vet ./...` then `GOOS=darwin go vet ./...`. Type-checks the platform-specific files a Linux runner otherwise never sees; see the workflows fragment. Use `go vet` rather than `go build`: `go build` skips test files, so a test behind `//go:build windows` can stop compiling with nothing to report it.
- `test`: `go test -race -count=1 ./...`. Build-tagged integration tests are excluded automatically.
- `test_integration`: `go test -race -count=1 -tags=integration ./...`. Never run in CI.
- `test_coverage`: run `go test -race -count=1 -tags=integration -coverpkg=./internal/... -coverprofile=coverage.out ./...`, then `go tool cover -func=coverage.out` to print the per-function table ending in the aggregate `total:` line, then `rm coverage.out`.
- `build`: `go build -ldflags="-s -w -X <module>/cmd.Version=$(VERSION)" -o dist/<binary> .`
- `snapshot`: `goreleaser release --snapshot --clean`
- `check_embed`: validate embedded content if the binary embeds any (see the tooling fragment); omit for a project with nothing embedded.
- `vuln`: `go run golang.org/x/vuln/cmd/govulncheck@latest ./...`. Deliberately not a prerequisite of `ci`: it needs network access and answers a question that is not about this commit, so it runs on a schedule instead (see the tooling and workflows fragments).
- `check_all`: `check_format check_mod vet check_cross`, plus `check_embed` where the project embeds anything. This is the only place the static checks are listed. `lint.yml` calls it and `ci` composes it, so there is no second copy to fall out of step.
- `ci`: `check_all test`
- `clean`: `rm -rf dist/ coverage.out` plus the release artefacts gpipe writes into the repository root; see the release fragment for the full list.
- `get_version`: `@echo "$(VERSION)"`. Prints the version `build` would stamp, so it can be checked before a release is cut rather than read back out of the binary afterwards.

The testing fragments own the rules these recipes implement: `-race -count=1` on every run, coverage over `./internal/...` only, and coverage including `-tags=integration`. One detail belongs here, because it is about reading the output rather than choosing the flags. The per-package percentages `go test` prints are each measured against the whole `-coverpkg` set, so they read low and do not sum; the real figure is the `total:` line from `go tool cover -func`, which is also the number used for the coverage badge.

## get_changelog

Prints the `CHANGELOG.md` section for one release to stdout. Git tags are `v`-prefixed (`v1.2.3`) but changelog headers are bare (`## 1.2.3 - ...`, see the changelog fragment), so the target strips a leading `v` from `TAG` before matching. It exits non-zero when `TAG` is empty or no entry matches, so a release never publishes empty notes. Use this implementation verbatim rather than rewriting the extraction per project:

```make
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

The `END` block trims blank lines from both ends of the captured section, so the release body starts at the first heading rather than an empty line. Matching is anchored with `index($$0,"## "tag" ")==1` rather than a regex, so `1.2` never matches the `1.2.3` header. It needs `SHELL := /bin/bash` for `[[`, which the Makefile conventions already require.
