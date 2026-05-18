# C++ Project Scaffolding

Standards and conventions for C++ projects. Use this as a reference when creating or
refactoring a C++ repository.

## Project Structure

```
src/
  CMakeLists.txt      # main executable target
  main.cpp
  <module>.cpp
  <module>.h
test/
  CMakeLists.txt      # unit and functional test targets
  subprocess_helper.h
  subprocess_helper.cpp
  fixtures/
    test_environment.h
    test_files.h
  functional/
    test_<subcommand>.cpp
  unit/
    test_<module>.cpp
extern/
  Catch2/             # submodule - pinned to v3.x
  subprocess.h/       # submodule - pinned to a specific commit
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
.clang-format
.clang-tidy
CMakeLists.txt        # root: project settings, options, add_subdirectory calls
Dockerfile.glibc
Dockerfile.musl
Makefile
CHANGELOG.md
README.md
```

- Business logic lives in `src/`; no header-only projects unless the implementation is 50 lines or fewer
- `extern/` contains only git submodules; never copy third-party headers manually
- Two Dockerfiles: one for glibc builds, one for musl (static) builds; see `tools/docker.md`

## Tools

| Tool | Purpose |
|---|---|
| `cmake` | Build system |
| `clang-format-18` | Formatter |
| `clang-tidy-18` | Static analysis |
| `ctest` | Test runner (via `make test`) |

Always pin clang tools to version 18. Never use the unversioned `clang-format` or
`clang-tidy` binaries.

**Not used:** `conan`, `vcpkg`, or any other package manager. All dependencies are git
submodules.

## Makefile

Adhere to the global Makefile structure in `tools/makefile.md`. Standard targets:

- `configure`: `cmake -B build -DCMAKE_BUILD_TYPE=Debug -DCMAKE_EXPORT_COMPILE_COMMANDS=ON`
- `build`: `cmake --build build`
- `fmt`: format all source files with `clang-format-18 -i`
- `fmt_check`: dry-run format check with `clang-format-18 --dry-run --Werror`
- `lint_cpp`: run `clang-tidy-18 -p build` over `src/`
- `install_clang_tools`: `sudo apt-get install -y clang-format-18 clang-tidy-18`
- `test`: `cmake --build build && cd build && ctest --output-on-failure`
- `test_unit`: run only Catch2 unit tests via `ctest -R unit`
- `test_functional`: run only Catch2 functional tests via `ctest -R functional`
- `ci`: `fmt_check lint_cpp test`
- `clean`: `rm -rf build/`

See `tools/cmake.md` for the full CMake configuration and `tools/makefile.md` for the
`# GET` section targets.

## Version

Declared in the root `CMakeLists.txt` `project()` call. See `cpp/style.md` for the
`configure_file` pattern that bakes the version into the binary at build time.
