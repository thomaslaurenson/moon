# CMake: library

Library-specific CMake conventions. Assumes the universal CMake conventions. Consumers use this library via git submodule and `add_subdirectory`, never `find_package`, so there is no `install()`/`export()` CMake package config here.

A library has no `main()` and ships no binary. If the project also ships a CLI, it is a lib-cli; see cmake-lib-cli.

## Layout

The two trees, `include/myproj/` mirrored by `src/`, and the rules that keep them apart are in the library scaffolding fragment; this fragment owns the targets that build them. The root `CMakeLists.txt` orchestrates: `add_subdirectory(src)`, then `add_subdirectory(test)` when testing is on, then `add_subdirectory(examples/<name>)` when examples are on. It defines no project targets of its own: `src/CMakeLists.txt` adds the modules and defines the aggregate.

## Module targets

Each directory under `src/` builds one library target, defaulting to `STATIC`. The target is named `myproj_<module>` and aliased into the project's namespace:

```cmake
# src/archive/CMakeLists.txt

add_library(myproj_archive STATIC
    archive.cpp
    crypto.cpp
)

target_include_directories(myproj_archive
    PUBLIC  "${PROJECT_SOURCE_DIR}/include"
            "${PROJECT_BINARY_DIR}/include"
    PRIVATE "${PROJECT_SOURCE_DIR}/src"
)

target_link_libraries(myproj_archive
    PUBLIC  zlibstatic myproj::common
    PRIVATE myproj::warnings
)

target_compile_features(myproj_archive PUBLIC cxx_std_20)

add_library(myproj::archive ALIAS myproj_archive)
```

- Source files are named relative to the directory holding the `CMakeLists.txt`. This file lives in `src/archive/`, so the entry is `archive.cpp`, never `src/archive/archive.cpp`, which would resolve to `src/archive/src/archive/archive.cpp` and fail to configure.
- `PUBLIC` on `include/` propagates that path to anything linking the module, so neither a sibling module nor an outside consumer needs an include path of its own.
- `"${PROJECT_BINARY_DIR}/include"` is the generated header tree, which holds `version.h`. Every module carries it, and carries it `PUBLIC`, so a module's own sources and an outside consumer reach the version constant the same way. See Generated version header.
- `PRIVATE` on `src/` lets modules include each other's private headers while keeping them off a consumer's include path entirely.
- `PRIVATE myproj::warnings` is the warning bar, and every module carries this line. It is never a per-module flag list: see Warnings in the universal fragment for why the bar is one target rather than a copy per module.
- `PROJECT_SOURCE_DIR`, never `CMAKE_SOURCE_DIR`: under the `add_subdirectory` consumption model this library is built for, the latter resolves to the consumer's root. See the universal fragment.
- The `myproj_` prefix is not decoration. Target names are global to the whole CMake build, and a module called `crypto`, `common` or `config` will collide the first time this library and another land in the same superbuild. See the universal fragment.
- A module never calls `find_package`. The root resolves external dependencies and the module consumes the resulting imported targets.
- Link sibling modules through the alias (`myproj::common`), never the bare target. CMake requires a name containing `::` to be a target that already exists, so a typo fails at configure time with "target not found". A bare name is assumed to be a system library instead, and surfaces much later as an obscure `-lmyproj_comon` linker error. The cost is that the module `add_subdirectory` calls in `src/CMakeLists.txt` must run in dependency order, since an alias has to exist before it is linked.
- A module is never its own standalone CMake project. Do not wrap it in `if(NOT DEFINED MYPROJ_ROOT_BUILD) project(...)`: it duplicates the root's compiler and standard setup once per module, and those copies drift. A consumer wanting one module gets it through `EXCLUDE_FROM_ALL`; see the aggregate section.

A library small enough to have no modules defines one target in `src/CMakeLists.txt` the same way, named plainly after the project. Modules are the shape a library grows into, not a ceremony to start with; the aggregate below is worth adding as soon as there are two.

## Aggregate target

`src/CMakeLists.txt` defines an `INTERFACE` target that links every module and carries the public include path. This is what a consumer links, and it is the only name they should need to know. It lives beside the module `add_subdirectory` calls, not in the root, so the root defines no targets of its own:

```cmake
# src/CMakeLists.txt, after the module add_subdirectory calls

add_library(myproj INTERFACE)
target_link_libraries(myproj INTERFACE
    myproj::common
    myproj::archive
    myproj::record
)
target_include_directories(myproj INTERFACE "${PROJECT_SOURCE_DIR}/include")
add_library(myproj::myproj ALIAS myproj)
```

A consumer then writes `target_link_libraries(theirs PRIVATE myproj::myproj)` and gets every module, the public headers, and any transitive dependency. Without the aggregate they have to know the module map and list `myproj::archive myproj::record myproj::common` themselves, which turns every internal reorganisation into a breaking change.

Use an alias in every `target_link_libraries`, never the bare target name. This holds inside the project as well as outside it, for the configure-time error described above.

### Consuming a single module

The aggregate is the default, but a consumer needing one module links its alias directly and adds the library with `EXCLUDE_FROM_ALL`:

```cmake
# In the consumer's CMakeLists.txt
add_subdirectory(extern/myproj EXCLUDE_FROM_ALL)

target_link_libraries(theirapp PRIVATE myproj::record)
```

`EXCLUDE_FROM_ALL` is what makes this worth doing. Without it every module is added to the default build target and compiles whether or not anything links it. With it, CMake builds only `myproj_record` and whatever it depends on: the other modules are configured but never compiled.

This is the whole reason a library is split into module targets rather than built as one archive: a consumer pays for the parts they use. It also means module dependencies should be kept honest, since linking one module drags in everything it declares.

## Examples

A library may ship example programs demonstrating its API. They live in `examples/`, one directory per program, behind an option that defaults to `OFF`:

```cmake
option(MYPROJ_BUILD_EXAMPLES "Build example programs" OFF)

if(MYPROJ_BUILD_EXAMPLES)
    add_subdirectory(examples/auth)
endif()
```

```cmake
# examples/auth/CMakeLists.txt
add_executable(myproj_example_auth main.cpp)
target_link_libraries(myproj_example_auth PRIVATE myproj::myproj myproj::warnings)

# The target name is namespaced; the binary a reader runs is not.
set_target_properties(myproj_example_auth PROPERTIES OUTPUT_NAME "myproj-auth")
```

The directory is named for what it demonstrates (`examples/auth/`), and the target is `myproj_example_<name>`. The `_example_` infix is not decoration: `myproj_<name>` is exactly the module target pattern, so an example demonstrating the `auth` module would otherwise declare a second target called `myproj_auth` and fail the configure with a duplicate-target error. Give the executable an `OUTPUT_NAME` so the file on disk stays readable.

An example is not an application, and a library with one is still a library. The distinction is what the project ships:

- `examples/` is a demonstration. It is built on request, has no tests, no Docker image, no release artefact and no badge. Its job is to prove the public API is usable and to give a consumer something to copy. It links the aggregate through its alias, exactly as a consumer would, which is what makes it an honest demonstration rather than a program with special access.
- `app/` is a shipped binary, and having one makes the project a lib-cli, not a library. That tier adds functional tests, the Docker release matrix and released-binary badges. See cmake-lib-cli.

Default `OFF` because an example is dead weight in a consumer's build. Keep examples compiling: an example that no longer builds is a worse advertisement than no example at all. CI turns the option on in the existing test job, so they are compiled and nothing runs them; see workflows-lib.md.

## Generated version header

The version comes from `project(myproj VERSION 1.2.3)` in the root (see the C++ style fragment). Generate it into the public include tree so consumers can query it, rather than putting the whole binary directory on their include path. The `configure_file` call goes at the top of `src/CMakeLists.txt`, above the module `add_subdirectory` calls, so a module's own sources can include the header it writes:

```cmake
# src/CMakeLists.txt, before the module add_subdirectory calls

configure_file(
    "${PROJECT_SOURCE_DIR}/cmake/version.h.in"
    "${PROJECT_BINARY_DIR}/include/myproj/version.h"
    @ONLY
)
```

The generated directory is then carried by the module targets (see Module targets), and never added to the aggregate. The aggregate is an `INTERFACE` library, and `target_include_directories(myproj PUBLIC ...)` on one of those is not a style preference but a configure error: "target_include_directories may only set INTERFACE properties on INTERFACE targets". Writing `INTERFACE` there instead configures cleanly and is still wrong, because only code linking the aggregate would see the header: a module wanting its own version constant would fail to compile, and the aggregate is declared after the modules that would need it. A `PUBLIC` path on each module reaches both, since the aggregate links the modules and inherits their interface.

A consumer then writes `#include <myproj/version.h>`, matching every other public header. Naming `"${PROJECT_BINARY_DIR}"` itself instead would put every generated file in the build tree on their include path.

A library small enough to have no modules has no aggregate either: its single `add_library` in `src/CMakeLists.txt` carries both include directories as `PUBLIC` directly, which is the same arrangement with one target instead of several.

## Testing

A library has a unit layer always, plus an integration layer where it has real data to test against and a fuzz layer where it parses untrusted input. All of them link the library. There is no functional layer: that one spawns a compiled binary, and a library has none.

The unit layer therefore carries the whole load, and it is meant to. A unit test may build synthetic archives, write temp files and drive loopback sockets; what it may not do is depend on data the machine must already have. See cpp/testing.md.

```cmake
# test/CMakeLists.txt

add_executable(myproj_unit_tests
    unit/archive/test_archive.cpp
    unit/archive/test_crypto.cpp
    unit/record/test_reader.cpp
)

target_include_directories(myproj_unit_tests PRIVATE
    "${CMAKE_CURRENT_SOURCE_DIR}/fixtures"
    "${PROJECT_SOURCE_DIR}/src"
)

target_link_libraries(myproj_unit_tests PRIVATE
    myproj::myproj
    myproj::warnings
    Catch2::Catch2WithMain
)

target_compile_definitions(myproj_unit_tests PRIVATE
    MYPROJ_TEST_DIR="${CMAKE_CURRENT_SOURCE_DIR}"
)

catch_discover_tests(myproj_unit_tests
    PROPERTIES LABELS "unit" SKIP_RETURN_CODE 4)
```

Link the aggregate `myproj::myproj` unless a test binary genuinely covers one module, in which case link that module's alias and keep the binary small.

`enable_testing()` and `include(Catch)` are called once in the root `CMakeLists.txt`, not here; see the universal fragment. `LABELS` is what lets `ctest -L unit` select a layer, and `SKIP_RETURN_CODE 4` is what stops a `SKIP()` being reported as a failure; both are explained in cpp/testing.md.

`MYPROJ_TEST_DIR` gives tests the path to their own source directory, so they can find checked-in data under `test/data/` without runtime path discovery.

`src/` is on the unit binary's include path, on a fuzz harness's for the same reason, and on no other test binary's. The layer tests logic, and some of it is deliberately not API: a private helper in `src/` gets a unit test like anything else, and the test includes its header by the same path the implementation does. The functional binary never sees `src/`, because it exercises the binary as a user would, and a consumer never does, because the library keeps that directory `PRIVATE`.

The integration and fuzz targets are guarded by their own options and defined alongside this one; see cpp/testing-integration.md and cpp/testing-fuzz.md.
