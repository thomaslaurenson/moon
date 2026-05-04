# Pull Request Conventions

## Contents

- [Design principles](#design-principles) — PR titles mirror branch names, number appended
- [PR titles](#pr-titles) — format, examples, capitalisation rules

## Design Principles

- PR titles mirror the branch name with a capitalised type and the PR number appended

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
