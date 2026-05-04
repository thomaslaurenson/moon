# Commit Message Conventions

## Contents

- [Design principles](#design-principles) — sentence case, past tense, no conventional-commits prefix
- [Commit messages](#commit-messages) — format, examples, multi-change lines
- [Branch names](#branch-names) — type prefix, kebab-case, type table

## Design Principles

- Commit messages are short, sentence-case, past tense, no trailing period
- Branches use a type prefix and kebab-case description
- Dependabot commits are left as-is, do not reformat them

## Commit Messages

Direct commits (not PR merges) are short past-tense phrases:

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
- Description should be short but unambiguous (3-5 words)
- No issue numbers in branch names

| Type | When to use |
|---|---|
| `feature/` | New capabilities, subcommands, flags |
| `refactor/` | Internal restructuring with no behaviour change |
| `fix/` | Bug fixes, crash fixes, error corrections |
| `update/` | Dependency bumps or toolchain upgrades |
