# C++ Style Guide

Style conventions for C++ code in this project.

## Unusual Characters

- Never use em dash (—)

## Spelling

Use British English spellings:

- `Initialise` not `Initialize`
- `Colour` not `Color`

---

## Formatting

Formatting is enforced automatically by clang-format. Running `make format` will reformat all source files. The canonical configuration lives in `.clang-format` at the project root.

Key settings that diverge from stock LLVM style:

| Setting | Value |
|---|---|
| `BasedOnStyle` | `LLVM` |
| `IndentWidth` | `4` |
| `ColumnLimit` | `100` |
| `PointerAlignment` | `Right` |
| `BreakBeforeBraces` | `Attach` |

### Suppressing clang-format

`// clang-format off` and `// clang-format on` are permitted only when manual alignment genuinely aids readability; for example, a table or lookup structure where columnar layout communicates meaning that clang-format would destroy.

The suppressed block must be kept as short as possible, and a comment must explain why formatting is suppressed:

```cpp
// clang-format off -- columnar alignment shows command-to-handler mapping clearly
static const CommandEntry COMMANDS[] = {
    { "extract",  extract_command,  "Extract files from an archive" },
    { "add",      add_command,      "Add files to an archive"       },
    { "create",   create_command,   "Create a new archive"          },
};
// clang-format on
```

Never use `clang-format off` to preserve arbitrary personal formatting preferences.

---

## Naming Conventions

Naming is enforced by clang-tidy's `readability-identifier-naming` check.

### Functions

Use `snake_case`:

```cpp
// Good
bool open_archive(const std::string &path, HANDLE *handle);
std::string normalise_file_path(const std::string &path);

// Bad
bool OpenArchive(const std::string &path, HANDLE *handle);
```

### Variables and Parameters

Use `snake_case`:

```cpp
// Good
int file_count = 0;
std::string archive_path = get_path();

// Bad
int fileCount = 0;
std::string archivePath = get_path();
```

### Member Variables

Use `snake_case` with a trailing underscore. The trailing underscore distinguishes member variables from local variables and parameters, and avoids shadowing in constructors:

```cpp
class AppConfig {
public:
    AppConfig(OperationMode mode) : mode_(mode) {}

private:
    OperationMode mode_;
    std::vector<ProcessingRule> rules_;
};
```

### Types, Classes, Structs, and Enums

Use `PascalCase`:

```cpp
class AppConfig { ... };
struct ProcessingRule { ... };
enum class OperationMode { ... };
```

### Enum Values

Use `UPPER_SNAKE_CASE`:

```cpp
enum class OperationMode {
    DEFAULT,
    FAST,
    VERBOSE,
};
```

### Constants

Use `UPPER_SNAKE_CASE`. Declare with `constexpr` or `const` as appropriate:

```cpp
constexpr int MAX_FILE_COUNT = 1024;
const std::string DEFAULT_LOCALE = "enUS";
```

### No Hungarian Notation

Do not use Hungarian notation prefixes. The type is already stated in the declaration; prefixes add noise without value:

```cpp
// Good
HANDLE archive;
std::string file_name;
DWORD flags;

// Bad
HANDLE hArchive;
std::string szFileName;
DWORD dwFlags;
```

Third-party libraries that use Hungarian notation in their own APIs are explicitly exempt; do not rename parameters or members that originate from external library types.

### File Names

Use `snake_case` for all source and header files:

```
app_config.cpp
app_config.h
file_helpers.cpp
file_helpers.h
```

### File Extensions

Use `.cpp` for implementation files and `.h` for header files. Never use `.hpp`, `.hxx`, or any other variant; the project uses `.h` throughout and mixing extensions creates unnecessary inconsistency.

### Header-only vs Split Files

Whether a file is header-only or split into `.h`/`.cpp` depends on the complexity of the implementation:

- **Header-only `.h`**: use for simple structs, fixtures, and small helpers where the entire definition is 50 lines or fewer. Putting a 30-line struct across two files is unnecessary ceremony.
- **`.h`/`.cpp` split**: use when the implementation has real complexity, multiple private functions, or would meaningfully slow down compilation if included everywhere. The declaration in `.h` is the public contract; the implementation in `.cpp` is the detail.

---

## Project Version

Every C++ project must declare its version in the root `CMakeLists.txt` using the `project()` command's `VERSION` parameter. This is the single source of truth for the project version; never hardcode a version string anywhere else:

```cmake
project(MyApp VERSION 1.2.3)
```

The version is then available in CMake as `${PROJECT_VERSION}` and can be baked into the binary at configure time using `configure_file`:

```cmake
# In root CMakeLists.txt
configure_file(src/version.h.in src/version.h)
```

```cpp
// src/version.h.in
#pragma once
#define MYAPP_VERSION "@PROJECT_VERSION@"
```

```cpp
// Usage in source
#include "version.h"
std::cout << "myapp " << MYAPP_VERSION << "\n";
```

Never hardcode a version string in a `.cpp` file. Never read the version from `git describe` at runtime; bake it at configure time.

---

## Comments

### Function Comments

Every function declared in a header file must have a Doxygen comment on its declaration. Functions defined only in a `.cpp` file (static helpers, private implementation) must have their Doxygen comment in the `.cpp` file.

Never duplicate a comment in both the `.h` declaration and the `.cpp` definition. The declaration is the contract; that is where the comment lives.

Use triple-slash `///` style. The first line is a single summary sentence with no full stop. An optional extended description may follow after a blank `///` line. Every parameter is documented with `@param`. `@return` is always present unless the function returns `void`:

```cpp
/// Opens an archive from the given path.
///
/// The handle must be closed with close_archive when no longer needed.
///
/// @param path Path to the archive file.
/// @param handle Pointer to the handle to populate.
/// @param flags Open flags.
/// @return true on success, false on failure.
bool open_archive(const std::string &path, HANDLE *handle, uint32_t flags);
```

For `void` functions, omit `@return`:

```cpp
/// Normalises a file path to use backslashes and lowercase.
///
/// @param path Path string to normalise in place.
void normalise_file_path(std::string &path);
```

### Inline Comments

- Start with `//` followed by a single space.
- First word is capitalised.
- Never use a full stop, unless the comment is multiple sentences.
- No decorative dividers; avoid `// ---`, `// ===`, `// ***`, or similar.

```cpp
// Good: single-line comment
int next = next_power_of_two(count);

// Good: multi-line comment where only the first line is capitalised,
// continuation lines do not need to start with a capital letter.
std::transform(path.begin(), path.end(), path.begin(),
               [](unsigned char c) { return std::tolower(c); });

// Bad: decorative divider
// --- file helpers ---
```

### Comment Hygiene

- Do not write step narration comments that describe the next line of code. Bad: `// Loop through files`, `// Check if handle is valid`
- Preserve comments that explain why something is done, not what. Good: `// Library expects backslashes regardless of platform`
- Do not inject `TODO` or `FIXME` comments unless they refer to a real, known issue.

---

## Static Analysis

Static analysis is enforced by clang-tidy. Running `make lint` will run the full check suite. The canonical configuration lives in `.clang-tidy` at the project root.

### Baseline checks

Every project enables this baseline and nothing more by default:

```yaml
Checks: >
  clang-analyzer-core.*,
  clang-analyzer-cplusplus.*,
  clang-analyzer-deadcode.*,
  modernize-use-nullptr,
  readability-identifier-naming
WarningsAsErrors: "*"
FormatStyle: file
```

- `clang-analyzer-*`: deep static analysis for real bugs; null dereferences, memory leaks, dead code, and undefined behaviour
- `modernize-use-nullptr`: enforces `nullptr` over `NULL` throughout
- `readability-identifier-naming`: enforces the naming conventions defined above
- `WarningsAsErrors`: all warnings are errors; if a check is worth having it is worth fixing

### Project-specific checks and suppressions

Additional checks and suppressions are added per project in the project's own `.clang-tidy` file. Every suppression must include a comment explaining why it is justified:

```yaml
Checks: >
  -bugprone-narrowing-conversions
# Suppressed: third-party library uses DWORD/int32_t interchangeably at its API boundary.
# Fixing these would require wrapping every library call with explicit casts
# that add noise without improving safety.
```

Never suppress a check without a documented reason.
