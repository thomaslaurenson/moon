# clang-format Conventions

Configuration and usage conventions for clang-format across all C++ projects.

## Design Principles

- clang-format is a formatter, not a style analyser. It enforces layout only; naming and code quality are clang-tidy's job.
- The project `.clang-format` is the single source of truth; never override it with command-line flags
- `BasedOnStyle: LLVM` is the baseline; only add explicit overrides for settings that genuinely differ from LLVM defaults
- A minimal override list is easier to maintain and easier to reason about than a full config dump

## Prerequisites

`clang-format-18` must be installed. The `make install_clang_tools` target handles this. Always use the pinned version, because different versions can produce different formatting output.

## Running clang-format

```bash
# Check for formatting violations (used in CI)
make fmt_check

# Auto-fix all formatting in place (used locally)
make fmt
```

`make fmt_check` exits non-zero if any file differs from the formatted output, causing CI to fail. `make fmt` rewrites files in place; run it before committing.

## Canonical `.clang-format`

```yaml
BasedOnStyle: LLVM
IndentWidth: 4
ColumnLimit: 100
AccessModifierOffset: -4
AllowShortFunctionsOnASingleLine: InlineOnly
IncludeBlocks: Regroup
IncludeCategories:
  - Regex:    '^<'
    Priority: 1
  - Regex:    '^"'
    Priority: 2
```

### What each override does

**`IndentWidth: 4`** -- LLVM default is `2`. Four spaces is more readable at the nesting depths typical in this codebase.

**`ColumnLimit: 100`** -- LLVM default is `80`. A hundred-column limit avoids wrapping long API call chains while still preventing runaway line lengths.

**`AccessModifierOffset: -4`** -- LLVM default is `-2`. Must equal `-IndentWidth` so that `public:` and `private:` labels sit flush with the enclosing class body, not indented relative to it.

**`AllowShortFunctionsOnASingleLine: InlineOnly`** -- LLVM default is `All`. Restricts single-line functions to trivial getters defined inside the class body. Standalone function definitions always get their own line.

**`IncludeBlocks: Regroup`** -- LLVM default is `Preserve`. Enforces the three-tier include ordering defined below.

### Settings that need no override

These are already the LLVM defaults and must not be added as overrides:

| Setting | LLVM default |
|---|---|
| `PointerAlignment` | `Right` -- `int *p`, not `int* p` |
| `BreakBeforeBraces` | `Attach` -- opening brace on same line |
| `IndentCaseLabels` | `false` -- `case:` labels at switch level, not indented |
| `Cpp11BracedListStyle` | `true` -- no spaces inside `{}` initialisers |
| `DerivePointerAlignment` | `false` -- never auto-detect from existing code |

## Include ordering

`IncludeBlocks: Regroup` sorts and groups includes into tiers. Configure `IncludeCategories` to match the project's include structure. The standard two-tier split:

```yaml
IncludeCategories:
  - Regex:    '^<'       # Tier 1: angle-bracket headers (stdlib + third-party)
    Priority: 1
  - Regex:    '^"'       # Tier 2: project headers
    Priority: 2
```

For projects with a large third-party dependency that needs its own tier, split tier 1:

```yaml
IncludeCategories:
  - Regex:    '^<(StormLib|CLI)/'   # Tier 2: named third-party (higher priority number = lower tier)
    Priority: 2
  - Regex:    '^<'                  # Tier 1: stdlib
    Priority: 1
  - Regex:    '^"'                  # Tier 3: project headers
    Priority: 3
```

Within each tier, includes are sorted alphabetically.

## Suppressing clang-format

See `cpp/style.md` (Suppressing clang-format) for when suppression is permitted and the required comment format.
