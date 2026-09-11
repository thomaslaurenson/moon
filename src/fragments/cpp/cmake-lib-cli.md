# CMake: library with a bundled CLI

CMake conventions for a project that is a reusable library first and ships a thin CLI binary on top of it. Assumes the universal CMake conventions.

This tier is both of its neighbours at once: a public API behind `include/` like a library, and a shipped binary with functional tests and a release matrix like an application. It is the single tier fragment for a lib-cli project and states the combined shape in full, so nothing here defers to cmake-lib or cmake-app.

How it differs from those neighbours: a plain library has no compiled binary to ship or spawn, so it needs no `app/`, no Docker release and no functional test layer. A plain application has a binary but no reusable core behind a public API, so its `src/` builds an internal core library with no `include/` and no alias. A lib-cli has both halves. All real logic lives in the library so it stays unit-testable and reusable by other projects; the executable is a thin wrapper that parses arguments and calls into the library.

## Repository layout additions

```
include/myproj/         # public headers - the API the library exposes
src/                   # library implementation (.cpp and private headers); no main()
  CMakeLists.txt       # add_library
app/                   # the CLI: main() plus argument wiring only
  CMakeLists.txt       # add_executable, links the library
Dockerfile             # static musl build into scratch; see the C++ Docker fragment
.dockerignore
.gpipe.yml             # installer/checksum config; see workflows-app.md
```

`src/` never contains a `main()`; keeping the entry point in `app/` stops it being compiled into the library and keeps the library free of CLI concerns. The release Dockerfile ships the `app/` binary exactly as for an application; see the C++ Docker fragment.

The root `CMakeLists.txt` orchestrates in order: `add_subdirectory(src)`, then `add_subdirectory(app)`, then `add_subdirectory(test)` when testing is on.

## Library target

`src/CMakeLists.txt` defines the reusable core with `add_library`, defaulting to `STATIC` unless there is a specific reason to build shared:

```cmake
add_library(myproj_lib STATIC
    parser.cpp
    archive.cpp
)

target_include_directories(myproj_lib
    PUBLIC  "${PROJECT_SOURCE_DIR}/include"
    PRIVATE "${CMAKE_CURRENT_SOURCE_DIR}"
)

target_link_libraries(myproj_lib PRIVATE myproj::warnings)

target_compile_features(myproj_lib PUBLIC cxx_std_20)

add_library(myproj::myproj ALIAS myproj_lib)
```

- The target is `myproj_lib` and the executable in `app/` is `myproj`. The binary a user runs has to carry the bare name, target names are global to the build, and the library's raw name is one nothing outside this file ever writes, since every link goes through the `myproj::myproj` alias. This is the general rule in Target names in the universal fragment: the bare name goes to what a user reaches for, and everything else carries the prefix.
- Source files are named relative to the directory holding the `CMakeLists.txt`. This file lives in `src/`, so the entry is `parser.cpp`, never `src/parser.cpp`, which would resolve to `src/src/parser.cpp` and fail to configure.
- `PUBLIC` on `include/` propagates that path to everything that links the library, so neither the CLI nor an outside consumer needs an include path of its own; `target_link_libraries(myproj PRIVATE myproj::myproj)` is the whole wiring.
- `PRIVATE` on the current directory keeps implementation headers off the consumer's include path entirely. The split is the API boundary expressed in CMake: what is in `include/` is promised, what is in `src/` can change freely.
- `PROJECT_SOURCE_DIR`, never `CMAKE_SOURCE_DIR`. A consumer pulls this library in as a submodule and calls `add_subdirectory`, at which point `CMAKE_SOURCE_DIR` is *their* root and the public include path would silently point at their `include/`.
- The `myproj::myproj` `ALIAS` gives a consistent namespaced link name whether the library is added by this project or by a consumer's superbuild. Use the alias in every `target_link_libraries`, never the bare target name, so nothing changes if the linking mechanism does.
- Headers live in `include/myproj/`, not directly in `include/`, so includes read `#include <myproj/parser.h>` and cannot collide with another dependency's `parser.h`.

## Generated version header

The version comes from `project(myproj VERSION 1.2.3)` in the root (see the C++ style fragment). Generate it into the public include tree, so the library, the CLI, and an outside consumer all read the same compile-time constant through the same include:

```cmake
configure_file(
    "${PROJECT_SOURCE_DIR}/cmake/version.h.in"
    "${PROJECT_BINARY_DIR}/include/myproj/version.h"
    @ONLY
)

target_include_directories(myproj_lib PUBLIC "${PROJECT_BINARY_DIR}/include")
```

`#include <myproj/version.h>` then matches every other public header. Putting `PROJECT_BINARY_DIR` itself on the public path instead would hand consumers every generated file in the build tree.

## CLI target

`app/CMakeLists.txt` defines the executable. It links the library through its alias and pulls any CLI-only dependency (for example CLI11) from `extern/` as a SYSTEM include:

```cmake
add_executable(myproj
    main.cpp
    options.cpp
)

target_link_libraries(myproj PRIVATE myproj::myproj myproj::warnings)

target_include_directories(myproj SYSTEM PRIVATE
    "${PROJECT_SOURCE_DIR}/extern/CLI11/include"
)
```

The executable stays thin: it parses arguments, calls library functions, and turns the library's exceptions into exit codes and messages on `stderr`. Anything worth unit testing belongs in the library. A CLI dependency is linked here and never by the library, so a consumer of the core does not inherit an argument parser they will not use. The binary lands in the build configuration's `bin/` (for example `build/dev/bin/`) via the universal `CMAKE_RUNTIME_OUTPUT_DIRECTORY` setting.

## Testing

A lib-cli is the one tier that can have all four layers. Unit, integration and fuzz link the library; functional spawns the binary:

- Unit tests link `myproj::myproj` and test logic, with fixtures supplying whatever input they need (see cpp/testing.md).
- Integration tests link `myproj::myproj` and run against real data the machine must already have. Opt-in (see cpp/testing-integration.md).
- Functional tests spawn the compiled `myproj` and verify its CLI behaviour end-to-end (see cpp/testing-functional.md). This layer applies because a lib-cli ships a binary, unlike a plain library.
- Fuzz harnesses link `myproj::myproj` and drive its parsers with hostile input. Built on demand (see cpp/testing-fuzz.md).

```cmake
# test/CMakeLists.txt

add_executable(myproj_unit_tests
    unit/test_parser.cpp
    unit/test_archive.cpp
)
target_include_directories(myproj_unit_tests PRIVATE
    "${CMAKE_CURRENT_SOURCE_DIR}/fixtures"
)
target_link_libraries(myproj_unit_tests PRIVATE
    myproj::myproj
    myproj::warnings
    Catch2::Catch2WithMain
)

add_executable(myproj_functional_tests
    subprocess_helper.cpp
    functional/test_create.cpp
)
# Project-owned test headers use PRIVATE without SYSTEM.
target_include_directories(myproj_functional_tests PRIVATE
    ${CMAKE_CURRENT_SOURCE_DIR}
)
# extern/subprocess.h is the submodule directory; mark it SYSTEM and keep it in its
# own call - never combine SYSTEM and non-SYSTEM paths.
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

The unit binary links the library rather than listing `src/*.cpp` again: the library already compiles that source once, and linking it keeps the two builds from drifting. The functional binary links neither the library nor the executable, because it exercises the binary through its CLI as a user would; giving it the library would let a functional test quietly call a function instead of running the command.

The split between those two is the tier's main testing question, and it has a default answer: a lib-cli puts its logic in the library, so almost everything is unit-testable without a subprocess. Test the library through the library, and keep the functional layer for what only it can see: argv parsing, exit codes, and stdout.

`enable_testing()` and `include(Catch)` are called once in the root `CMakeLists.txt`, not here; see the universal fragment. `LABELS` is what lets `ctest -L unit` select a layer, and `SKIP_RETURN_CODE 4` is what stops a `SKIP()` being reported as a failure; both are explained in cpp/testing.md.

## Baking paths into test binaries

Functional tests need the path to the compiled `myproj`, and both non-unit layers need the path to their own source directory for checked-in data. Bake both in at configure time rather than discovering them at runtime:

```cmake
# In test/CMakeLists.txt

if(WIN32)
    set(MYPROJ_BINARY_PATH "${PROJECT_BINARY_DIR}/bin/myproj.exe")
else()
    set(MYPROJ_BINARY_PATH "${PROJECT_BINARY_DIR}/bin/myproj")
endif()

# MYPROJ_BINARY_PATH_OVERRIDE points the functional tests at a binary other than
# the one just built: a downloaded release asset, or an installed copy, to check
# a published artifact behaves. Unset, which is the normal case including in CI,
# the build-tree path above is used.
if(MYPROJ_BINARY_PATH_OVERRIDE)
    set(MYPROJ_BINARY_PATH "${MYPROJ_BINARY_PATH_OVERRIDE}")
endif()

target_compile_definitions(myproj_functional_tests PRIVATE
    MYPROJ_BINARY_PATH="${MYPROJ_BINARY_PATH}"
    MYPROJ_TEST_DIR="${CMAKE_CURRENT_SOURCE_DIR}"
)

target_compile_definitions(myproj_unit_tests PRIVATE
    MYPROJ_TEST_DIR="${CMAKE_CURRENT_SOURCE_DIR}"
)
```

This eliminates a whole class of path-resolution bugs and makes each test binary self-contained. Tests reach the baked paths through the `TestEnvironment` singleton; see cpp/testing-functional.md.

## Release

A lib-cli ships its CLI binary, so it uses the application release path in full: a static musl Dockerfile, native macOS and Windows build jobs, the build/test/release workflow set, and the released-binary badges. See the C++ Docker fragment and workflows-app. The library half is not separately packaged for `find_package`; a consumer who wants the core links it via git submodule and `add_subdirectory`, the same as for a plain library.
