# Go Makefile targets

Targets common to every Go project (see the Makefile conventions fragment for structure).

- `fmt`: `gofmt -w .`
- `fmt_check`: capture `gofmt -l .` and fail if non-empty (`out="$(gofmt -l .)"; test -z "$out"`). Do not write `gofmt -l . && git diff --exit-code`: `gofmt -l` never changes files and always exits 0, so that form can never fail.
- `mod_check`: `go mod tidy && git diff --exit-code go.mod go.sum`
- `vet`: `go vet ./...` followed by `go vet -tags=integration ./...`. The second run is what compiles the integration files; without it a tagged test can stop building and nothing notices (see the integration testing fragment).
- `build_check`: `GOOS=windows go build ./...` then `GOOS=darwin go build ./...`. Compiles the platform-specific files that a Linux runner otherwise never sees; see the workflows fragment.
- `test`: `go test -race -count=1 ./...`. Build-tagged integration tests are excluded automatically.
- `test_integration`: `go test -race -count=1 -tags=integration ./...`. Never run in CI.
- `test_coverage`: run `go test -race -count=1 -tags=integration -coverpkg=./internal/... -coverprofile=coverage.out ./...`, then `go tool cover -func=coverage.out` to print the per-function table ending in the aggregate `total:` line, then `rm coverage.out`.
- `build`: `go build -ldflags="-s -w -X <module>/cmd.Version=$(VERSION)" -o dist/<binary> .`
- `snapshot`: `goreleaser release --snapshot --clean`
- `check`: validate embedded content if the binary embeds any (see the tooling fragment); omit for a project with nothing embedded.
- `vuln`: `go run golang.org/x/vuln/cmd/govulncheck@latest ./...`. Deliberately not a prerequisite of `ci`: it needs network access and answers a question that is not about this commit, so it runs on a schedule instead (see the tooling and workflows fragments).
- `ci`: `fmt_check mod_check vet test`
- `clean`: `rm -rf dist/ coverage.out` plus the release artefacts gpipe writes into the repository root; see the release fragment for the full list.

Tests always include `-race -count=1`, including coverage. Coverage runs with `-tags=integration` so the figure covers everything the project can exercise, which means it depends on the machine having the resources those tests need; see the integration testing fragment. Coverage is measured over `./internal/...` only; `cmd/` and the root package are excluded as wiring-only. The per-package percentages `go test` prints are each measured against the whole `-coverpkg` set, so they read low and do not sum; the real figure is the `total:` line from `go tool cover -func`, which is also the number used for the coverage badge.

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
