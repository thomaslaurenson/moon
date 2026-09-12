# Branches and commits

How a change reaches `main`: on a branch, through a pull request, recorded by commits that read as plain English.

## Branches

- One branch per piece of work, cut from `main` and merged back by pull request.
- The name is `type/short-description`: lowercase, words separated by hyphens.
- The pull request title is the branch name with the type capitalised and the hyphens turned to spaces, so the two never need reconciling; see the pull request fragment.

The type is one of:

| Type | When |
|---|---|
| `feature` | New capabilities, subcommands, flags |
| `refactor` | Internal restructuring, no behaviour change |
| `fix` | Bug fixes |
| `update` | Dependency bumps or toolchain upgrades |

```text
feature/review-bundles
fix/golang-spec-correctness
refactor/cpp-cleanup
```

## Commit messages

One line, and nothing else. Past tense, sentence case, capitalise the first word only, no trailing period, no conventional-commits prefix (`feat:`, `fix:`).

- Say what changed, in plain English, as briefly as it can be said. Multiple small changes may be comma-separated.
- No body. The reasoning belongs in the pull request and the detail is in the diff.
- No trailers of any kind: no `Co-Authored-By`, no `Signed-off-by`, no generated-by line. The author field is the whole attribution.
- Dependabot commits are left as-is; do not reformat them.

```text
Fixed permission error in release workflow
Added dependabot and Python linting
Moved output markers out of core
```

Commit messages use past tense. Changelog entries, where the project keeps one, use the imperative; they are written for different readers.
