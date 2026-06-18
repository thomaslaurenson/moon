# CMake Conventions

Conventions for CMake-based C++ projects.

## Design Principles

- CMake is the build system for all C++ projects; never use raw compiler invocations
- The Makefile is a task runner that wraps CMake; CI calls `make <target>`, never raw `cmake` commands
- One `build/` directory for everything; no separate lint or release build directories
- Dependencies are always git submodules pinned to a specific commit, never system-installed libraries

## Repository Layout

Every C++ project must contain these files and directories at the root:

```
.clang-format
.clang-tidy
.github/
  workflows/
    build.yml
    lint.yml
    test.yml
    release.yml
    prerelease.yml
    pr.yml
    tag.yml
    main.yml
  dependabot.yml
CMakeLists.txt        # root: project settings, options, add_subdirectory calls
Dockerfile.glibc
Dockerfile.musl
Makefile
CHANGELOG.md
README.md
src/                  # business logic; see CMakeLists.txt Structure below
extern/               # git submodules only; never copy third-party headers manually
test/                 # see cpp/testing.md for internal structure
```

- `src/` contains all application source; no header-only projects unless the implementation is 50 lines or fewer
- `extern/` contains only git submodules; never manually copied headers or installed libraries
- Two Dockerfiles are required: one for glibc builds, one for musl (static) builds; see `tools/docker.md`

## Minimum Version

All projects must declare a minimum CMake version of 3.21:

```cmake
cmake_minimum_required(VERSION 3.21)
```

CMake 3.21 is the oldest version found on any supported build environment. Never use `cmake_minimum_required(VERSION 3.10)` or other outdated minimums; they unlock legacy behaviour that conflicts with modern CMake practices.

## C++ Standard

All projects must set a minimum C++ standard of 17. New projects should prefer 20:

```cmake
set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)
```

- `CMAKE_CXX_STANDARD_REQUIRED ON`: fails the build if the compiler does not support the requested standard, rather than silently falling back
- `CMAKE_CXX_EXTENSIONS OFF`: disables compiler-specific extensions such as GNU extensions, ensuring the code is portable standard C++

## Required Project Settings

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
- `CMAKE_RUNTIME_OUTPUT_DIRECTORY`: all executables land in `build/bin/` regardless of how many targets the project defines

## Build Directory

All projects use a single `build/` directory:

```bash
cmake -B build
cmake --build build
```

Executables are always output to `build/bin/`. For a project with multiple binaries they all land in the same directory:

```
build/
  bin/
    mytool
    myapp_unit_tests
    myapp_functional_tests
  compile_commands.json
```

Never create separate build directories for lint, release, or test builds. The default `Debug` build type produces a `compile_commands.json` that covers all use cases.

## CMakeLists.txt Structure

Every directory that produces a target or manages a distinct concern has its own `CMakeLists.txt`. The root never defines targets directly; it orchestrates.

```
CMakeLists.txt        # project settings, dependencies, add_subdirectory calls
src/
  CMakeLists.txt      # defines the main executable target(s)
test/
  CMakeLists.txt      # defines test targets
extern/
  Catch2/             # submodule - never modify
  subprocess.h/       # submodule - never modify
  ThirdPartyLib/      # submodule - never modify
```

### Root CMakeLists.txt responsibilities

- `cmake_minimum_required` and `project`
- All required project settings (standard, build type, output directory)
- Project-wide options via `option()`
- Submodule existence checks and `add_subdirectory` for dependencies
- `add_subdirectory(src)`
- `add_subdirectory(test)` when `BUILD_TESTING` is on

### src/CMakeLists.txt responsibilities

- `add_executable` or `add_library`
- `target_include_directories`
- `target_link_libraries`
- `configure_file` for generated headers

### test/CMakeLists.txt responsibilities

- Two test executable targets; one for unit tests, one for functional tests
- `target_compile_definitions` to bake binary path and test directory into the functional test binary at configure time
- Linking against `Catch2::Catch2WithMain`
- `catch_discover_tests` for both binaries

## Testing Option

Every project must declare the `BUILD_TESTING` option in the root `CMakeLists.txt`. This allows tests to be disabled for release builds or when building on a system without test dependencies:

```cmake
option(BUILD_TESTING "Build tests" ON)

if(BUILD_TESTING)
    add_subdirectory(extern/Catch2)
    add_subdirectory(test)
endif()
```

## Baking Paths into Test Binaries

Functional tests need to know where the compiled binary lives at runtime. Rather than discovering it at runtime, bake the path in at CMake configure time using `target_compile_definitions`. This eliminates a whole class of path-resolution bugs and makes the test binary fully self-contained:

```cmake
# In test/CMakeLists.txt

if(WIN32)
    set(MYAPP_BINARY_PATH "${CMAKE_BINARY_DIR}/bin/myapp.exe")
else()
    set(MYAPP_BINARY_PATH "${CMAKE_BINARY_DIR}/bin/myapp")
endif()

target_compile_definitions(myapp_functional_tests PRIVATE
    MYAPP_BINARY_PATH="${MYAPP_BINARY_PATH}"
    MYAPP_TEST_DIR="${CMAKE_CURRENT_SOURCE_DIR}"
)
```

`MYAPP_TEST_DIR` provides the path to the `test/` source directory, replacing any runtime `__file__`-style path discovery. Tests access both via the `TestEnvironment` singleton; see `cpp/testing.md`.

In test code:

```cpp
auto result = run(MYAPP_BINARY_PATH, {"create", "--version", "1", input_dir});
REQUIRE(result.returncode_ == 0);
```

## Two-Binary Test Pattern

Projects with both unit and functional tests define two separate binaries. They have different dependencies and must not be mixed:

```cmake
# Unit tests - links against application source files directly
add_executable(myapp_unit_tests
    unit/test_helpers.cpp
    unit/test_config.cpp
)
target_include_directories(myapp_unit_tests PRIVATE "${CMAKE_SOURCE_DIR}/src")
target_link_libraries(myapp_unit_tests PRIVATE Catch2::Catch2WithMain)

# Functional tests - spawns the binary as a subprocess, never links src/
add_executable(myapp_functional_tests
    subprocess_helper.cpp
    functional/test_create.cpp
    functional/test_list.cpp
)
target_include_directories(myapp_functional_tests PRIVATE
    ${CMAKE_CURRENT_SOURCE_DIR}
    "${CMAKE_SOURCE_DIR}/extern/subprocess.h"
)
target_link_libraries(myapp_functional_tests PRIVATE Catch2::Catch2WithMain)

include(CTest)
include(Catch)
catch_discover_tests(myapp_unit_tests)
catch_discover_tests(myapp_functional_tests)
```

The separation is intentional; unit tests link against source, functional tests do not. Mixing them produces a binary with unclear dependencies.

## Dependencies

All external dependencies are git submodules pinned to a specific commit, stored under `extern/`:

```
extern/
  ThirdPartyLib/
  ArgParser/
  Catch2/
  subprocess.h/
```

Always pin to a specific commit hash, never a branch name. Branch names move; commit hashes do not:

```bash
cd extern/Catch2 && git checkout v3.6.0
cd extern/subprocess.h && git checkout a3f3b8d  # specific commit hash
```

Every dependency must have an existence check in the root `CMakeLists.txt` before its `add_subdirectory` call. The error message must name the dependency, explain why it is needed, and tell the developer exactly how to fix it:

```cmake
if(NOT EXISTS "${CMAKE_SOURCE_DIR}/extern/ThirdPartyLib/CMakeLists.txt")
    message(FATAL_ERROR
"Missing dependency: ThirdPartyLib
myapp requires the ThirdPartyLib library.
It is provided as a submodule of this repository.
Did you forget to run the following commands?
   git submodule init
   git submodule update")
endif()

add_subdirectory(extern/ThirdPartyLib)
```

Single-header libraries check for the header file directly rather than a `CMakeLists.txt`:

```cmake
if(NOT EXISTS "${CMAKE_SOURCE_DIR}/extern/subprocess.h/subprocess.h")
    message(FATAL_ERROR
"Missing dependency: subprocess.h
Tests require the subprocess.h library.
It is provided as a submodule of this repository.
Did you forget to run the following commands?
   git submodule init
   git submodule update")
endif()
```

Never assume submodules are initialised. Always guard every dependency.

### Including extern/ headers

Use the `SYSTEM` keyword on every `target_include_directories` call that points into `extern/`. This marks those paths as system headers, so clang-tidy and the compiler suppress all warnings from third-party code by default:

```cmake
# Application source includes a third-party library header:
target_include_directories(myapp SYSTEM PRIVATE
    "${CMAKE_SOURCE_DIR}/extern/ThirdPartyLib/src"
)

# Project-owned headers (generated files, src/) use PRIVATE without SYSTEM:
target_include_directories(myapp PRIVATE
    "${CMAKE_BINARY_DIR}"
)
```

- `SYSTEM PRIVATE` tells CMake to pass `-isystem` instead of `-I` for those paths
- clang-tidy excludes system headers from all analysis by default; without `SYSTEM`, third-party headers generate thousands of suppressed warnings that inflate output and slow analysis
- Never combine `extern/` and project-owned paths in one `target_include_directories` call; they require different keywords

## Clang Tooling

Clang tools are pinned to version 18 across all projects for reproducibility. Never use the unversioned `clang-format` or `clang-tidy` binaries as the system default may differ between machines and CI runners.

### Installation

The Makefile must provide an `install_clang_tools` target:

```makefile
.PHONY: install_clang_tools
install_clang_tools: ## Install clang-format and clang-tidy at pinned version
	sudo apt-get install -y clang-format-18 clang-tidy-18
```

### Makefile targets

Use the versioned binaries explicitly in all targets:

```makefile
.PHONY: configure
configure: ## Configure the cmake build
	cmake -B build \
	  -DCMAKE_BUILD_TYPE=Debug \
	  -DCMAKE_EXPORT_COMPILE_COMMANDS=ON

.PHONY: build
build: ## Build the project
	cmake --build build

.PHONY: fmt
fmt: ## Format all source files with clang-format
	find src test -name "*.cpp" -o -name "*.h" | xargs clang-format-18 -i

.PHONY: fmt_check
fmt_check: ## Check formatting without modifying files
	find src test -name "*.cpp" -o -name "*.h" | xargs clang-format-18 --dry-run --Werror

.PHONY: lint_cpp
lint_cpp: ## Run clang-tidy static analysis (requires: make configure)
	clang-tidy-18 --quiet -p build \
	--header-filter="$(CURDIR)/src/.*" src/*.cpp 2>&1 \
	| grep -v " warnings generated"; \
	exit $${PIPESTATUS[0]}
```

- `--quiet` suppresses the "Suppressed N warnings" summary and hint lines
- `--header-filter="$(CURDIR)/src/.*"` limits diagnostic output to project source headers; extern/ headers are already excluded as system headers (see Including extern/ headers in this file) but this provides belt-and-suspenders coverage
- `grep -v " warnings generated"` strips the per-file progress counter, which counts all warnings before any filtering and is always misleading when third-party headers are present; `exit $${PIPESTATUS[0]}` preserves clang-tidy's exit code through the pipe
- `make configure` must be run before `make lint_cpp`; clang-tidy reads `build/compile_commands.json` to resolve include paths

Note: `fmt` and `fmt_check` include the `test/` directory; test code is subject to the same formatting standards as application code.

### Configuration files

Both `.clang-format` and `.clang-tidy` live at the project root. CMake is pointed at the build directory via `-p build` so clang-tidy can find `compile_commands.json`. The `FormatStyle: file` setting in `.clang-tidy` tells clang-tidy to use the root `.clang-format` for any formatting checks.
