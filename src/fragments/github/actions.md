# GitHub Actions conventions

Language-agnostic CI conventions. Per-language paths filters, setup steps, and reusable workflow bodies live in the relevant language workflow fragment.

- Prefer official `actions/*` before third-party alternatives. Use the `gh` CLI for releases by default (`goreleaser-action` is the only exception, Go only).
- Create a Makefile target for a workflow step when it is useful locally, appears in more than one workflow, or contains non-trivial logic. Version and changelog extraction must always go through Makefile targets.
- Minimal permissions: `contents: read` by default; `contents: write` only in release and prerelease workflows. Test workflows never declare `contents: write`.
- No `fetch-depth: 0` unless a step actually reads git history. Changelog extraction does not: `get_changelog` reads `CHANGELOG.md` out of the working tree, which a default depth-1 checkout has in full. The real cases are tools that inspect history or tags, such as goreleaser, tag-based versioning, and `git tag -f` against an existing tag. A release that only runs `gh release create` needs the default depth.
- `gh` infers the repository from the local git remote, so any job that calls it without an `actions/checkout` step must set `GH_REPO: ${{ github.repository }}`. Otherwise every call fails with `not a git repository`, which is not a not-found answer and must not be treated as one. Set it at job level alongside `GH_TOKEN` rather than per step.
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
| Release pipelines | `thomaslaurenson/gpipe` |

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
