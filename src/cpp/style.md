# C++ Style

C++-specific style. Assumes the core conventions. Formatting is enforced by clang-format and naming by clang-tidy (see the clang-format and clang-tidy fragments).

## Formatting

Enforced by clang-format; `make fmt` reformats all source. Deviations from LLVM defaults: `IndentWidth: 4`, `ColumnLimit: 100`, `AccessModifierOffset: -4`, `AllowShortFunctionsOnASingleLine: InlineOnly`, `IncludeBlocks: Regroup`. All other settings inherit from LLVM.

`// clang-format off`/`on` is permitted only where manual alignment aids readability (lookup tables); keep the block short and comment why. Never use it for personal preference.

## Naming

Enforced by clang-tidy's `readability-identifier-naming`.

- Functions and methods: `PascalCase`.
- Variables and parameters: `snake_case`.
- Member variables: `snake_case_` with a trailing underscore.
- Types, classes, structs, enums: `PascalCase`.
- Enum values: `UPPER_SNAKE_CASE`.
- Constants: `snake_case`, declared `constexpr` or `const`.
- No Hungarian notation (`hArchive`, `szName`); the type is already stated. Third-party APIs that use it are exempt.

## Files

- `snake_case` for all source and header file names.
- `.cpp` for implementation, `.h` for headers. Never `.hpp`/`.hxx`.
- Header-only `.h` for simple structs and helpers under ~50 lines; split into `.h`/`.cpp` when the implementation has real complexity.

## Project version

Declare the version once in the root `CMakeLists.txt` via `project(MyProject VERSION 1.2.3)`. Bake it into the target at configure time with `configure_file` and a `version.h.in`, so both a binary and a library's consumers can query it as a compile-time constant. Never hardcode a version string in a `.cpp`, and never read it from `git describe` at runtime.

## Comments

- Every function declared in a header has a Doxygen comment on its declaration; static/private helpers defined only in a `.cpp` have theirs in the `.cpp`. Never duplicate the comment in both.
- Use `///` triple-slash. First line is a single summary sentence with no full stop. Document every parameter with `@param`; include `@return` unless the function returns `void`.

```cpp
/// Opens an archive from the given path.
///
/// @param path Path to the archive file.
/// @param handle Pointer to the handle to populate.
/// @return true on success, false on failure.
bool OpenArchive(const std::string &path, HANDLE *handle);
```
