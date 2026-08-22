# Changelog format

The changelog tells someone who uses the project what changed between releases. It is not a record of the work done, and not a summary of the diff. A reader who wants the reasoning reads the pull request; a reader who wants the detail reads the log.

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

- File header is `# Changelog`. Version line `## X.Y.Z - YYYY-MM-DD` (ISO date, no `v` prefix). Git tags are `v`-prefixed (`v1.2.3`); the changelog header stays bare. The release workflow's `get_changelog` target strips the leading `v` from the tag before matching this header, so the tag convention and the changelog convention never need to share a prefix.
- Section order: Added > Changed > Fixed > Removed > Updated > Thanks.
- Never invent a section or qualify one. There is no "Changed (testing)"; those entries go under Changed.
- Commit messages use past tense; changelog entries use imperative. They are different audiences.

## Length

A release is a handful of bullets. Two is normal, eight is a lot, and twenty means the entries are being written per commit or per file instead of per change.

Each bullet is one line, roughly ten to fifteen words. A bullet needing a subordinate clause, a second sentence, or a colon and a list is carrying detail that belongs in the PR.

Length is the symptom, not the disease. A long changelog is usually a changelog written from the diff, walking the files that changed, rather than from the release, naming what a user can now do.

## Group related changes

One bullet names an area and the class of change in it. A set of similar fixes across many files is one entry, not one per file:

```markdown
# Good - one bullet, names the area
- Fix path handling on Windows across the config loader and the installer

# Bad - one bullet per file, restating the diff
- Fix the config loader's path separator
- Fix the installer's path separator
- Fix path quoting in the archive extractor
- Fix path quoting in the checksum reader
```

Two changes that a reader would think of as one thing share a bullet, joined by a comma:

```markdown
- Add JSON output to the list command, sort its results by name
```

## Leave out

- **Rationale.** Say what changed, not why the old way was wrong. Not "Resolve the config path against the executable rather than the working directory, which read the wrong file when the tool was run from a subdirectory", but "Fix config lookup from a subdirectory".
- **Mechanism.** The flag, option, or function that implements the change is detail. "Retry on a dropped connection", not "Set MaxRetries to 3 on the HTTP client".
- **File paths.** Name the area ("the config loader", "the release workflow"), not the file (`internal/config/load.go`). Name a symbol only when the symbol is the thing a user interacts with.
- **Anything a user cannot observe.** A refactor with no visible effect gets no entry. If nothing about the project's behaviour, interface, or output changed, the log already records it and the changelog does not need to.
