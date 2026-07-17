# CMake: application (CLI binary)

Application-specific CMake conventions. Assumes the universal CMake conventions.

An application ships a binary and has no public API: nothing outside the repository links its code, so there is no `include/` directory and no alias. If external consumers do need the core, it is a lib-cli; see cmake-lib-cli.

## Repository layout additions

```
src/                   # implementation, built as an internal core library; no main()
  CMakeLists.txt       # add_library
app/                   # the CLI: main() plus argument wiring only
  CMakeLists.txt       # add_executable, links the core
Dockerfile.glibc
Dockerfile.musl
.github/workflows/
  build.yml
  release.yml
  prerelease.yml
```

Two Dockerfiles are required: one for glibc builds, one for musl (static) builds; see the Docker fragment.

The root `CMakeLists.txt` orchestrates in order: `add_subdirectory(src)`, then `add_subdirectory(app)`, then `add_subdirectory(test)` when testing is on.

## Targets

An application defines two targets: an internal core library holding all the logic, and a thin executable that wires up the CLI and links it.

`src/CMakeLists.txt` builds the core. It is a normal `STATIC` library with no `include/` and no alias, because nothing outside this repository links it:

```cmake
add_library(myapp_core STATIC
    parser.cpp
    config.cpp
)

target_include_directories(myapp_core PUBLIC "${CMAKE_CURRENT_SOURCE_DIR}")

target_compile_features(myapp_core PUBLIC cxx_std_20)
```

`app/CMakeLists.txt` builds the binary:

```cmake
add_executable(myapp
    main.cpp
    options.cpp
)

target_link_libraries(myapp PRIVATE myapp_core)

target_include_directories(myapp SYSTEM PRIVATE
    "${PROJECT_SOURCE_DIR}/extern/CLI11/include"
)
```

The `PUBLIC` include on `myapp_core` is what lets `app/` and the test binaries include its headers by name; it is public to the targets in this project, which is as far as an internal library travels.

Splitting the core out of the executable is what makes the logic testable. A test binary cannot link an executable, so any code living beside `main()` can only be tested by recompiling its `.cpp` files into the test binary, which is a second build of the same source that drifts from the first. Compiling it once as a library and linking it everywhere removes that whole class of problem.

An app whose implementation is genuinely one `main.cpp` with nothing worth unit testing may skip `src/` and the core library entirely, and define the executable directly in `app/`. Add the split when there is logic to test, not before.

The binary lands in `build/bin/` via the universal `CMAKE_RUNTIME_OUTPUT_DIRECTORY` setting:

```
build/
  bin/
    myapp
    myapp_unit_tests
    myapp_functional_tests
  compile_commands.json
```

## Baking paths into test binaries

Functional tests need to know where the compiled binary lives at runtime. Rather than discovering it at runtime, bake the path in at CMake configure time using `target_compile_definitions`. This eliminates a whole class of path-resolution bugs and makes the test binary fully self-contained:

```cmake
# In test/CMakeLists.txt

if(WIN32)
    set(MYAPP_BINARY_PATH "${PROJECT_BINARY_DIR}/bin/myapp.exe")
else()
    set(MYAPP_BINARY_PATH "${PROJECT_BINARY_DIR}/bin/myapp")
endif()

# CI runs functional tests against a downloaded release binary rather than the
# one just built. MYAPP_BINARY_PATH_OVERRIDE lets the workflow point the tests at
# that artifact at configure time; locally it is unset and the build-tree path
# above is used.
if(MYAPP_BINARY_PATH_OVERRIDE)
    set(MYAPP_BINARY_PATH "${MYAPP_BINARY_PATH_OVERRIDE}")
endif()

target_compile_definitions(myapp_functional_tests PRIVATE
    MYAPP_BINARY_PATH="${MYAPP_BINARY_PATH}"
    MYAPP_TEST_DIR="${CMAKE_CURRENT_SOURCE_DIR}"
)
```

`MYAPP_TEST_DIR` provides the path to the `test/` source directory, replacing any runtime `__file__`-style path discovery. Tests access both via the `TestEnvironment` singleton; see the C++ testing-functional fragment.

In test code:

```cpp
auto result = run(MYAPP_BINARY_PATH, {"create", "--version", "1", input_dir});
REQUIRE(result.returncode_ == 0);
```

## Test targets

Unit and functional tests are separate binaries with different dependencies, and must not be mixed:

```cmake
# Unit tests - link the core library, never recompile its sources
add_executable(myapp_unit_tests
    unit/test_helpers.cpp
    unit/test_config.cpp
)
target_link_libraries(myapp_unit_tests PRIVATE myapp_core Catch2::Catch2WithMain)

# Functional tests - spawn the binary as a subprocess, and link neither it nor the core
add_executable(myapp_functional_tests
    subprocess_helper.cpp
    functional/test_create.cpp
    functional/test_list.cpp
)
# Project-owned test headers use PRIVATE without SYSTEM.
target_include_directories(myapp_functional_tests PRIVATE
    ${CMAKE_CURRENT_SOURCE_DIR}
)
# extern/subprocess.h is the submodule directory (the repo is literally named
# "subprocess.h"); mark it SYSTEM so clang-tidy and the compiler ignore it, and
# keep it in its own call - never combine SYSTEM and non-SYSTEM paths.
target_include_directories(myapp_functional_tests SYSTEM PRIVATE
    "${PROJECT_SOURCE_DIR}/extern/subprocess.h"
)
target_link_libraries(myapp_functional_tests PRIVATE Catch2::Catch2WithMain)

catch_discover_tests(myapp_unit_tests
    PROPERTIES LABELS "unit" SKIP_RETURN_CODE 4)
catch_discover_tests(myapp_functional_tests
    PROPERTIES LABELS "functional" SKIP_RETURN_CODE 4)
```

The separation is intentional. Unit tests link the core because they call it directly; functional tests link neither the core nor the binary, because they exercise the binary through its CLI as a user would. Mixing them produces a test binary with unclear dependencies and lets a functional test quietly call a function instead of the command.

`enable_testing()` and `include(Catch)` are called once in the root `CMakeLists.txt`, not here; see the universal fragment. `LABELS` is what lets `ctest -L unit` select a layer, and `SKIP_RETURN_CODE 4` is what stops a `SKIP()` being reported as a failure; both are explained in cpp/testing.md.

These two layers are what an application always has. It adds an integration layer if it has real data worth testing against below the CLI (see cpp/testing-integration.md), and a fuzz layer if the core parses untrusted input (see cpp/testing-fuzz.md).

Between the unit and functional layers, prefer the unit layer. It is faster, it fails with a stack trace instead of a diff of stdout, and a fixture can provoke a case that would take a contrived command line to reach. Reserve the functional layer for what only it can see: argv parsing, exit codes, and what lands on stdout.
