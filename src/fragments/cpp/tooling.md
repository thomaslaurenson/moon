# C++ tooling and dependencies

What a project is allowed to depend on, and what it builds with. Assumes the CMake fragment, which owns the mechanics: submodules under `extern/`, pinning, existence checks, and the `SYSTEM` include rules.

Applies to every tier. The embedded assets example is a CLI's completion scripts, but the mechanism is the same for a library that embeds a resource of its own.

## The toolchain

| Tool | Purpose |
|---|---|
| `cmake` | Configure and build |
| `ctest` | Run the test binaries |
| `clang-format` | Format source files |
| `clang-tidy` | Static analysis and naming |

That is the whole list, and it is deliberately short. There is no second formatter, no second analyser, and no build system behind CMake. The two clang tools are version-resolved in the Makefile; see the CMake fragment for how, and why clang is pinned to one major version.

## Choosing a dependency

**Standard library first.** Reach for a dependency where the standard library genuinely has no answer, not to save a few lines. `<filesystem>`, `<optional>`, `<string_view>`, `<charconv>` and `<chrono>` cover a great deal of what a small tool reaches for, and each one avoided is a submodule nobody has to update.

Where the answer is a dependency, the bar is that it does something the project should not be writing itself:

| Dependency | Why it earns its place |
|---|---|
| Catch2 | Test framework and runner, with CTest integration |
| CLI11 | Declarative argument parsing and generated help |
| subprocess.h | Spawning a process portably, for functional tests |
| A format or protocol library | Reading a file format correctly is the hard part, not the plumbing |

Prefer header-only or single-purpose libraries that build with the project. A dependency that wants a system package, a package manager, or its own build step is a dependency that breaks somebody's build; everything here compiles from `extern/` with no prerequisites beyond a compiler.

A dependency is worth taking when the standard library has no answer and the problem is genuinely hard. It is not worth taking to avoid twenty lines, because here it costs a submodule, a pin, an existence check, a `SYSTEM` include, and a Dependabot pull request every time upstream moves.

## Embedded assets

A binary that needs data at runtime embeds it rather than shipping files alongside it. C++ has no `#embed` before C23 and no `go:embed` equivalent, so the mechanism is `configure_file` into a generated header holding a raw string literal:

```cmake
set_property(DIRECTORY APPEND PROPERTY CMAKE_CONFIGURE_DEPENDS
    "${PROJECT_SOURCE_DIR}/completion/myproj.bash")
file(READ "${PROJECT_SOURCE_DIR}/completion/myproj.bash" BASH_COMPLETION_SCRIPT)
configure_file("${PROJECT_SOURCE_DIR}/cmake/completion_data.h.in"
               "${PROJECT_BINARY_DIR}/include/myproj/completion_data.h" @ONLY)
```

```cpp
// cmake/completion_data.h.in
// Generated from completion/myproj.bash, do not edit
#pragma once

namespace myproj {

inline constexpr char bash_completion_script[] = R"BASH_MYPROJ(@BASH_COMPLETION_SCRIPT@)BASH_MYPROJ";

} // namespace myproj
```

Two details are load-bearing, and both fail quietly when missed.

**Give the raw string literal a delimiter of its own.** A bare `R"(...)"` ends at the first `)"` in the embedded content, which is an ordinary sequence in shell and PowerShell. The result is a file that either fails to compile with an error pointing at the generated header rather than at the script, or, worse, compiles with the asset silently truncated. A project-specific delimiter such as `BASH_MYPROJ` cannot appear by accident.

**Declare the source with `CMAKE_CONFIGURE_DEPENDS`.** `file(READ)` runs at configure time only, so without it an edited script is not re-read and the binary keeps shipping the previous version, with a successful build every time.

The template lives in `cmake/` and the generated header goes to the build tree; neither belongs in `src/`. See the tier fragment for the include path.

Where a project embeds something that has to be valid, add a `check_embed` target to the Makefile and to `check_all`; see the Makefile targets fragment. The compiler proves only that the file was read, never that its contents work, so an embedded shell script with a syntax error compiles in and fails at the user's prompt. The Makefile targets fragment has the recipe for the completion scripts every CLI embeds; a library embedding something else writes the matching check for whatever the asset is, and a project that embeds nothing omits the target.

## Vulnerability scanning

There is no C++ equivalent of `govulncheck`, and it is worth saying so plainly rather than leaving each project to look for one. Nothing here can tell you whether a vulnerability in a vendored library is reachable from this code, because the analysis that would answer that does not exist for a tree of pinned submodules.

What is available is narrower:

- **Dependabot with the `gitsubmodule` ecosystem** opens a pull request when a submodule's upstream moves; see the Dependabot fragment. It tracks commits, not advisories, so it tells you a bump exists and never why it matters.
- **Upstream release notes** are where a security fix is actually announced. With a handful of dependencies, watching their releases is a realistic thing to do and is the only route that reports severity.

So the practice is to keep the pins current rather than to scan. A submodule that has not moved in two years is the risk, and the Dependabot pull request is what surfaces it.

Do not add a scanner that reports on the whole vendored tree without reachability. It produces a list dominated by findings in code the project never calls, and a list nobody can act on is one nobody reads.

## The C++ standard

The standard is set once in the root `CMakeLists.txt`; see the CMake fragment for the exact settings. Raise it deliberately, as a decision about which compilers a project still supports, and never through a `-std=` flag on a single target. A project with one target built at a different standard from the rest is one translation unit away from an ODR violation that links cleanly and misbehaves at runtime.
