# CMake conventions

Conventions for CMake-based C++ projects. Universal to every tier; the target definitions themselves (what `src/` builds, whether there is an `app/` binary or a public include path, and the test layers each implies) live in the tier fragment: cmake-lib, cmake-app, or cmake-lib-cli.

## Design principles

- CMake is the build system for all C++ projects; never use raw compiler invocations
- The Makefile is a task runner that wraps CMake; CI calls `make <target>`, never raw `cmake` commands
- One `build/` directory for everything; no separate lint or release build directories
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
src/                  # implementation, built as a library target; never contains main()
extern/               # git submodules only; never copy third-party headers manually
test/                 # see cpp/testing.md for internal structure
```

- `extern/` contains only git submodules; never manually copied headers or installed libraries

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
set(CMAKE_RUNTIME_OUTPUT_DIRECTORY "${CMAKE_BINARY_DIR}/bin")
```

- `CMAKE_BUILD_TYPE` defaults to `Debug`: this ensures `compile_commands.json` is always generated with full debug information for clang-tidy
- `CMAKE_POSITION_INDEPENDENT_CODE ON`: required for shared libraries and good practice for all targets
- `CMAKE_EXPORT_COMPILE_COMMANDS ON`: generates `compile_commands.json` in the build directory, required for clang-tidy
- `CMAKE_RUNTIME_OUTPUT_DIRECTORY`: all executables (the app binary, or a library's test binaries) land in `build/bin/` regardless of how many targets the project defines

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

All projects use a single `build/` directory:

```bash
cmake -B build
cmake --build build
```

Never create separate build directories for lint, release, or test builds. The default `Debug` build type produces a `compile_commands.json` that covers all use cases.

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

- `add_library`; see the tier fragment for the target name and whether it carries a public include path
- `target_include_directories`
- `target_link_libraries`
- `configure_file` for generated headers

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

`enable_testing()` must be called here, in the root, and before the `add_subdirectory` calls that register tests. CTest only writes the test manifest for the directory that enabled testing and its children, so calling it in `test/CMakeLists.txt` leaves `ctest --test-dir build` finding nothing.

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
    add_compile_options(-fsanitize=address,undefined -fno-omit-frame-pointer -g)
    add_link_options(-fsanitize=address,undefined)
endif()
```

This is the one legitimate use of the directory-scoped `add_compile_options` rather than `target_compile_options`. A sanitizer is not a per-target property: instrumenting the library but not the test binary that links it produces link errors and false negatives. It has to be all or nothing, and it has to be set before the first target is declared.

Default `OFF`, because ASan costs roughly 2x runtime and 3x memory. Run it locally when hunting a bug, and in a dedicated CI job rather than the main test job.

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

Clang tools are pinned to version 18 across all projects for reproducibility. Never use the unversioned `clang-format` or `clang-tidy` binaries as the system default may differ between machines and CI runners.

### Installation

The Makefile must provide an `install_clang_tools` target:

```makefile
.PHONY: install_clang_tools
install_clang_tools: ## Install clang-format and clang-tidy at pinned version
	sudo apt-get install -y clang-format-18 clang-tidy-18
```

### Makefile targets

Use the versioned binaries explicitly in all targets. Both targets below take their directory list from `wildcard`, so one Makefile covers every tier: a library has no `app/`, an application has no `include/`, and the expansion simply omits what is absent rather than failing.

```makefile
# Project-owned C++ directories, in whichever of them this tier actually has
CPP_DIRS      := $(wildcard include src app test)
CPP_LINT_DIRS := $(wildcard src app)

.PHONY: configure
configure: ## Configure the cmake build
	cmake -B build \
	  -DCMAKE_BUILD_TYPE=Debug \
	  -DCMAKE_EXPORT_COMPILE_COMMANDS=ON \
	  $(CMAKE_ARGS)

.PHONY: build
build: ## Build the project
	cmake --build build

.PHONY: fmt
fmt: ## Format all source files with clang-format
	find $(CPP_DIRS) \( -name "*.cpp" -o -name "*.h" \) | xargs clang-format-18 -i

.PHONY: fmt_check
fmt_check: ## Check formatting without modifying files
	find $(CPP_DIRS) \( -name "*.cpp" -o -name "*.h" \) | xargs clang-format-18 --dry-run --Werror

.PHONY: lint_cpp
lint_cpp: ## Run clang-tidy static analysis (requires: make configure)
	clang-tidy-18 --quiet -p build \
	--header-filter="$(CURDIR)/(include|src|app)/.*" $$(find $(CPP_LINT_DIRS) -name "*.cpp") 2>&1 \
	| grep -v " warnings generated"; \
	exit $${PIPESTATUS[0]}
```

- `--quiet` suppresses the "Suppressed N warnings" summary and hint lines
- `find` covers every implementation file in those directories, including nested subdirectories; a bare `src/*.cpp` glob would miss anything below the top level
- `include/` is formatted but not tidied directly: its headers carry no `.cpp` of their own, and clang-tidy reaches them through the `--header-filter` when it analyses the `src/` files that include them
- `--header-filter="$(CURDIR)/(include|src|app)/.*"` limits diagnostic output to project headers; extern/ headers are already excluded as system headers (see Including extern/ headers in this file) but this provides belt-and-suspenders coverage
- `grep -v " warnings generated"` strips the per-file progress counter, which counts all warnings before any filtering and is always misleading when third-party headers are present; `exit $${PIPESTATUS[0]}` preserves clang-tidy's exit code through the pipe
- `make configure` must be run before `make lint_cpp`; clang-tidy reads `build/compile_commands.json` to resolve include paths
- `CMAKE_ARGS` passes extra `-D` flags through to `cmake` (for example CI's `-DMYAPP_BINARY_PATH_OVERRIDE=...`); it is empty for a normal local configure

Note: `fmt` and `fmt_check` include the `test/` directory; test code is held to the same formatting standard as production code. `lint_cpp` deliberately does not run clang-tidy over `test/`: test files use Catch2 macros and fixture patterns that trip naming and readability checks written for production code. Format tests, but do not tidy them.

### Configuration files

Both `.clang-format` and `.clang-tidy` live at the project root. CMake is pointed at the build directory via `-p build` so clang-tidy can find `compile_commands.json`. The `FormatStyle: file` setting in `.clang-tidy` tells clang-tidy to use the root `.clang-format` for any formatting checks.
