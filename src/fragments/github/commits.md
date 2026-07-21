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

The type is one of:

| Type | When |
|---|---|
| `Feature/` | New capabilities, subcommands, flags |
| `Refactor/` | Internal restructuring, no behaviour change |
| `Fix/` | Bug fixes |
| `Update/` | Dependency bumps or toolchain upgrades |

Commit messages use past tense; changelog entries use imperative. They are different audiences; see the changelog fragment.
