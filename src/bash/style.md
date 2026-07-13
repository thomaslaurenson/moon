# Bash style

Bash-specific style. Assumes the core conventions.

## Header

Executable scripts start with `#!/usr/bin/env bash` (portable; matters on macOS where `/bin/bash` is 3.2), then immediately:

```bash
set -euo pipefail
```

Do not set `set -euo pipefail` in sourced files; they run in the caller's shell.

## Formatting

- Indent with 2 spaces, never tabs. Max line length 100.
- Put `; then` and `; do` on the same line as `if`/`for`/`while`.

## Naming

- Functions and locals: `lowercase_with_underscores`; private functions take a leading underscore.
- Declare all function-local variables with `local`. Declare and assign separately when the value comes from a command substitution, so the exit code is not lost.
- Constants and exports: `UPPER_SNAKE_CASE`, declared `readonly`.

## Quoting and tests

- Always quote variables; prefer `"${var}"` over `"$var"`. Use `"$@"` to forward arguments, never `$*`.
- Prefer `[[ ... ]]` over `[ ... ]`. Use `-z`/`-n` explicitly. Use `(( ... ))` for arithmetic; never `expr` or `let`.
- Use `$(...)`, never backticks.

## Arrays and output

- Use arrays for lists; expand with `"${arr[@]}"`. Use process substitution or `readarray` instead of piping into `while` (a pipe subshell loses assignments).
- Use `printf`, not `echo`. All error messages go to stderr.

## Error handling and structure

- Define a `die` helper at the top of every executable script: `die() { printf '%s: %s\n' "${0##*/}" "$*" >&2; exit 1; }` and use it consistently.
- Any function that is not both obvious and short has a header comment; all functions in a sourced library have one regardless of length.
- For scripts with more than one function, wrap the entry point in `main` and call `main "$@"` at the bottom.
