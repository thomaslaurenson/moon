# C++ style

C++-specific style. Assumes the core conventions. Where the project includes the clang tooling, formatting is enforced by clang-format and naming by clang-tidy (see the clang-format and clang-tidy fragments); the rules below define the intended style and apply whether or not that tooling is wired up.

## Formatting

Enforced by clang-format; `make fmt` reformats all source. Deviations from LLVM defaults: `IndentWidth: 4`, `ColumnLimit: 100`, `AccessModifierOffset: -4`, `AllowShortFunctionsOnASingleLine: InlineOnly`, `IncludeBlocks: Regroup`. All other settings inherit from LLVM.

## Suppressing clang-format

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

## Namespaces

Everything a project compiles into a library goes in a namespace named after the library, in `snake_case`:

```cpp
namespace mylib {
// ...
}  // namespace mylib
```

A library that leaves `OpenArchive()` at global scope is broken for its consumers: the name collides with any other dependency that had the same idea, and the collision surfaces at link time in someone else's build. The namespace is not decoration, it is what makes the library linkable alongside code you have never seen.

- Close every namespace with a `}  // namespace mylib` comment; the opening brace is often hundreds of lines away.
- Never use `using namespace` at file scope in a header. It forces the import on every consumer that includes it. Inside a `.cpp`, or inside a function, it is fine.
- Nest sparingly. One level for the library, plus `detail` for implementation types that must be in a header but are not API. Deep nesting reads as directory structure leaking into code.
- Prefer an anonymous namespace over `static` for file-local helpers in a `.cpp`; it applies to types as well as functions.

An application binary's own translation units (`app/`) need no namespace: nothing links against them.

## Project version

Declare the version once in the root `CMakeLists.txt` via `project(MyProject VERSION 1.2.3)`. Bake it into the target at configure time with `configure_file` and a `version.h.in`, so both a binary and a library's consumers can query it as a compile-time constant. Never hardcode a version string in a `.cpp`, and never read it from `git describe` at runtime.

## Comments

The core conventions govern implementation comments; this narrows them for headers.

- Use `///` triple-slash for anything documenting a declaration, never `//` or `/** */`. The first line is a single summary sentence with no full stop.
- A comment on a declaration lives on the declaration, never duplicated onto the definition in the `.cpp`.

```cpp
/// Opens an archive from the given path
///
/// @param path Path to the archive file.
/// @return A handle to the opened archive.
Archive OpenArchive(const std::string &path);
```

A project with a public API under `include/` documents it in full as a consumer contract; see the Doxygen fragment for the rules that apply there.
