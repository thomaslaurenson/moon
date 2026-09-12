# C++ library scaffolding

How a library is laid out and what a consumer sees of it. Assumes the C++ style and Doxygen fragments, and the library or library-with-CLI tier fragment for the CMake side.

## Repository layout

A library separates its public API from its implementation, and organises both by module:

```text
include/myproj/         # public headers: the API consumers see, and nothing else
  errors.h              # the exception hierarchy; see the error handling fragment
  archive/
    archive.h
  record/
    reader.h
src/                    # implementation: .cpp files and private headers; never main()
  CMakeLists.txt        # the module add_subdirectory calls and the aggregate
  common/
    CMakeLists.txt
  archive/
    CMakeLists.txt      # add_library(myproj_archive ...)
    archive.cpp
    archive_internal.h  # private: never included from include/
  record/
    CMakeLists.txt
examples/               # optional; see the library tier fragment
```

`include/` mirrors `src/`: `src/archive/archive.cpp` implements `include/myproj/archive/archive.h`, and a consumer writes `#include <myproj/archive/archive.h>`. The `myproj/` directory under `include/` is what makes that include unambiguous. It is the same prefix on every public header, so nothing a consumer writes can collide with another dependency's `archive.h`.

A library small enough to have no modules keeps the same two trees with no subdirectories under either. Modules are the shape it grows into; see the library tier fragment.

Generated headers, of which `version.h` is the one every library has, are written under the same `myproj/` prefix in the build tree, so `<myproj/version.h>` reads like every other public header. The template and the `configure_file` wiring are in the CMake fragments.

## Public headers

A file under `include/` is a promise. It is what a consumer compiles against and what the Doxygen fragment documents in full; a file under `src/` can change freely. Four rules keep the line where it is.

- **Self-contained.** A public header compiles on its own, with every include it needs (see Includes in the style fragment), so a consumer can include any one of them first.
- **`#pragma once`**, not an include guard. Every compiler these projects build with supports it, it cannot be mistyped, and it cannot go stale when a file is renamed.
- **Never reach down.** A public header includes other public headers and the standard library. It never includes anything from `src/`, which is not on a consumer's include path, and it never includes a dependency's headers unless that dependency is part of the API a consumer sees; see Dependencies in the CMake fragment.
- **Implementation stays in `src/`** unless the language forces it into the header: a template, a `constexpr`, a small `inline` accessor. The style fragment's allowance for header-only helpers is a ceiling in `include/`, not a target, because a public header is what every consumer recompiles on each change.

A type a consumer never names stays in `src/`, in a private header beside the implementation that uses it. Where it has to appear in a public header, as a pointer member or a friend, it goes in the `detail` namespace the style fragment describes.

## The two headers every library has

`errors.h` declares the exception hierarchy, rooted at `myproj::Error`, and is the header a consumer includes to catch anything the library throws; the error handling fragment gives its shape. `version.h` is generated from `project(... VERSION ...)` and gives the version as constants. A consumer that includes nothing else can still ask what it linked and catch what it throws.

## What the README says

A library's README is the first thing a consumer reads and the only documentation many will. It carries, in this order:

- What the library is and is not for, in a paragraph.
- A modules table: one row per module, its namespace, and what it covers.
- How to consume it: the `add_subdirectory` and `target_link_libraries(... PRIVATE myproj::myproj)` lines, and the include prefix.
- The CMake options table: every `option()` and cache variable the root declares, with its default and effect. Keep it in step with the root `CMakeLists.txt`. An option the README does not list is one a consumer cannot find, and one it lists under an old name is worse.
- How to run the tests, including how the integration inputs are obtained where the project has that layer; see the integration testing fragment.
