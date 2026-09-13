# GitHub Actions conventions

Language-agnostic CI conventions. Per-language paths filters, setup steps, and reusable workflow bodies live in the relevant language workflow fragment.

- Prefer official `actions/*` wherever one exists, and publish releases with the `gh` CLI rather than a release action. The third-party actions in the table below each do a job no `actions/*` covers. `goreleaser-action` is the one that looks like an exception and is not: it builds binaries and does not publish, so `gh release create` still does the publishing.
- Which workflow steps become `make` targets is settled in the Makefile fragment; version and changelog extraction always are.
- Minimal permissions: `contents: read` by default; `contents: write` only in release and prerelease workflows. Test workflows never declare `contents: write`. Every caller declares `contents: read` at the top and widens it on the jobs that need more, since a caller with no block inherits the repository default, which may be read and write.
- No `fetch-depth: 0` unless a step actually reads git history. Changelog extraction does not: `get_changelog` reads `CHANGELOG.md` out of the working tree, which a default depth-1 checkout has in full. The real cases are tools that inspect history or tags, such as goreleaser, tag-based versioning, and `git tag -f` against an existing tag. A release that only runs `gh release create` needs the default depth.
- `gh` infers the repository from the local git remote, so any job that calls it without an `actions/checkout` step must set `GH_REPO: ${{ github.repository }}`. Otherwise every call fails with `not a git repository`, which is not a not-found answer and must not be treated as one. Set it at job level alongside `GH_TOKEN` rather than per step.
- Never let a failed `gh` call stand in for a negative answer. An existence check has three outcomes, not two: it is there, it is not there, or the API could not say. Match the not-found message explicitly and fail the job on anything else. Both `|| true` and a bare `if gh view ...; then` collapse a rate limit, an auth failure or a flaky API into "it does not exist", and the step then does the wrong thing confidently.

Pin runners to a versioned label; never `-latest`, and never a Preview image. Linux jobs default to `ubuntu-24.04`. Take any other label from GitHub's current runner list rather than from this document, since images are added and retired on GitHub's schedule.

Architecture is part of the label, not a flag: on macOS the bare label is Apple Silicon and the `-intel` suffix is x64, so both architectures build natively and neither needs cross-compiling.

## What a job costs

A private repository is billed per runner-minute, and not every minute costs the same. GitHub applies a multiplier by platform:

| Platform | Multiplier |
|---|---|
| Linux | 1x |
| Windows | 2x |
| macOS | 10x |

A public repository is not billed at all today. Design as though it might be: the allowance is GitHub's to change, and a project that goes private later should not need its CI rethought to become affordable.

Two rules follow, and they apply to every workflow set here:

- **Do the work on Linux unless the platform is the point.** Two compilers on one Linux runner cost a fifth of one macOS job and find more. Reach for another platform when it is a deployment target, not for extra confidence in the same code.
- **Cheap by default, expensive by choice.** A new project starts with the smallest set that catches real breaks, and adds platforms and jobs deliberately. Anything that only a release needs is gated behind the trigger that releases, never run on every pull request. A default that costs nothing is one a project can afford to leave alone; a default that bills at 10x is one somebody has to notice.

Canonical action per purpose:

| Purpose | Action |
|---|---|
| Checkout | `actions/checkout` |
| Upload artefacts | `actions/upload-artifact` |
| Download artefacts | `actions/download-artifact` |
| Go setup | `actions/setup-go` |
| Python setup | `actions/setup-python` |
| uv setup | `astral-sh/setup-uv` |
| Ruff | `astral-sh/ruff-action` |
| GoReleaser (Go only) | `goreleaser/goreleaser-action` |
| Artifact signing | `sigstore/cosign-installer` |
| Install scripts and checksums | `thomaslaurenson/gpipe` |

Never use `@latest`. How tightly to pin below that depends on who publishes the action, because a tag is mutable: whoever owns the repository can move `@v7` to any commit at any time, and repointing a tag is a supply chain attack that has happened in the wild.

- **`actions/*`** pin to the current major (`@v7`). These are GitHub's own, on GitHub's infrastructure, and the mutable tag is an accepted risk in exchange for automatic patch and minor fixes.
- **Everything else** pins to a full commit SHA, with the version it corresponds to in a trailing comment. A SHA is the only immutable reference an action has.

```yaml
- uses: actions/checkout@vN
- uses: sigstore/cosign-installer@6f9f17788090df1f26f669e9d70d6ae9567deba6 # v4.1.2
```

The comment is not decoration: Dependabot reads it, bumps the SHA and rewrites the comment together, so a SHA pin costs no more to maintain than a tag. Without it the pin is an unreadable hex string that nobody dares touch.

An action published by the same person who owns the repository using it is not third-party in the sense that matters here, since compromising it and compromising the repository are the same event. Pin it to a major like `actions/*`.

`@vN` in the workflow fragments means the current major at authoring time, for example `@v7`. Do not treat any version number that has ever appeared in these documents as the target to match: a frozen version goes stale faster than the spec gets updated, and Dependabot (see below) keeps the pin current from there.

Use reusable workflows (`workflow_call`) for the job logic the callers share; callers compose them:

```text
.github/workflows/
  lint.yml        # reusable
  test.yml        # reusable
  release.yml     # reusable
  pr.yml          # caller: on pull requests
  tag.yml         # caller: on v* tags
  main.yml        # caller: on push to main
```

Three callers and at least three reusable workflows. Two more are conditional on what the project ships:

- **`build.yml`** exists only where a distributable artefact has to be produced before anything can consume it. A project whose tests build what they test does not need one, and adding it means a second build of the same sources.
- **`prerelease.yml`** exists only where there is an artefact to roll. A library ships no binary: its consumers take a git ref and compile it themselves, so there is nothing a rolling `dev` release could contain. Omit it, and `main.yml` is then lint and test alone, which still verifies the merged commit.

Both are decided by what the project ships, not by the language it is written in. "This language has a compile step" is the wrong test: a compiled library and a compiled application share a compiler and need different workflow sets.

- `pr.yml`: concurrency group `pr-${{ github.event.pull_request.number }}`, `cancel-in-progress: true`, with a `paths:` filter (language-specific). Grouping by pull request number means a push only ever cancels its own earlier run, and a superseded run of a branch nobody is looking at is pure waste.
- `main.yml`: concurrency group `main-${{ github.ref }}`, **`cancel-in-progress: false`**; `paths:` must match `pr.yml` exactly. Never cancel on main. A cancelled pull request run costs nothing but minutes, whereas a cancelled main run can stop midway through publishing, leaving a rolling release whose assets, tag and registry image disagree with each other. The two filters must match because a path that gates a pull request but not the merge lets main go red for a change no pull request ever ran on.
- `tag.yml`: no concurrency group and no `paths:` filter; every tag runs all jobs unconditionally. A release that skipped its tests because the tag happened to touch no matching path is worse than a slow one.
- No `push.yml`.

A workflow outside this set is allowed when it responds to something other than a change to the code: a scheduled scan, or a documentation site published from `docs/` on its own trigger. It follows every rule here, pinned runners and actions and least permissions, and it never wires into the three callers or gates a release. Where it has exactly one caller, body and caller share a file, since separating them would add a file and no reuse; a scheduled vulnerability scan is the shape.

The callers are the same shape in every language. Only the paths filter and the job list change, and the language workflow fragment supplies both.

```yaml
# pr.yml
name: PR

on:
  pull_request:
    paths: <language paths filter>

concurrency:
  group: pr-${{ github.event.pull_request.number }}
  cancel-in-progress: true

permissions:
  contents: read

jobs:
  lint:
    uses: ./.github/workflows/lint.yml
  test:
    uses: ./.github/workflows/test.yml
```

```yaml
# main.yml
name: Main

on:
  push:
    branches: [main]
    paths: <the same filter as pr.yml, character for character>

concurrency:
  group: main-${{ github.ref }}
  cancel-in-progress: false

permissions:
  contents: read

jobs:
  lint:
    uses: ./.github/workflows/lint.yml
  test:
    uses: ./.github/workflows/test.yml
  prerelease:
    needs: [lint, test]
    uses: ./.github/workflows/prerelease.yml
    permissions:
      contents: write
```

```yaml
# tag.yml
name: Tag

on:
  push:
    tags: ["v*"]

permissions:
  contents: read

jobs:
  lint:
    uses: ./.github/workflows/lint.yml
  test:
    uses: ./.github/workflows/test.yml
  release:
    needs: [lint, test]
    uses: ./.github/workflows/release.yml
    permissions:
      contents: write
      id-token: write
```

A reusable workflow takes no inputs and declares the least it needs:

```yaml
name: Lint

on:
  workflow_call:

jobs:
  lint:
    runs-on: ubuntu-24.04
    permissions:
      contents: read
    steps:
      - uses: actions/checkout@vN
      - <language setup step>
      - run: make check_all
```

The `permissions` block on a caller job is not decoration. A reusable workflow cannot grant itself more than its caller was given, so a `release.yml` asking for `id-token: write` fails at signing time unless `tag.yml` grants it too. That is why the write permissions appear in the caller and not only in the workflow that uses them.
