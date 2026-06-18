# Bash Style Guide

Style conventions for Bash scripts and sourced files.

## Unusual Characters

- Never use em dash (—)

## Shebang

Executable scripts must start with:

```bash
#!/usr/bin/env bash
```

Prefer `#!/usr/bin/env bash` over `#!/bin/bash` for portability; it resolves bash from the environment rather than assuming a fixed path, which is important on macOS where `/bin/bash` is 3.2.

Sourced files (`.sh` libraries) do not require a shebang line.

## Shell Options

Every executable script must set these options immediately after the shebang:

```bash
#!/usr/bin/env bash

set -euo pipefail
```

- `-e`: exit immediately if a command exits with a non-zero status
- `-u`: treat unset variables as an error
- `-o pipefail`: the return value of a pipeline is the status of the last command to exit with a non-zero status

Do not set `set -euo pipefail` in sourced files. Sourced files run in the caller's shell environment and applying `set -e` can cause unexpected exits in the parent shell.

## Spelling

Use British English spellings:

- `Initialise` not `Initialize`
- `Colour` not `Color`

## Formatting

### Indentation

Indent with 2 spaces. Never use tabs.

### Line Length

Maximum line length is 100 characters. For long strings or commands, use line continuation with `\` or a heredoc.

```bash
# Good: split long command with continuation
cmake -B build \
  -DCMAKE_BUILD_TYPE=Release \
  -DBUILD_MYAPP=ON

# Good: heredoc for long strings
cat <<'EOF'
This is a long string that would exceed
the line length limit if written inline.
EOF
```

### Pipelines

If a pipeline fits on one line, keep it on one line. If not, split at each pipe with the `|` at the start of the continuation line, indented 2 spaces:

```bash
# Good: fits on one line
find . -name "*.sh" | sort

# Good: split across lines
find . -name "*.txt" \( -type f -o -type l \) \
  | while IFS= read -r f; do printf '%s\n' "${f##*/}"; done \
  | sort
```

### Control Flow

Put `; then` and `; do` on the same line as `if`, `for`, and `while`. `else` and `fi`/`done` go on their own lines:

```bash
for entry in "${entries[@]}"; do
  if [[ -f "$entry" ]]; then
    process "$entry"
  else
    die "entry not found: $entry"
  fi
done
```

### Case Statements

Indent alternatives by 2 spaces. Place `;;` on its own line for multi-line alternatives. One-line alternatives may keep `;;` on the same line:

```bash
case "$cmd" in
  help|-h|--help) show_help ;;
  version)        print_version ;;
  run)
    do_run "$@"
    ;;
  *)
    die "unknown command: $cmd"
    ;;
esac
```

## Naming Conventions

### Functions

Use `lowercase_with_underscores`. Private functions (not part of the public interface) use a leading underscore:

```bash
# Public function
process_item() { ... }

# Private function
_select_item() { ... }
```

### Variables

Local variables and parameters use `lowercase_with_underscores`. Declare all function-local variables with `local`:

```bash
process_file() {
  local file_path="$1"
  local line_count
  line_count="$(wc -l < "$file_path")"
  ...
}
```

Declare and assign separately when the value comes from a command substitution. `local` does not propagate the exit code of the command substitution, so combining them silently swallows errors:

```bash
# Bad: exit code of wc is lost
local line_count="$(wc -l < "$file")"

# Good: exit code is preserved
local line_count
line_count="$(wc -l < "$file")"
```

### Constants and Environment Variables

Constants and exported variables use `UPPER_SNAKE_CASE`. Declare with `readonly` for constants:

```bash
readonly VERSION="1.0.0"
readonly DEFAULT_DATA_DIR="${HOME}/.myapp"
```

### File Names

Executable scripts use no extension or a `.sh` extension. Sourced library files use a `.sh` extension and must not be executable.

## Variables and Quoting

Always quote variables to prevent word splitting and glob expansion. Prefer `"${var}"` over `"$var"` for all variables except single-character shell specials:

```bash
# Good
echo "${my_var}"
cp "${source_file}" "${dest_dir}/"

# Good: single-character specials do not need braces
echo "$1" "$?" "$#"

# Bad: unquoted variable
cp $source_file $dest_dir/
```

Use `"$@"` to forward arguments. Never use `$*` unless you specifically need all arguments joined as a single string.

## Tests and Conditionals

Prefer `[[ ... ]]` over `[ ... ]` and `test`. The double-bracket form prevents word splitting and supports pattern matching:

```bash
# Good
if [[ -f "${file}" ]]; then ...
if [[ "${var}" == "expected" ]]; then ...
if [[ "${var}" =~ ^[A-Za-z_][A-Za-z0-9_]*$ ]]; then ...

# Bad
if [ -f "$file" ]; then ...
```

Use `-z` and `-n` explicitly to test for empty or non-empty strings:

```bash
# Good
if [[ -z "${var}" ]]; then ...
if [[ -n "${var}" ]]; then ...

# Bad: implicit truthiness test
if [[ "${var}" ]]; then ...
```

Use `(( ... ))` for arithmetic comparisons. Never use `$[ ... ]`, `expr`, or `let`:

```bash
# Good
if (( count > 0 )); then ...
(( total += 1 ))

# Bad
if [ "$count" -gt 0 ]; then ...
total=$(expr "$total" + 1)
```

## Command Substitution

Always use `$(...)` over backticks. Backticks require escaping when nested and are harder to read:

```bash
# Good
result="$(some_command)"
nested="$(outer "$(inner)")"

# Bad
result="`some_command`"
```

## Arrays

Use arrays to store lists of elements. Always expand with `"${arr[@]}"` to preserve elements with spaces:

```bash
# Good
local args=(--multi --height=40% --layout=reverse)
fzf "${args[@]}"

# Bad: string with multiple arguments
local args="--multi --height=40% --layout=reverse"
fzf $args
```

Use process substitution or `readarray` instead of piping into `while`, as pipes create a subshell and variable assignments are lost:

```bash
# Good: process substitution preserves variable scope
while IFS= read -r line; do
  entries+=("$line")
done < <(find . -name "*.txt")

# Bad: assignments inside the while loop are lost
find . -name "*.txt" | while IFS= read -r line; do
  entries+=("$line")
done
```

## Output

Use `printf` instead of `echo` for all output. `echo` behaviour varies across shells and platforms, particularly with escape sequences and flags:

```bash
# Good
printf 'status: %s\n' "$message"
printf '%s=%q\n' "$key" "$value"

# Bad
echo "status: $message"
```

All error messages go to stderr:

```bash
printf 'error: %s\n' "$message" >&2
```

## Error Handling

Define an error-exit helper at the top of every executable script to print a
message to stderr and exit with status 1. Name it `die` or `error`:

```bash
die()   { printf '%s: %s\n' "${0##*/}" "$*" >&2; exit 1; }
error() { printf '%s: %s\n' "${0##*/}" "$*" >&2; exit 1; }
```

Use the helper consistently rather than inline `echo ... >&2; exit 1` patterns:

```bash
# Good
[[ -f "${config}" ]] || die "config file not found: ${config}"

# Bad
if [[ ! -f "${config}" ]]; then
  echo "config file not found: ${config}" >&2
  exit 1
fi
```

## Function Comments

Any function that is not both obvious and short must have a header comment. All functions in a sourced library must have a header comment regardless of length.

The comment describes the function's behaviour, followed by labelled sections for arguments, environment variables read, outputs, and return values. Omit sections that do not apply:

```bash
# Resolve an item path, falling back to an interactive selector when not found directly.
#
# Arguments:
#   $1 - Candidate item path (optional; triggers selector if empty or not found)
# Environment:
#   DATA_DIR - root data directory (default: ~/.myapp)
# Outputs:
#   stdout: resolved item path(s), one per line
#   stderr: error message if the candidate is invalid
# Returns:
#   0 on success
#   exits 1 if the candidate is invalid or no item can be resolved
_resolve_item() {
  ...
}
```

## Inline Comments

- Start with `#` followed by a single space.
- First word is capitalised.
- Never use a full stop, unless the comment is multiple sentences.
- No decorative dividers; avoid `# ---`, `# ===`, `# ***`, or similar.

```bash
# Good: single-line comment
local data_dir="${DATA_DIR:-${HOME}/.myapp}"

# Good: multi-line comment where only the first line is capitalised,
# continuation lines do not need to start with a capital letter.
find "$data_dir" -name "*.txt" \( -type f -o -type l \)

# Bad: decorative divider
# --- item resolution ---
```

## Comment Hygiene

- Do not write step narration comments that describe the next line of code. Bad: `# Loop through entries`, `# Check if file exists`
- Preserve comments that explain why something is done, not what. Good: `# Strip trailing CR - handles CRLF files transparently`
- Do not inject `TODO` or `FIXME` comments unless they refer to a real, known issue.

## main Function

For scripts long enough to contain more than one function, wrap the entry point in a `main` function and call it at the bottom of the file. This keeps all executable code inside functions and allows all variables to be declared `local`:

```bash
# Entry point: parse the subcommand and dispatch to the appropriate handler.
#
# Arguments:
#   $1 - Subcommand name (default: help)
#   $@ - Arguments forwarded to the subcommand handler
main() {
  local cmd="${1:-help}"
  shift || true
  case "$cmd" in
    run) do_run "$@" ;;
    *)   die "unknown command: $cmd" ;;
  esac
}

main "$@"
```

For short linear scripts with no functions, `main` is not required.
