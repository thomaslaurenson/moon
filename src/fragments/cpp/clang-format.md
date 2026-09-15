# clang-format conventions

Configuration and usage conventions for clang-format across all C++ projects.

## Design principles

- clang-format is a formatter, not a style analyser. It enforces layout only; naming and code quality are clang-tidy's job.
- The project `.clang-format` is the single source of truth; never override it with command-line flags
- `BasedOnStyle: LLVM` is the baseline; only add explicit overrides for settings that genuinely differ from LLVM defaults
- A minimal override list is easier to maintain and easier to reason about than a full config dump

## Prerequisites

clang-format at the major version the CMake fragment pins under Clang tooling, which is also where the reasons for pinning and the per-platform install differences live. The Makefile targets fragment resolves whichever binary is present into `$(CLANG_FORMAT)`. A different major produces different output and fails `check_format` on lines nobody edited.

## Running clang-format

```bash
# Check for formatting violations (used in CI)
make check_format

# Auto-fix all formatting in place (used locally)
make format
```

`make check_format` exits non-zero if any file differs from the formatted output, causing CI to fail. `make format` rewrites files in place; run it before committing.

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

**`IndentWidth: 4`**. The LLVM default is `2`. Four spaces is more readable at the nesting depths typical in these projects.

**`ColumnLimit: 100`**. The LLVM default is `80`. A hundred-column limit avoids wrapping long API call chains while still preventing runaway line lengths.

**`AccessModifierOffset: -4`**. The LLVM default is `-2`. Must equal `-IndentWidth` so that `public:` and `private:` labels sit flush with the enclosing class body, not indented relative to it.

**`AllowShortFunctionsOnASingleLine: InlineOnly`**. The LLVM default is `All`. Restricts single-line functions to trivial getters defined inside the class body. Standalone function definitions always get their own line.

**`IncludeBlocks: Regroup`**. The LLVM default is `Preserve`. Enforces the three-tier include ordering defined below.

### Settings that need no override

These are already the LLVM defaults and must not be added as overrides:

| Setting | LLVM default |
|---|---|
| `PointerAlignment` | `Right`: `int *p`, not `int* p` |
| `BreakBeforeBraces` | `Attach`: opening brace on same line |
| `IndentCaseLabels` | `false`: `case:` labels at switch level, not indented |
| `Cpp11BracedListStyle` | `true`: no spaces inside `{}` initialisers |
| `DerivePointerAlignment` | `false`: never auto-detect from existing code |

`IncludeIsMainRegex` is also left at its default. The `([-_]test)?$` variant found in older projects matches a `foo_test.cpp` suffix, and tests here are named `test_foo.cpp`, so it never fires; a setting that does nothing is one a reader still has to understand.

## Include ordering

`IncludeBlocks: Regroup` sorts and groups includes into tiers. Configure `IncludeCategories` to match the project's include structure. The standard two-tier split:

```yaml
IncludeCategories:
  - Regex:    '^<'       # Tier 1: angle-bracket headers (stdlib + third-party)
    Priority: 1
  - Regex:    '^"'       # Tier 2: project headers
    Priority: 2
```

A tier with a public API needs a third tier, because its own headers are angle-included. The library scaffolding fragment puts every public header under `include/myproj/` and has consumers write `#include <myproj/archive/archive.h>`, and the project's own sources do the same. Under the two-tier split above those match `^<` like any standard header and sort alphabetically among them, so `<myproj/archive.h>` lands between `<optional>` and `<span>`. Give them a tier of their own:

```yaml
IncludeCategories:
  - Regex:    '^<myproj/'   # Tier 2: this project's public headers
    Priority: 2
  - Regex:    '^<'          # Tier 1: stdlib and third-party
    Priority: 1
  - Regex:    '^"'          # Tier 3: private headers under src/
    Priority: 3
```

Substitute the real project name in the regex, as everywhere else `myproj` appears. An application has no `include/myproj/` and needs no such tier; the two-tier split is the whole of its configuration.

For projects with a large third-party dependency that needs its own tier, split tier 1:

```yaml
IncludeCategories:
  - Regex:    '^<(catch2|CLI)/'     # Tier 2: named third-party (higher priority number = lower tier)
    Priority: 2
  - Regex:    '^<'                  # Tier 1: stdlib
    Priority: 1
  - Regex:    '^"'                  # Tier 3: project headers
    Priority: 3
```

Within each tier, includes are sorted alphabetically.

## Suppressing clang-format

See the C++ style fragment (Suppressing clang-format) for when suppression is permitted and the required comment format.
