# C++ style

C++-specific style. Assumes the core conventions. Where the project includes the clang tooling, formatting is enforced by clang-format and naming by clang-tidy (see the clang-format and clang-tidy fragments); the rules below define the intended style and apply whether or not that tooling is wired up.

Throughout the C++ fragments, `myproj` stands for the project's name. It is at once the CMake project, the namespace, the `include/myproj/` directory where the tier has one, the `MYPROJ_` prefix on every option, the prefix on every CMake target except the one a user reaches for, and the binary where the tier ships one. Substitute the real name in every position at once: a project that is one name in its namespace and another in its options has two names, and nothing lines up.

## Formatting

Enforced by clang-format; `make format` reformats all source. The canonical `.clang-format`, the deviations from LLVM defaults and the reasoning for each are in the clang-format fragment, which is the only place they are written down.

## Suppressing clang-format

`// clang-format off`/`on` is permitted only where manual alignment aids readability (lookup tables); keep the block short and comment why. Never use it for personal preference.

## Naming

Enforced by clang-tidy's `readability-identifier-naming`.

- Functions and methods: `PascalCase`.
- Variables and parameters: `snake_case`.
- Public members, and every member of a `struct`: `snake_case`, with no suffix.
- Private and protected members: `snake_case_` with a trailing underscore. The suffix marks a member as an implementation detail, so putting it on a public one says the opposite of what it means. This is the split clang-tidy enforces through `MemberCase` and `PrivateMemberSuffix`; see the clang-tidy fragment.
- Types, classes, structs, enums: `PascalCase`.
- Enum values: `UPPER_SNAKE_CASE`.
- Constants: `snake_case`, declared `constexpr` or `const`.
- No Hungarian notation (`hArchive`, `szName`); the type is already stated. Third-party APIs that use it are exempt.

## Files

- `snake_case` for all source and header file names.
- `.cpp` for implementation, `.h` for headers. Never `.hpp`/`.hxx`.
- Header-only `.h` for simple structs and helpers under ~50 lines; split into `.h`/`.cpp` when the implementation has real complexity.

## Includes

**Include what you use.** A file that names `std::array` includes `<array>` itself, even when a project header it already includes happens to provide it.

Relying on a transitive include is not a style preference, it is a portability bug that only one of the three standard libraries reports. libstdc++ and libc++ pull in far more than they promise; MSVC's STL does not. The result is a file that compiles on Linux and macOS for years and fails the first time anyone builds it on Windows, with an error that names a type rather than the missing header:

```
error C2079: 'data' uses undefined class 'std::array<uint8_t,1>'
```

Nothing changed in that file to break it: a project header simply stopped including `<array>`, or a Windows job was added. The same rule covers `<cstdint>` for the fixed-width integer types, `<string>`, `<vector>`, `<span>` and `<algorithm>`, which are the ones most often inherited by accident.

## Filesystem and binary I/O

Two Windows failures that a POSIX-only build never surfaces.

**Always open binary data with `std::ios::binary`.** Without it, Windows translates line endings on the way through and silently corrupts anything that is not text. There is no error and no warning; the file is simply wrong, and typically only in the bytes that happen to be `0x0A`.

```cpp
std::ifstream file(path, std::ios::binary);
```

**Take paths as `std::filesystem::path`, never `std::string`.** On Windows a narrow string is interpreted in the active ANSI code page, so a path containing anything outside it cannot be opened at all. `std::filesystem::path` stores `wchar_t` there and goes through the wide API, which handles any Unicode path:

```cpp
uint32_t Crc32File(const std::filesystem::path &path); // opens any path
uint32_t Crc32File(const std::string &path);           // fails outside the code page
```

The failure reaches users and not CI: runner paths are always ASCII, so a Windows job proves nothing here. It is one of the few portability classes that has to be got right by construction rather than caught by testing.

Build paths with `operator/` rather than string concatenation, so the separator is the platform's own.

Where a path has to become text, for a log line or an exception message, call `path.string()`. Never `std::string{path}`: that reaches `path::operator string_type()`, and `string_type` is `std::wstring` on Windows, so the conversion compiles everywhere the rule was not needed and fails on the one platform it was written for. `path.string()` returns a `std::string` on every platform. The same applies to `+`, which has no overload for a literal and a path at all, so a message built by concatenation needs the explicit call.

## Namespaces

Everything a project compiles into a library goes in a namespace named after the library, in `snake_case`:

```cpp
namespace myproj {
// ...
} // namespace myproj
```

A library that leaves `OpenArchive()` at global scope is broken for its consumers: the name collides with any other dependency that had the same idea, and the collision surfaces at link time in someone else's build. The namespace is not decoration, it is what makes the library linkable alongside code you have never seen.

- Close every namespace with a `} // namespace myproj` comment; the opening brace is often hundreds of lines away.
- Never use `using namespace` at file scope in a header. It forces the import on every consumer that includes it. Inside a `.cpp`, or inside a function, it is fine.
- Nest sparingly. One level for the library, plus `detail` for implementation types that must be in a header but are not API. Deep nesting reads as directory structure leaking into code.
- Prefer an anonymous namespace over `static` for file-local helpers in a `.cpp`; it applies to types as well as functions.

An application binary's own translation units (`app/`) need no namespace: nothing links against them.

## Project version

Declare the version once in the root `CMakeLists.txt` via `project(myproj VERSION 1.2.3)`. Bake it into the target at configure time with `configure_file` and a `version.h.in`, so both a binary and a library's consumers can query it as a compile-time constant. Never hardcode a version string in a `.cpp`, and never read it from `git describe` at runtime.

## Comments

The core conventions govern implementation comments; this narrows them for headers.

- Use `///` triple-slash for anything documenting a declaration, never `//` or `/** */`. The first line is a single summary sentence with no full stop.
- A comment on a declaration lives on the declaration, never duplicated onto the definition in the `.cpp`.

```cpp
/// Opens an archive from the given path
///
/// @param path Path to the archive file.
/// @return A handle to the opened archive.
Archive OpenArchive(const std::filesystem::path &path);
```

A project with a public API under `include/` documents it in full as a consumer contract; see the Doxygen fragment for the rules that apply there.
