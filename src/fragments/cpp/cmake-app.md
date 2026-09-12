# CMake: application (CLI binary)

Application-specific CMake conventions. Assumes the universal CMake conventions.

An application ships a binary and has no public API: nothing outside the repository links its code, so there is no `include/` directory. If external consumers do need the core, it is a lib-cli; see cmake-lib-cli.

## Layout

The directories an application adds, `app/`, `completion/` and the release files, are in the CLI scaffolding fragment; this fragment owns the targets that build them. The root `CMakeLists.txt` orchestrates in order: `add_subdirectory(src)`, then `add_subdirectory(app)`, then `add_subdirectory(test)` when testing is on.

## Targets

An application defines two targets: an internal core library holding all the logic, and a thin executable that wires up the CLI and links it. The executable takes the bare project name, `myproj`, and the core is `myproj_core`, linked as `myproj::core`; see Target names in the universal fragment.

`src/CMakeLists.txt` builds the core. It is a normal `STATIC` library with no `include/`, because nothing outside this repository links it:

```cmake
add_library(myproj_core STATIC
    parser.cpp
    config.cpp
)

target_include_directories(myproj_core PUBLIC "${CMAKE_CURRENT_SOURCE_DIR}")

target_link_libraries(myproj_core PRIVATE myproj::warnings)

target_compile_features(myproj_core PUBLIC cxx_std_20)

add_library(myproj::core ALIAS myproj_core)
```

The alias is for the configure-time typo check the universal fragment describes, not for consumers: nothing outside the repository links the core, and an alias changes nothing about that. `myproj::warnings` is the warning bar from the same fragment.

`app/CMakeLists.txt` builds the binary:

```cmake
add_executable(myproj
    main.cpp
    options.cpp
)

target_link_libraries(myproj PRIVATE myproj::core myproj::warnings)

target_include_directories(myproj SYSTEM PRIVATE
    "${PROJECT_SOURCE_DIR}/extern/CLI11/include"
)
```

The `PUBLIC` include on `myproj_core` is what lets `app/` and the test binaries include its headers by name, `#include "config.h"`, with no `include/` tree and no `<myproj/...>` prefix; it is public to the targets in this project, which is as far as an internal library travels.

Splitting the core out of the executable is what makes the logic testable. A test binary cannot link an executable, so any code living beside `main()` can only be tested by recompiling its `.cpp` files into the test binary, which is a second build of the same source that drifts from the first. Compiling it once as a library and linking it everywhere removes that whole class of problem.

There is no one-file exemption. The smallest application still has `src/` and a core: the error handling, output and testing fragments all assume that boundary, and a tool that starts as one file rarely stays one. A `main.cpp` with nothing worth testing yet costs one `add_library` to keep the shape.

The binary lands in the build configuration's `bin/` (for example `build/dev/bin/`) via the universal `CMAKE_RUNTIME_OUTPUT_DIRECTORY` setting:

```text
build/
  dev/
    bin/
      myproj
      myproj_unit_tests
      myproj_functional_tests
    compile_commands.json
```

## Generated version header

The version comes from `project(myproj VERSION 1.2.3)` in the root (see the C++ style fragment). Generate it in `src/CMakeLists.txt`, beside the core library, so the core and the CLI read the same constant:

```cmake
configure_file(
    "${PROJECT_SOURCE_DIR}/cmake/version.h.in"
    "${PROJECT_BINARY_DIR}/include/myproj/version.h"
    @ONLY
)

target_include_directories(myproj_core PUBLIC "${PROJECT_BINARY_DIR}/include")
```

An application has no public API, so nothing outside the repository reads this header. Generate it into an include tree anyway rather than putting `"${PROJECT_BINARY_DIR}"` itself on the include path, which would expose every generated file in the build tree to `#include`. `app/` picks the header up through the core's `PUBLIC` include directory, so `--version` prints the same value the library reports and there is one place to change it.

Keep the template in `cmake/`, never in `src/`: it is a build input, not something the compiler ever sees.

## Baking paths into test binaries

Functional tests need to know where the compiled binary lives at runtime. Rather than discovering it at runtime, bake the path in at CMake configure time using `target_compile_definitions`. This eliminates a whole class of path-resolution bugs and makes the test binary fully self-contained:

```cmake
# In test/CMakeLists.txt

if(WIN32)
    set(MYPROJ_BINARY_PATH "${PROJECT_BINARY_DIR}/bin/myproj.exe")
else()
    set(MYPROJ_BINARY_PATH "${PROJECT_BINARY_DIR}/bin/myproj")
endif()

# MYPROJ_BINARY_PATH_OVERRIDE points the functional tests at a binary other than
# the one just built: a downloaded release asset, or an installed copy, to check
# a published artefact behaves. Unset, which is the normal case including in CI,
# the build-tree path above is used.
if(MYPROJ_BINARY_PATH_OVERRIDE)
    set(MYPROJ_BINARY_PATH "${MYPROJ_BINARY_PATH_OVERRIDE}")
endif()

target_compile_definitions(myproj_functional_tests PRIVATE
    MYPROJ_BINARY_PATH="${MYPROJ_BINARY_PATH}"
    MYPROJ_TEST_DIR="${CMAKE_CURRENT_SOURCE_DIR}"
)
```

`MYPROJ_TEST_DIR` provides the path to the `test/` source directory, replacing any runtime path discovery relative to the test binary. Tests access both via the `TestEnvironment` singleton; see the C++ testing-functional fragment.

In test code:

```cpp
auto result = Run(MYPROJ_BINARY_PATH, {"create", input_dir, "out.dat"});
REQUIRE(result.returncode == 0);
```

## Test targets

Unit and functional tests are separate binaries with different dependencies, and must not be mixed:

```cmake
# Unit tests - link the core library, never recompile its sources
add_executable(myproj_unit_tests
    unit/test_helpers.cpp
    unit/test_config.cpp
)
target_include_directories(myproj_unit_tests PRIVATE
    "${CMAKE_CURRENT_SOURCE_DIR}/fixtures"
)
target_link_libraries(myproj_unit_tests PRIVATE
    myproj::core
    myproj::warnings
    Catch2::Catch2WithMain
)

# Functional tests - spawn the binary as a subprocess, and link neither it nor the core
add_executable(myproj_functional_tests
    subprocess_helper.cpp
    functional/test_create.cpp
    functional/test_list.cpp
)
# Project-owned test headers use PRIVATE without SYSTEM: the helper beside this
# file, and the fixtures every layer shares.
target_include_directories(myproj_functional_tests PRIVATE
    "${CMAKE_CURRENT_SOURCE_DIR}"
    "${CMAKE_CURRENT_SOURCE_DIR}/fixtures"
)
# extern/subprocess.h is the submodule directory (the repo is literally named
# "subprocess.h"); mark it SYSTEM so clang-tidy and the compiler ignore it, and
# keep it in its own call - never combine SYSTEM and non-SYSTEM paths.
target_include_directories(myproj_functional_tests SYSTEM PRIVATE
    "${PROJECT_SOURCE_DIR}/extern/subprocess.h"
)
target_link_libraries(myproj_functional_tests PRIVATE
    myproj::warnings
    Catch2::Catch2WithMain
)

catch_discover_tests(myproj_unit_tests
    PROPERTIES LABELS "unit" SKIP_RETURN_CODE 4)
catch_discover_tests(myproj_functional_tests
    PROPERTIES LABELS "functional" SKIP_RETURN_CODE 4)
```

The separation is intentional. Unit tests link the core because they call it directly; functional tests link neither the core nor the binary, because they exercise the binary through its CLI as a user would. Mixing them produces a test binary with unclear dependencies and lets a functional test quietly call a function instead of the command. The unit binary sees every header in `src/`, private ones included, through the core's `PUBLIC` include directory; the functional binary sees only `test/`.

`enable_testing()` and `include(Catch)` are called once in the root `CMakeLists.txt`, not here; see the universal fragment. `LABELS` is what lets `ctest -L unit` select a layer, and `SKIP_RETURN_CODE 4` is what stops a `SKIP()` being reported as a failure; both are explained in cpp/testing.md.

These two layers are what an application always has. It adds an integration layer if it has real data worth testing against below the CLI (see cpp/testing-integration.md), and a fuzz layer if the core parses untrusted input (see cpp/testing-fuzz.md).

Between the unit and functional layers, prefer the unit layer. It is faster, it fails with a stack trace instead of a diff of stdout, and a fixture can provoke a case that would take a contrived command line to reach. Reserve the functional layer for what only it can see: argv parsing, exit codes, and what lands on stdout.
