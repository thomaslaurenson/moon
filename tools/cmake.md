# CMake Conventions

Conventions for CMake-based C++ projects.

## Design Principles

- CMake is the build system for all C++ projects — never use raw compiler invocations
- The Makefile is a task runner that wraps CMake — CI calls `make <target>`, never raw `cmake` commands
- One `build/` directory for everything — no separate lint or release build directories
- Dependencies are always git submodules pinned to a specific commit, never system-installed libraries

---

## Minimum Version

All projects must declare a minimum CMake version of 3.21:

```cmake
cmake_minimum_required(VERSION 3.21)
```

CMake 3.21 is the oldest version found on any supported build environment.
Never use `cmake_minimum_required(VERSION 3.10)` or other outdated minimums —
they unlock legacy behaviour that conflicts with modern CMake practices.

---

## C++ Standard

All projects must set a minimum C++ standard of 17. New projects should prefer 20:

```cmake
set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)
```

- `CMAKE_CXX_STANDARD_REQUIRED ON` — fails the build if the compiler does not
  support the requested standard, rather than silently falling back
- `CMAKE_CXX_EXTENSIONS OFF` — disables compiler-specific extensions such as
  GNU extensions, ensuring the code is portable standard C++

---

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

- `CMAKE_BUILD_TYPE` defaults to `Debug` — this ensures `compile_commands.json`
  is always generated with full debug information for clang-tidy
- `CMAKE_POSITION_INDEPENDENT_CODE ON` — required for shared libraries and
  good practice for all targets
- `CMAKE_EXPORT_COMPILE_COMMANDS ON` — generates `compile_commands.json` in
  the build directory, required for clang-tidy
- `CMAKE_RUNTIME_OUTPUT_DIRECTORY` — all executables land in `build/bin/`
  regardless of how many targets the project defines

---

## Build Directory

All projects use a single `build/` directory:

```bash
cmake -B build
cmake --build build
```

Executables are always output to `build/bin/`. For a project with multiple
binaries they all land in the same directory:

```
build/
  bin/
    mytool
    myothertool
  compile_commands.json
```

Never create separate build directories for lint, release, or test builds.
The default `Debug` build type produces a `compile_commands.json` that covers
all use cases.

---

## CMakeLists.txt Structure

Every directory that produces a target or manages a distinct concern has its
own `CMakeLists.txt`. The root never defines targets directly — it orchestrates.

```
CMakeLists.txt        # project settings, dependencies, add_subdirectory calls
src/
  CMakeLists.txt      # defines the main executable target(s)
test/
  CMakeLists.txt      # defines test targets
extern/
  Catch2/             # submodule — never modify
  StormLib/           # submodule — never modify
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

- Test executable targets
- Linking against `Catch2::Catch2WithMain`
- `catch_discover_tests`

---

## Testing Option

Every project must declare the `BUILD_TESTING` option in the root
`CMakeLists.txt`. This allows tests to be disabled for release builds or
when building on a system without test dependencies:

```cmake
option(BUILD_TESTING "Build tests" ON)

if(BUILD_TESTING)
    add_subdirectory(extern/Catch2)
    add_subdirectory(test)
endif()
```

---

## Dependencies

All external dependencies are git submodules pinned to a specific commit,
stored under `extern/`:

```
extern/
  StormLib/
  CLI11/
  Catch2/
```

Every dependency must have an existence check in the root `CMakeLists.txt`
before its `add_subdirectory` call. The error message must name the dependency,
explain why it is needed, and tell the developer exactly how to fix it:

```cmake
if(NOT EXISTS "${CMAKE_SOURCE_DIR}/extern/StormLib/CMakeLists.txt")
    message(FATAL_ERROR
"Missing dependency: StormLib
mpqcli requires the StormLib library.
It is provided as a submodule of this repository.
Did you forget to run the following commands?
   git submodule init
   git submodule update")
endif()

add_subdirectory(extern/StormLib)
```

Never assume submodules are initialised. Always guard every dependency.

---

## Clang Tooling

Clang tools are pinned to version 19 across all projects for reproducibility.
Never use the unversioned `clang-format` or `clang-tidy` binaries as the
system default may differ between machines and CI runners.

### Installation

The Makefile must provide an `install_clang_tools` target:

```makefile
.PHONY: install_clang_tools
install_clang_tools: ## Install clang-format and clang-tidy at pinned version
	sudo apt-get install -y clang-format-19 clang-tidy-19
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

.PHONY: format
format: ## Format all source files with clang-format
	find src -name "*.cpp" -o -name "*.h" | xargs clang-format-19 -i

.PHONY: format_check
format_check: ## Check formatting without modifying files
	find src -name "*.cpp" -o -name "*.h" | xargs clang-format-19 --dry-run --Werror

.PHONY: lint_cpp
lint_cpp: ## Run clang-tidy static analysis
	clang-tidy-19 -p build $(shell find src -name "*.cpp")
```

### Configuration files

Both `.clang-format` and `.clang-tidy` live at the project root. CMake is
pointed at the build directory via `-p build` so clang-tidy can find
`compile_commands.json`. The `FormatStyle: file` setting in `.clang-tidy`
tells clang-tidy to use the root `.clang-format` for any formatting checks.
