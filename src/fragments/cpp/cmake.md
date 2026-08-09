# CMake conventions

Conventions for CMake-based C++ projects. Universal to every tier; the target definitions themselves (what `src/` builds, whether there is an `app/` binary or a public include path, and the test layers each implies) live in the tier fragment: cmake-lib, cmake-app, or cmake-lib-cli.

## Design principles

- CMake is the build system for all C++ projects; never use raw compiler invocations
- The Makefile is a task runner that wraps CMake; CI calls `make <target>`, never raw `cmake` commands
- All build output lives under `build/`, one subdirectory per configuration; see Build directory
- Dependencies are always git submodules pinned to a specific commit, never system-installed libraries

## Repository layout

Every C++ project contains at least these at the root; a project-tier fragment (cmake-lib, cmake-app, or cmake-lib-cli) adds `include/`, `app/`, and the release Dockerfiles as its tier requires:

```
.clang-format
.clang-tidy
.github/
  workflows/
  dependabot.yml
CMakeLists.txt        # root: project settings, options, add_subdirectory calls
Makefile
CHANGELOG.md
README.md
cmake/                # version.h.in, plus any CMake helper modules
src/                  # implementation, built as a library target; never contains main()
extern/               # git submodules only; never copy third-party headers manually
test/                 # see cpp/testing.md for internal structure
```

- `extern/` contains only git submodules; never manually copied headers or installed libraries
- `cmake/` holds build inputs that are neither source nor public headers: `version.h.in`, which every project has (see the version rule in cpp/style.md), plus helper modules such as `mark_system.cmake` where a project links a dependency that exports its own target. The version template is what makes this directory universal rather than optional. Keep templates out of `include/`: that tree is the API a consumer includes, and a file there that cannot be included misrepresents it.

Two rules hold across every tier, and the tier fragments assume them:

- **`src/` never contains `main()`.** The entry point lives in `app/`, in the tiers that have one. Keeping it out of `src/` is what allows the whole implementation to be compiled once, linked by both the binary and the test binaries, and reused by another project later.
- **`src/` always builds a library target.** For a library that target is the deliverable; for an application it is an internal detail with no `include/` and no alias. Either way, tests link it rather than recompiling its sources, so the test build cannot drift from the real one.

The two questions that place a project in a tier are therefore independent: does it expose a public API (`include/`, so cmake-lib or cmake-lib-cli), and does it ship a binary (`app/`, so cmake-app or cmake-lib-cli)?

## Minimum version

All projects must declare a minimum CMake version of 3.21:

```cmake
cmake_minimum_required(VERSION 3.21)
```

CMake 3.21 is the oldest version found on any supported build environment. Never use `cmake_minimum_required(VERSION 3.10)` or other outdated minimums; they unlock legacy behaviour that conflicts with modern CMake practices.

## C++ standard

All projects must set a minimum C++ standard of 17. New projects should prefer 20:

```cmake
set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)
```

- `CMAKE_CXX_STANDARD_REQUIRED ON`: fails the build if the compiler does not support the requested standard, rather than silently falling back
- `CMAKE_CXX_EXTENSIONS OFF`: disables compiler-specific extensions such as GNU extensions, ensuring the code is portable standard C++

## Required project settings

Every root `CMakeLists.txt` must set these options immediately after `project()`:

```cmake
cmake_minimum_required(VERSION 3.21)

project(MyProject VERSION 1.0.0)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)

if(NOT CMAKE_BUILD_TYPE)
    set(CMAKE_BUILD_TYPE Debug)
endif()

set(CMAKE_POSITION_INDEPENDENT_CODE ON)
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)
set(CMAKE_RUNTIME_OUTPUT_DIRECTORY "${PROJECT_BINARY_DIR}/bin")

# Multi-config generators (Visual Studio, Xcode) append the config name to the
# output directory, so the bare variable above would put the binary in
# bin/Release/ rather than bin/.
foreach(cfg IN ITEMS DEBUG RELEASE RELWITHDEBINFO MINSIZEREL)
    set(CMAKE_RUNTIME_OUTPUT_DIRECTORY_${cfg} "${PROJECT_BINARY_DIR}/bin")
endforeach()
```

- `CMAKE_BUILD_TYPE` defaults to `Debug`: this ensures `compile_commands.json` is always generated with full debug information for clang-tidy
- `CMAKE_POSITION_INDEPENDENT_CODE ON`: required for shared libraries and good practice for all targets
- `CMAKE_EXPORT_COMPILE_COMMANDS ON`: generates `compile_commands.json` in the build directory, required for clang-tidy
- `CMAKE_RUNTIME_OUTPUT_DIRECTORY`: all executables (the app binary, or a library's test binaries) land in the configuration's own `bin/` (`build/dev/bin/`) regardless of how many targets the project defines
- The per-config loop is what keeps that true on a multi-config generator. Without it, a Visual Studio build emits `build/dev/bin/Release/myapp.exe`, and every consumer of the path (a functional test's baked-in binary path, a CI step that moves the artifact) silently looks in the wrong place. Set all four configs, not just `RELEASE`, so a Debug build in an IDE behaves the same way

## Referring to project paths

Use `PROJECT_SOURCE_DIR` and `PROJECT_BINARY_DIR` to refer to this project's own directories. Never use `CMAKE_SOURCE_DIR` or `CMAKE_BINARY_DIR`:

```cmake
# Good - resolves to this project's root, however it is being built
target_include_directories(mylib PUBLIC "${PROJECT_SOURCE_DIR}/include")

# Bad - resolves to the top-level project's root, which may not be this one
target_include_directories(mylib PUBLIC "${CMAKE_SOURCE_DIR}/include")
```

`CMAKE_SOURCE_DIR` is the root of the *outermost* project, not of the project the file belongs to. The two are the same only while a project is built directly. The moment a consumer pulls this project in as a submodule and calls `add_subdirectory(extern/mylib)`, `CMAKE_SOURCE_DIR` becomes the consumer's root, and `"${CMAKE_SOURCE_DIR}/include"` silently points at the consumer's `include/` directory instead of this library's. Submodule plus `add_subdirectory` is exactly how a library here is meant to be consumed, so this is not a hypothetical.

`PROJECT_SOURCE_DIR` tracks the nearest enclosing `project()` call and is correct in both cases. Use it everywhere, including in an application, where the two currently coincide: the habit costs nothing and the failure it prevents is a silent one.

Within a single directory's `CMakeLists.txt`, prefer bare relative paths for source files (`parser.cpp`, not `"${PROJECT_SOURCE_DIR}/src/parser.cpp"`); CMake resolves them against the current directory, and the shorter form is what makes an accidental `src/src/parser.cpp` obvious on sight.

## Build directory

All build output goes under `build/`, one subdirectory per configuration, named for whatever makes that configuration different:

```bash
cmake -B build/dev -D<PROJECT>_BUILD_TESTING=ON
cmake --build build/dev
```

| Directory | Why it is separate |
|---|---|
| `build/dev` | The default: everything testable, `compile_commands.json`, the daily build |
| `build/release` | Different build type |
| `build/fuzz` | Needs Clang and `-fsanitize=fuzzer` |
| `build/asan` | Different code generation |
| `build/32` | Different architecture |

Only create a directory when the configuration genuinely cannot share one. A configuration that differs by a `-D` option affecting compiler, architecture or instrumentation cannot: reconfiguring in place silently replaces the previous one, and nothing afterwards tells you which of them you are looking at. Naming the directory for the configuration puts that answer in the path.

**The test layers are not a reason to split.** Unit, integration and functional tests all build into `build/dev` and are selected when they are *run*, with `ctest -L unit`. An integration layer gated behind an option is still built into the same directory; the option controls whether the target exists, not where it lives. Never create `build/test` or a directory per test layer.

`.gitignore` needs one entry, `build/`, and `rm -rf build` removes everything.

The default directory is `dev`, not `debug`, because it is named for what it is for rather than for a build type. A multi-config generator (Visual Studio, Xcode) picks the build type at build time, so the same directory serves `--config Debug` and `--config Release`; a CI job that builds Release and runs the test suite still belongs in `build/dev`.

Because binaries land in `${PROJECT_BINARY_DIR}/bin`, a path that was `build/bin/myapp` becomes `build/dev/bin/myapp`. `compile_commands.json` for clang tooling comes from `build/dev`, which is why that configuration always has testing on.

## CMakeLists.txt structure

Every directory that produces a target or manages a distinct concern has its own `CMakeLists.txt`. The root never defines targets directly; it orchestrates.

```
CMakeLists.txt        # project settings, dependencies, add_subdirectory calls
src/
  CMakeLists.txt      # defines the library target; see the tier fragment
app/
  CMakeLists.txt      # defines the binary; only in tiers that ship one
test/
  CMakeLists.txt      # defines test targets
extern/
  Catch2/             # submodule - never modify
```

### Root CMakeLists.txt responsibilities

- `cmake_minimum_required` and `project`
- All required project settings (standard, build type, output directory)
- Project-wide options via `option()`
- Submodule existence checks and `add_subdirectory` for dependencies
- `enable_testing()`, and `include(Catch)`, when tests are on
- `add_subdirectory(src)`, then `add_subdirectory(app)` in tiers that have one
- `add_subdirectory(test)` when testing is on

### src/CMakeLists.txt responsibilities

`src/CMakeLists.txt` is where the target this project exposes is defined. That holds whichever shape the project has taken, which is what keeps the root purely orchestrational:

- **Single target** — `add_library` directly, listing every source. Subdirectories under `src/` are source organisation and have no `CMakeLists.txt` of their own.
- **Modular** — `add_subdirectory` for each module, then the aggregate target that links them. See the library tier fragment.

Either way this file owns:

- `add_library`; see the tier fragment for the target name and whether it carries a public include path
- `target_include_directories`
- `target_link_libraries`
- `configure_file` for generated headers
- the module `add_subdirectory` calls, in a modular project, ordered so a module is added before anything linking its alias

A modular project puts the module list here rather than in the root for two reasons: the root stays a fixed size as modules are added, and adding one means editing the file next to the modules instead of a file otherwise concerned with dependency setup.

## Testing option

Every project declares a project-scoped testing option in the root `CMakeLists.txt`, named `<PROJECT>_BUILD_TESTING` and defaulting to `PROJECT_IS_TOP_LEVEL`:

```cmake
option(MYLIB_BUILD_TESTING "Build mylib tests" ${PROJECT_IS_TOP_LEVEL})

# enable_testing must be called before any add_subdirectory, so CTest
# discovers the tests those subdirectories register.
if(MYLIB_BUILD_TESTING)
    enable_testing()

    add_subdirectory(extern/Catch2)
    mark_system(Catch2)

    list(APPEND CMAKE_MODULE_PATH "${PROJECT_SOURCE_DIR}/extern/Catch2/extras")
    include(Catch)
endif()

# ... module add_subdirectory calls ...

if(MYLIB_BUILD_TESTING)
    add_subdirectory(test)
endif()
```

Never use the bare `BUILD_TESTING` name for this. It is a single global that CTest itself declares, and claiming it has two consequences, both of which bite in practice:

- A consumer who adds this project via `add_subdirectory` with testing on for their own code silently gets this project's tests built and run as part of theirs.
- Declaring it as a cache variable here turns every vendored dependency's own `option(BUILD_TESTING ... OFF)` into a no-op, because the cache entry already exists. The dependency inherits this project's `ON` and builds its demos and self-tests into this project's CTest run. Working around that needs a save-force-restore dance around each `add_subdirectory`, and the whole problem disappears with a project-scoped name.

`PROJECT_IS_TOP_LEVEL` (CMake 3.21, the declared minimum here) makes the default correct automatically: on when the project is built directly, off when it is somebody's subdirectory. Do not hand-roll it with a `set(MYLIB_ROOT_BUILD TRUE)` marker.

`enable_testing()` must be called here, in the root, and before the `add_subdirectory` calls that register tests. CTest only writes the test manifest for the directory that enabled testing and its children, so calling it in `test/CMakeLists.txt` leaves `ctest --test-dir build/dev` finding nothing.

**Call `enable_testing()`, never `include(CTest)`.** They look interchangeable and are not: `include(CTest)` calls `enable_testing()` for you, but it also declares `BUILD_TESTING` as a cache variable defaulting to `ON`, which is precisely the global the project-scoped name above exists to avoid. A project that scopes its own option correctly and then calls `include(CTest)` has reintroduced the problem through the back door, and the symptom is a vendored dependency's self-tests appearing in `ctest -N` output — with nothing in the project's own CMake mentioning `BUILD_TESTING` to explain why. `include(CTest)` also adds CDash dashboard targets (`Experimental`, `Nightly`, `Continuous`) that no project here uses. `enable_testing()` plus `include(Catch)` is the whole requirement.

Do not call `include(CTest)`. Its purpose is to declare the global `BUILD_TESTING` option and call `enable_testing()` for you, which is exactly what this section replaces; it also drags in CDash submission targets no project here uses. `include(Catch)` is the only include needed, once, at the root, after `CMAKE_MODULE_PATH` picks up Catch2's `extras`. `test/CMakeLists.txt` then just calls `catch_discover_tests`.

## Target names

Every target name is global to the CMake build, including a consumer's. Prefix every target with the project name:

```cmake
# Good - cannot collide with anything
add_library(mylib_crypto STATIC ...)
add_library(mylib::crypto ALIAS mylib_crypto)

# Bad - claims a name any other project might want
add_library(crypto STATIC ...)
```

Unprefixed module names like `crypto`, `common`, `config`, `net` or `parser` are the ones most likely to collide, because they are the names every project reaches for. The collision does not appear while the project is built directly. It appears the first time two of them are pulled into the same superbuild, as a duplicate-target configure error in somebody else's build, naming a target neither of them wrote.

Consumers link the alias, never the raw name, so the prefix costs nothing at the call site.

## Warnings

Every project-owned target sets its warning bar explicitly, and a project-scoped option promotes warnings to errors:

```cmake
option(MYLIB_WERROR "Treat warnings as errors" OFF)

target_compile_options(mylib PRIVATE -Wall -Wextra)

if(MYLIB_WERROR)
    target_compile_options(mylib PRIVATE -Werror)
endif()
```

`PRIVATE`, so the bar applies to this project's own code and is never imposed on a consumer. Default `OFF` for `-Werror`, turned on in CI: a new compiler version routinely adds a warning, and a developer whose build breaks because they upgraded clang cannot get any work done.

Vendored C or C++ compiled into a project target is exempt. Do not fix a third-party file's warnings, and do not lower the project's bar to accommodate it; silence it at the source:

```cmake
# blast.c is third-party C; do not hold it to the project's warning bar
set_source_files_properties("${BLAST_C}" PROPERTIES COMPILE_OPTIONS "-Wno-unused-parameter")
```

## Sanitizers

Every project declares a project-scoped sanitizer option, applied globally so that every target and every test is instrumented consistently:

```cmake
option(MYLIB_ASAN "Build with Address + UB sanitizers" OFF)

# Applied before any target is declared, so every module and test is instrumented
if(MYLIB_ASAN)
    if(MSVC)
        # MSVC has AddressSanitizer but no UndefinedBehaviorSanitizer, and links
        # its runtime automatically, so there is no matching add_link_options.
        add_compile_options(/fsanitize=address /Oy-)
    else()
        add_compile_options(-fsanitize=address,undefined -fno-omit-frame-pointer -g)
        add_link_options(-fsanitize=address,undefined)
    endif()
endif()
```

The compiler branch is not optional on a project that builds on Windows. `-fsanitize=address,undefined` is GCC and Clang syntax; MSVC rejects it, so without the branch turning the option on fails the build outright rather than producing an uninstrumented one. `-fno-omit-frame-pointer` is `/Oy-` there, and UB sanitizing is simply unavailable — a Windows sanitizer run catches memory errors only, which is worth stating in a bug report that compares platforms.

This is the one legitimate use of the directory-scoped `add_compile_options` rather than `target_compile_options`. A sanitizer is not a per-target property: instrumenting the library but not the test binary that links it produces link errors and false negatives. It has to be all or nothing, and it has to be set before the first target is declared.

Default `OFF`, because ASan costs roughly 2x runtime and 3x memory. Run it locally when hunting a bug, and in a dedicated CI job rather than the main test job.

### Makefile targets

A sanitized build changes code generation, so it gets its own directory and cannot share `build/dev`:

```makefile
.PHONY: configure_asan
configure_asan: ## Configure build/asan with Address + UB sanitizers
	cmake -B build/asan \
	  -DCMAKE_BUILD_TYPE=Debug \
	  -DMYLIB_ASAN=ON \
	  -DMYLIB_BUILD_TESTING=ON

.PHONY: test_asan
test_asan: configure_asan ## Build and run the unit tests under sanitizers
	cmake --build build/asan --parallel $(JOBS)
	ctest --test-dir build/asan --output-on-failure --parallel $(JOBS) -L unit
```

`test_asan` runs the unit layer only. That layer needs no external data or server, so it is the one that can run anywhere, and sanitizer findings in it point at the project's own code rather than at a fixture. Without these targets the option is reachable only through a raw `cmake -D` invocation, which the Makefile exists to prevent.

## Dependencies

All external dependencies are git submodules pinned to a specific commit, stored under `extern/`:

```
extern/
  ThirdPartyLib/
  Catch2/
```

Always pin to an immutable reference: a release tag or a commit hash, never a moving branch name. Branch names move; a release tag or hash does not:

```bash
cd extern/Catch2 && git checkout v3.6.0
```

Every dependency must have an existence check in the root `CMakeLists.txt` before its `add_subdirectory` call. The error message must name the dependency, explain why it is needed, and tell the developer exactly how to fix it:

```cmake
if(NOT EXISTS "${PROJECT_SOURCE_DIR}/extern/ThirdPartyLib/CMakeLists.txt")
    message(FATAL_ERROR
"Missing dependency: ThirdPartyLib
This project requires the ThirdPartyLib library.
It is provided as a submodule of this repository.
Did you forget to run the following commands?
   git submodule init
   git submodule update")
endif()

add_subdirectory(extern/ThirdPartyLib)
```

Single-header libraries check for the header file directly rather than a `CMakeLists.txt`. Never assume submodules are initialised. Always guard every dependency.

### Including extern/ headers

Use the `SYSTEM` keyword on every `target_include_directories` call that points into `extern/`. This marks those paths as system headers, so clang-tidy and the compiler suppress all warnings from third-party code by default:

```cmake
target_include_directories(mytarget SYSTEM PRIVATE
    "${PROJECT_SOURCE_DIR}/extern/ThirdPartyLib/src"
)

# Project-owned headers (generated files) use PRIVATE without SYSTEM:
target_include_directories(mytarget PRIVATE
    "${PROJECT_BINARY_DIR}"
)
```

- `SYSTEM PRIVATE` tells CMake to pass `-isystem` instead of `-I` for those paths
- clang-tidy excludes system headers from all analysis by default; without `SYSTEM`, third-party headers generate thousands of suppressed warnings that inflate output and slow analysis
- Never combine `extern/` and project-owned paths in one `target_include_directories` call; they require different keywords

### Dependencies that export their own target

The rule above only covers a path pointed at directly. A submodule with its own `CMakeLists.txt` exports a target (`Catch2`, `zlibstatic`, `libtommath`) that carries its include directories as an interface property, and consuming it with `target_link_libraries` picks those up as ordinary `-I` paths. The `SYSTEM` keyword has nothing to attach to.

Promote them by moving the property, immediately after the `add_subdirectory` that created the target:

```cmake
# cmake/mark_system.cmake

# Re-declare a dependency's interface includes as SYSTEM includes, so clang-tidy
# and the compiler ignore them. Needed for any submodule that exports its own
# target: linking it otherwise pulls its headers in as ordinary -I paths.
function(mark_system target)
    get_target_property(_incs ${target} INTERFACE_INCLUDE_DIRECTORIES)
    if(_incs)
        set_target_properties(${target} PROPERTIES
            INTERFACE_SYSTEM_INCLUDE_DIRECTORIES "${_incs}")
    endif()
endfunction()
```

```cmake
add_subdirectory(extern/Catch2)
mark_system(Catch2)
```

Put the function in `cmake/` and `include()` it from the root rather than repeating the property dance at each dependency; a project with four submodules otherwise carries four copies of it.

### Suppressing a dependency's own tests

A submodule that builds its own test suite adds noise to this project's CTest run and, under a global sanitizer option, may not even link. Force its testing option off before adding it, using whatever name that project uses:

```cmake
set(ZLIB_BUILD_TESTING OFF CACHE BOOL "" FORCE)
add_subdirectory(extern/zlib)
```

This is only reliable when this project's own testing option is project-scoped. A root that declares the global `BUILD_TESTING` as a cache variable turns a dependency's own `option(BUILD_TESTING ... OFF)` into a no-op, because the cache entry already exists; the dependency then inherits this project's `ON` and builds its demos regardless. See the testing option section.

### Header-only dependencies

Wrap a header-only dependency in an `INTERFACE` library once, at the root, rather than repeating a `SYSTEM PRIVATE` include path at each target that needs it:

```cmake
add_library(asio INTERFACE)
target_include_directories(asio SYSTEM INTERFACE
    "${PROJECT_SOURCE_DIR}/extern/asio/asio/include"
)
target_compile_definitions(asio INTERFACE
    ASIO_STANDALONE          # No Boost headers
    ASIO_NO_DEPRECATED       # Fail loudly on any deprecated Asio API usage
)
```

Consumers then write `target_link_libraries(mylib_transport PRIVATE asio)` and inherit the include path, the `SYSTEM` marking, and any required compile definitions together. Those definitions are the real argument for this: a project that repeats the include path at five targets and the definitions at four has a bug waiting in the fifth.

## Clang tooling

Clang tools are pinned to major version 18 across all projects, because formatting output and check behaviour differ between major versions: a tree formatted with one and checked with another fails `fmt_check` on lines nobody touched.

How that version is installed differs per platform, so the Makefile resolves the binary rather than naming it:

- Debian and Ubuntu install versioned binaries (`clang-format-18`) from apt.llvm.org
- Homebrew and the LLVM Windows installer provide unversioned `clang-format` from a versioned install

There is no `install_clang_tools` target. Installing a system toolchain is the environment's job, not the build's: a target that runs `sudo apt-get` fails outright on macOS and Windows runners, and shipping one per platform is a package manager written in Make.

### Resolving the binaries

```makefile
CLANG_VERSION ?= 18
CLANG_FORMAT  ?= $(shell command -v clang-format-$(CLANG_VERSION) 2>/dev/null || echo clang-format)
CLANG_TIDY    ?= $(shell command -v clang-tidy-$(CLANG_VERSION) 2>/dev/null || echo clang-tidy)
```

Prefer the versioned name, fall back to the plain one, and let either be overridden from the command line (`make fmt CLANG_FORMAT=/opt/homebrew/opt/llvm/bin/clang-format`). Falling back to the bare name rather than failing keeps the failure legible: an absent tool reports `clang-format: command not found`, which is clearer than a Make-level error about an empty variable.

CI pins the version explicitly, so drift between a contributor's local major version and the enforced one surfaces there rather than in review.

### Makefile targets

Use the resolved variables in all targets, never a literal binary name. Both targets below take their directory list from `wildcard`, so one Makefile covers every tier: a library has no `app/`, an application has no `include/`, and the expansion simply omits what is absent rather than failing. `JOBS` is declared here, once, because `build` is the first target that needs it; every later fragment's `cmake --build` and `ctest` targets reuse the same variable rather than redeclaring it.

```makefile
# Project-owned C++ directories, in whichever of them this tier actually has
CPP_DIRS      := $(wildcard include src app test)
CPP_LINT_DIRS := $(wildcard src app)
JOBS          ?= $(shell nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 4)

.PHONY: configure
configure: ## Configure the cmake build
	cmake -B build/dev \
	  -DCMAKE_BUILD_TYPE=Debug \
	  -DCMAKE_EXPORT_COMPILE_COMMANDS=ON \
	  $(CMAKE_ARGS)

.PHONY: build
build: ## Build the project
	cmake --build build/dev --parallel $(JOBS)

.PHONY: fmt
fmt: ## Format all source files with clang-format
	find $(CPP_DIRS) \( -name "*.cpp" -o -name "*.h" \) | xargs $(CLANG_FORMAT) -i

.PHONY: fmt_check
fmt_check: ## Check formatting without modifying files
	find $(CPP_DIRS) \( -name "*.cpp" -o -name "*.h" \) | xargs $(CLANG_FORMAT) --dry-run --Werror

.PHONY: lint_cpp
lint_cpp: ## Run clang-tidy static analysis (requires: make configure)
	$(CLANG_TIDY) --quiet -p build/dev \
	--header-filter="$(CURDIR)/(include|src|app)/.*" $$(find $(CPP_LINT_DIRS) -name "*.cpp") 2>&1 \
	| grep -v " warnings generated"; \
	exit $${PIPESTATUS[0]}

# GET

.PHONY: get_changelog
get_changelog: ## Print the CHANGELOG.md entry for TAG=vX.Y.Z (fails if missing)
	@test -n "$(TAG)" || { echo "TAG is required" >&2; exit 2; }
	@awk -v raw="$(TAG)" '\
	  BEGIN { v = raw; sub(/^v/, "", v) } \
	  /^## / { if (found) exit; if ($$2 == v) { found = 1; print; next } } \
	  found { print } \
	  END { if (!found) exit 1 }' CHANGELOG.md
```

- `--quiet` suppresses the "Suppressed N warnings" summary and hint lines
- `find` covers every implementation file in those directories, including nested subdirectories; a bare `src/*.cpp` glob would miss anything below the top level
- `include/` is formatted but not tidied directly: its headers carry no `.cpp` of their own, and clang-tidy reaches them through the `--header-filter` when it analyses the `src/` files that include them
- `--header-filter="$(CURDIR)/(include|src|app)/.*"` limits diagnostic output to project headers; extern/ headers are already excluded as system headers (see Including extern/ headers in this file) but this provides belt-and-suspenders coverage
- `grep -v " warnings generated"` strips the per-file progress counter, which counts all warnings before any filtering and is always misleading when third-party headers are present; `exit $${PIPESTATUS[0]}` preserves clang-tidy's exit code through the pipe
- `make configure` must be run before `make lint_cpp`; clang-tidy reads `build/compile_commands.json` to resolve include paths
- `CMAKE_ARGS` passes extra `-D` flags through to `cmake` (for example CI's `-DMYAPP_BINARY_PATH_OVERRIDE=...`); it is empty for a normal local configure
- `get_changelog` is defined here, not left to the project, because `release.yml` calls it directly (see cpp/workflows.md) and a release that reaches that step without the target fails after the artifacts are already built. It uses only POSIX `awk`, and strips a leading `v` from `TAG` because git tags are `v1.2.3` while changelog headers are bare `## 1.2.3 - ...` (see github/changelog.md). It exits non-zero on an empty `TAG` or an unmatched version, so a release never publishes empty notes

Every target a workflow invokes must be defined by one of these fragments. A workflow calling `make <something>` that no fragment defines is a scaffolding bug that only surfaces on a real release, in the job that publishes it.

Note: `fmt` and `fmt_check` include the `test/` directory; test code is held to the same formatting standard as production code. `lint_cpp` deliberately does not run clang-tidy over `test/`: test files use Catch2 macros and fixture patterns that trip naming and readability checks written for production code. Format tests, but do not tidy them.

### Configuration files

Both `.clang-format` and `.clang-tidy` live at the project root. CMake is pointed at the build directory via `-p build/dev` so clang-tidy can find `compile_commands.json`. The `FormatStyle: file` setting in `.clang-tidy` tells clang-tidy to use the root `.clang-format` for any formatting checks.
