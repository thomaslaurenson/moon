# Copilot Instructions

## Always Apply

These rules apply to every file, every language, every task without exception.

- No em dash (`—`). Rewrite using a comma, parentheses, or two sentences.
- No smart or curly quotes (`"` `"` `'` `'`). Always use straight ASCII quotes (`"` `'`).
- No non-ASCII characters in code or comments.
- No decorative dividers in code (`// ---`, `// ===`, `# ---`, `# ===`, etc.). Delete them.
- No step narration comments. Never write a comment that just describes the next line of code.
  Bad: `// Open the file`, `# Loop through results`
- Preserve comments that explain the 'why' (business logic, architecture). Aggressively delete and refactor comments that narrate the 'what' (step-by-step code narration).
- Use British English. `initialise` not `initialize`. `colour` not `color`.

## Navigation

Use this table to find the relevant instruction file for your task.

| Task | File |
|---|---|
| Writing or reviewing Go code | `golang/style.md` |
| Scaffolding a new Go project | `golang/scaffolding.md` |
| GoReleaser config | `golang/goreleaser.md` |
| Writing or reviewing Python code | `python/style.md` |
| Scaffolding a new Python project | `python/scaffolding.md` |
| Writing a Python library | `python/library.md` |
| Sphinx documentation | `python/docs.md` |
| Writing or reviewing Python tests | `python/testing.md` |
| C++ code (conventions not yet defined) | `cpp/style.md` |
| GitHub Actions workflows | `github/actions.md` |
| README badges | `github/badges.md` |
| CHANGELOG entries | `github/changelog.md` |
| Dependabot config | `github/dependabot.md` |
| Commit messages, branch names, and PR titles | `github/conventions.md` |
| Writing a Makefile | `tools/makefile.md` |
