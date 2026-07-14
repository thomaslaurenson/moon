# CMake: application (CLI binary)

Application-specific CMake conventions. Assumes the universal CMake conventions.

## Repository layout additions

An application adds release packaging on top of the universal layout:

```
Dockerfile.glibc
Dockerfile.musl
.github/workflows/
  build.yml
  release.yml
  prerelease.yml
```

Two Dockerfiles are required: one for glibc builds, one for musl (static) builds; see the Docker fragment.

## Target

`src/CMakeLists.txt` defines the binary with `add_executable`. All application source lives directly in `src/`; no header-only layout unless a given file's implementation is 50 lines or fewer.

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
    set(MYAPP_BINARY_PATH "${CMAKE_BINARY_DIR}/bin/myapp.exe")
else()
    set(MYAPP_BINARY_PATH "${CMAKE_BINARY_DIR}/bin/myapp")
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

## Two-Binary test pattern

An application with both unit and functional tests defines two separate binaries. They have different dependencies and must not be mixed:

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
# Project-owned test headers use PRIVATE without SYSTEM.
target_include_directories(myapp_functional_tests PRIVATE
    ${CMAKE_CURRENT_SOURCE_DIR}
)
# extern/subprocess.h is the submodule directory (the repo is literally named
# "subprocess.h"); mark it SYSTEM so clang-tidy and the compiler ignore it, and
# keep it in its own call - never combine SYSTEM and non-SYSTEM paths.
target_include_directories(myapp_functional_tests SYSTEM PRIVATE
    "${CMAKE_SOURCE_DIR}/extern/subprocess.h"
)
target_link_libraries(myapp_functional_tests PRIVATE Catch2::Catch2WithMain)

include(CTest)
include(Catch)
catch_discover_tests(myapp_unit_tests)
catch_discover_tests(myapp_functional_tests)
```

The separation is intentional; unit tests link against source, functional tests do not. Mixing them produces a binary with unclear dependencies.
