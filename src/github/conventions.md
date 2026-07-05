# Git Conventions

Branch names, commit messages, and PR titles.

## Commit messages

Direct commits use past tense, sentence case, capitalise the first word only, no trailing period. Multiple small changes may be comma-separated. No conventional-commits prefixes (`feat:`, `fix:`).

```
Fixed permission error in release workflow
Added dependabot and Python linting
```

## Branch names

Format `type/short-description`, all lowercase kebab-case, 3-5 words, no issue numbers.

| Type | When |
|---|---|
| `feature/` | New capabilities, subcommands, flags |
| `refactor/` | Internal restructuring, no behaviour change |
| `fix/` | Bug fixes |
| `update/` | Dependency bumps or toolchain upgrades |

## PR titles

Format `Type/short description (#N)`: capitalised type, lowercase description with spaces (not kebab-case), PR number appended.

```
Feature/add install script (#13)
Fix/local deployment (#44)
```

Dependabot commits are left as-is; do not reformat them.
