# Git Conventions

Conventions for branch names, commit messages, and PR titles across all projects.

## Design Principles

- Branches use a type prefix and kebab-case description
- Commit messages are short, sentence-case, past tense, no trailing period
- PR titles mirror the branch name with a capitalised type and the PR number appended
- Dependabot commits are left as-is, do not reformat them

## Commit Messages

Direct commits (not PR merges) are short imperative or past-tense phrases:

```
Fixed permission error in release workflow
Bumped golang, tidied sshfs function, fixed bash history bug
Added dependabot and Python linting
Removed unused tag and release script
```

- Sentence case, capitalise the first word only
- No trailing period
- Multiple small changes on one line can be comma-separated
- No conventional-commits prefixes (`feat:`, `fix:`, etc.)

## Branch Names

Format: `type/short-description`

```
feature/add-install-script
refactor/shell-design
fix/local-deployment
update/mpqeditor-and-uv
```

- All lowercase, kebab-case
- Description should be short but unambiguous (3–5 words)
- No issue numbers in branch names

| Type | When to use |
|---|---|
| `feature/` | New capabilities, subcommands, flags |
| `refactor/` | Internal restructuring with no behaviour change |
| `fix/` | Bug fixes, crash fixes, error corrections |
| `update/` | Dependency bumps or toolchain upgrades |

## PR Titles

Format: `Type/short description (#N)`

```
Feature/add install script (#13)
Refactor/shell design (#18)
Fix/local deployment (#44)
Update/mpqeditor and uv (#40)
```

- Type is capitalised (`Feature`, `Refactor`, `Fix`, `Update`)
- Description is lowercase with spaces, not kebab-case
- PR number is appended in parentheses
- Matches the branch name closely but uses spaces instead of hyphens
