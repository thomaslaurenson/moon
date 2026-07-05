# CMake: Library

Library-specific CMake conventions. Assumes the universal CMake conventions.
Consumers use this library via git submodule and `add_subdirectory`, never
`find_package`, so there is no `install()`/`export()` CMake package config here.

## Repository layout additions

A library separates its public API from its implementation, which an application
doesn't need to:

```
include/<lib>/         # public headers - this is the API consumers see
src/                    # implementation (.cpp, and any private headers)
```

## Target

`src/CMakeLists.txt` defines the library with `add_library`, defaulting to `STATIC`
unless there's a specific reason to build shared:

```cmake
add_library(mylib STATIC
    src/parser.cpp
    src/archive.cpp
)

target_include_directories(mylib
    PUBLIC  "${CMAKE_SOURCE_DIR}/include"
    PRIVATE "${CMAKE_SOURCE_DIR}/src"
)

target_compile_features(mylib PUBLIC cxx_std_20)

add_library(mylib::mylib ALIAS mylib)
```

- `PUBLIC` on `include/` propagates that path to anything that links the library, so a consumer only needs `target_link_libraries(consumer PRIVATE mylib::mylib)`, not a separate include path of their own.
- `PRIVATE` on `src/` keeps implementation headers out of a consumer's include path entirely.
- The `mylib::mylib` `ALIAS` gives a consistent namespaced link name whether the library is added via `add_subdirectory` in this project or a monorepo superbuild; use the alias name in `target_link_libraries`, never the bare target name, so consumer code doesn't change if the linking mechanism ever does.
- Headers in `include/<lib>/` (not just `include/`) so a consumer's own includes read `#include <mylib/parser.h>`, not an unqualified `#include <parser.h>` that could collide with another dependency.

## Testing

A library has one test layer: unit tests linking the library target directly (see cpp/testing.md). There is no functional/subprocess layer, because there's no compiled binary to spawn; testing-functional.md does not apply here.

```cmake
add_executable(mylib_unit_tests
    unit/test_parser.cpp
    unit/test_archive.cpp
)
target_link_libraries(mylib_unit_tests PRIVATE mylib::mylib Catch2::Catch2WithMain)

include(CTest)
include(Catch)
catch_discover_tests(mylib_unit_tests)
```
