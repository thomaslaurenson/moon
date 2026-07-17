# Commit and PR conventions

Commit messages and pull request titles.

## Commit messages

Past tense, sentence case, capitalise the first word only, no trailing period. Multiple small changes may be comma-separated. No conventional-commits prefixes (`feat:`, `fix:`).

```
Fixed permission error in release workflow
Added dependabot and Python linting
```

Dependabot commits are left as-is; do not reformat them.

## PR titles

Format `Type/short description (#N)`: capitalised type, lowercase description with spaces (not kebab-case), PR number appended.

```
Feature/add install script (#13)
Fix/local deployment (#44)
```

The type is the same set used for branch names, capitalised:

| Type | When |
|---|---|
| `Feature/` | New capabilities, subcommands, flags |
| `Refactor/` | Internal restructuring, no behaviour change |
| `Fix/` | Bug fixes |
| `Update/` | Dependency bumps or toolchain upgrades |

A PR title is the branch name rewritten for a human: `feature/add-install-script` becomes `Feature/add install script (#13)`. Take the type and description from the branch rather than inventing new ones, so the two stay recognisably the same change.

Commit messages use past tense; changelog entries use imperative. They are different audiences; see the changelog fragment.
