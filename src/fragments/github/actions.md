# GitHub Actions conventions

Language-agnostic CI conventions. Per-language paths filters, setup steps, and reusable workflow bodies live in the relevant language workflow fragment.

- Prefer official `actions/*` before third-party alternatives. Use the `gh` CLI for releases by default (`goreleaser-action` is the only exception, Go only).
- Create a Makefile target for a workflow step when it is useful locally, appears in more than one workflow, or contains non-trivial logic. Version and changelog extraction must always go through Makefile targets.
- Minimal permissions: `contents: read` by default; `contents: write` only in release and prerelease workflows. Test workflows never declare `contents: write`.
- No `fetch-depth: 0` except in release workflows where changelog extraction requires it.
- Never let a failed `gh` call stand in for a negative answer. An existence check has three outcomes, not two: it is there, it is not there, or the API could not say. Match the not-found message explicitly and fail the job on anything else. Both `|| true` and a bare `if gh view ...; then` collapse a rate limit, an auth failure or a flaky API into "it does not exist", and the step then does the wrong thing confidently.

Pin runners; never use `-latest`. Supported: `ubuntu-24.04`, `ubuntu-24.04-arm`, `macos-14`, `macos-15`, `windows-2022`, `windows-2025`.

Canonical action per purpose:

| Purpose | Action |
|---|---|
| Checkout | `actions/checkout` |
| Upload artefacts | `actions/upload-artifact` |
| Go setup | `actions/setup-go` |
| Python setup | `actions/setup-python` |
| uv setup | `astral-sh/setup-uv` |
| Ruff | `astral-sh/ruff-action` |
| GoReleaser (Go only) | `goreleaser/goreleaser-action` |
| Artifact signing | `sigstore/cosign-installer` |
| Release pipelines | `thomaslaurenson/gpipe-action` |

Pin every action to a specific version, never `@latest`; use whichever version is current at the time of authoring. Do not treat any version number that has ever appeared in this doc as the target to match - a frozen version table goes stale faster than this spec gets updated. Dependabot (see below) keeps the pin current from there.

Use reusable workflows (`workflow_call`) for all job logic; callers compose them:

```
.github/workflows/
  lint.yml        # reusable
  test.yml        # reusable
  release.yml     # reusable
  prerelease.yml  # reusable
  pr.yml          # caller: lint + test on PRs
  tag.yml         # caller: lint + test + release on v* tags
  main.yml        # caller: lint + test + prerelease on push to main
```

- `pr.yml`: concurrency group `pr-${{ github.event.pull_request.number }}`, `cancel-in-progress: true`, with a `paths:` filter (language-specific).
- `main.yml`: concurrency group `main-${{ github.ref }}`, `cancel-in-progress: false`; `paths:` must match `pr.yml` exactly.
- `tag.yml`: no concurrency group and no `paths:` filter; every tag runs all jobs unconditionally.
- No `push.yml`. Languages with a compile step before lint/test add a `build.yml` reusable workflow and `needs: build` in callers.
