# Changelog format

- Source content from the PR description and recent commits; do not invent entries.
- One bullet per logical change; imperative voice ("Add support for X"), no trailing period, no inline code.
- Newest release at the top. Only include sections that have entries.

```markdown
# Changelog

## 1.2.3 - YYYY-MM-DD

### Added

- Short description of a new capability

### Fixed

- Short description of a bug fix
```

- File header is `# Changelog`. Version line `## X.Y.Z - YYYY-MM-DD` (ISO date, no `v` prefix).
- Section order: Added > Changed > Fixed > Removed > Updated > Thanks.
- Commit messages use past tense; changelog entries use imperative. They are different audiences.
