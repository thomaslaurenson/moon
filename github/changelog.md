# Changelog Format

## Design Principles

- Source content from the PR description and recent commits - do not invent entries
- One bullet per logical change - group related fixes into a single line
- Imperative voice: "Add support for X", not "Added support for X" or "Adds support for X"
- No trailing period on bullets
- No inline code in bullets - describe the change in plain English
- Newest release at the top
- Only include sections that have entries - omit empty sections entirely
- Omit patch-level detail (file names, line numbers, function names)

## Format

```markdown
# Changelog

## v1.2.3 - YYYY-MM-DD

### Added

- Short description of new feature or capability

### Changed

- Short description of a behaviour or interface change

### Fixed

- Short description of a bug fix

### Thanks

- Thanks to @contributor for the contribution
```

- File header is `# Changelog`, not `# CHANGELOG`
- Version line: `## vX.Y.Z - YYYY-MM-DD`
- Date is ISO 8601 (`YYYY-MM-DD`)
- Version prefix `v` is optional but must be consistent within a repo

## Sections

| Section | When to use |
|---|---|
| `### Added` | New features, subcommands, flags, or capabilities |
| `### Changed` | Behaviour changes, interface changes, refactors visible to users |
| `### Fixed` | Bug fixes, error corrections, crash fixes |
| `### Removed` | Features or behaviours that have been removed |
| `### Updated` | Dependency bumps or toolchain upgrades (use sparingly) |
| `### Thanks` | Credit external contributors by GitHub handle |

Section order: **Added → Changed → Fixed → Removed → Updated → Thanks**

## Example

```markdown
## v0.5.1 - 2026-04-15

### Added

- Shell prompt integration for oh-my-zsh and fish

### Fixed

- Shell prompt integration for bash and zsh
- Running the shell command inside VS Code no longer crashes the terminal
- Starting a shell session inside an existing session now exits cleanly
- Terminal state is now restored exactly once on exit
```
