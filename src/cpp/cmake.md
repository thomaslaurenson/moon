# CMake Conventions

Conventions for CMake-based C++ projects. Universal to both applications and
libraries; tier-specific target definitions (`add_executable` vs `add_library`,
and the test layers each implies) live in the cmake-app or cmake-lib fragment.

## Design Principles

- CMake is the build system for all C++ projects; never use raw compiler invocations
- The Makefile is a task runner that wraps CMake; CI calls `make <target>`, never raw `cmake` commands
- One `build/` directory for everything; no separate lint or release build directories
- Dependencies are always git submodules pinned to a specific commit, never system-installed libraries

## Repository Layout

Every C++ project contains at least these at the root; a project-tier fragment
(cmake-app or cmake-lib) adds the rest (release Dockerfiles, `include/`, and so on):

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
src/                  # implementation; see the tier fragment for what else it contains
extern/               # git submodules only; never copy third-party headers manually
test/                 # see cpp/testing.md for internal structure
```

- `extern/` contains only git submodules; never manually copied headers or installed libraries

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
- `CMAKE_RUNTIME_OUTPUT_DIRECTORY`: all executables (the app binary, or a library's test binaries) land in `build/bin/` regardless of how many targets the project defines

## Build Directory

All projects use a single `build/` directory:

```bash
cmake -B build
cmake --build build
```

Never create separate build directories for lint, release, or test builds. The default `Debug` build type produces a `compile_commands.json` that covers all use cases.

## CMakeLists.txt Structure

Every directory that produces a target or manages a distinct concern has its own `CMakeLists.txt`. The root never defines targets directly; it orchestrates.

```
CMakeLists.txt        # project settings, dependencies, add_subdirectory calls
src/
  CMakeLists.txt      # defines the main target(s); see the tier fragment
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
- `add_subdirectory(src)`
- `add_subdirectory(test)` when `BUILD_TESTING` is on

### src/CMakeLists.txt responsibilities

- `add_executable` or `add_library`; see the tier fragment for which and how
- `target_include_directories`
- `target_link_libraries`
- `configure_file` for generated headers

## Testing Option

Every project must declare the `BUILD_TESTING` option in the root `CMakeLists.txt`. This allows tests to be disabled when building on a system without test dependencies, or (for a library) when a consumer adds it as a subdirectory and doesn't want its tests built too:

```cmake
option(BUILD_TESTING "Build tests" ON)

if(BUILD_TESTING)
    add_subdirectory(extern/Catch2)
    add_subdirectory(test)
endif()
```

## Dependencies

All external dependencies are git submodules pinned to a specific commit, stored under `extern/`:

```
extern/
  ThirdPartyLib/
  Catch2/
```

Always pin to a specific commit hash, never a branch name. Branch names move; commit hashes do not:

```bash
cd extern/Catch2 && git checkout v3.6.0
```

Every dependency must have an existence check in the root `CMakeLists.txt` before its `add_subdirectory` call. The error message must name the dependency, explain why it is needed, and tell the developer exactly how to fix it:

```cmake
if(NOT EXISTS "${CMAKE_SOURCE_DIR}/extern/ThirdPartyLib/CMakeLists.txt")
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
    "${CMAKE_SOURCE_DIR}/extern/ThirdPartyLib/src"
)

# Project-owned headers (generated files) use PRIVATE without SYSTEM:
target_include_directories(mytarget PRIVATE
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
