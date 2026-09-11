# C++ CLI scaffolding

How `app/` is laid out and wired. Assumes the C++ style, error handling and output fragments, and the application or library-with-CLI tier fragment for the CMake side.

CLI11 is the argument parser, pinned as a submodule under `extern/CLI11` like every other dependency. It is header-only, it validates arguments declaratively, and it generates help that needs no maintenance. Nothing else here parses argv.

## Repository layout

A tier that ships a binary adds `app/` and the release files to the universal layout in the CMake fragment. In an application `src/` is an internal core with no `include/`; in a library with a bundled CLI it is the library the library scaffolding fragment lays out, and `app/` sits beside it unchanged:

```
src/                   # the core: every line of logic, built as a library; never main()
  CMakeLists.txt
app/                   # the CLI: main() plus argument wiring, and nothing else
  CMakeLists.txt
completion/            # one completion script per shell; see Shell completion
Dockerfile             # static musl build into scratch; see the C++ Docker fragment
.dockerignore
.gpipe.yml             # installer and checksum config; see the release fragment
.github/workflows/
  build.yml            # the artifacts; see the shipped-binary workflows fragment
  release.yml
  prerelease.yml
```

One Dockerfile, building a statically linked musl binary into a scratch image; the C++ Docker fragment has it, and the shipped-binary workflows fragment says why Linux ships one static binary per architecture. macOS and Windows binaries are built natively on their own runners, not via Docker.

## The app/ directory

`app/` holds the entry point and one file per subcommand, and nothing else:

```
app/
  CMakeLists.txt
  main.cpp        # builds the App, registers subcommands, catches
  add.cpp         # RegisterAdd
  extract.cpp     # RegisterExtract
  list.cpp        # RegisterList
  commands.h      # declares every Register* function
  validators.h    # shared CLI::Validator objects
  validators.cpp
```

One file per subcommand, named for the subcommand. The logic each one calls lives in `src/`, so these files stay thin: they declare options, then hand the settled values to the core.

## The root App

`main.cpp` builds the `CLI::App`, registers each subcommand through its own function, and catches:

```cpp
// app/main.cpp
#include <CLI/CLI.hpp>
#include <myproj/errors.h>
#include <myproj/version.h>

#include "commands.h"

int main(int argc, char **argv) {
    CLI::App app{"Reads and writes archives"};
    app.require_subcommand(1);
    app.set_version_flag("--version", myproj::version_string);

    RegisterAdd(app);
    RegisterExtract(app);
    RegisterList(app);

    try {
        app.parse(argc, argv);
    } catch (const CLI::ParseError &e) {
        return app.exit(e);
    } catch (const myproj::Error &e) {
        std::cerr << "[!] " << e.what() << "\n";
        return 1;
    } catch (const std::exception &e) {
        std::cerr << "[!] unexpected: " << e.what() << "\n";
        return 2;
    }
    return 0;
}
```

- `require_subcommand(1)` is what makes a bare invocation fail rather than succeeding silently. Without it CLI11 parses nothing, throws nothing, and `main` returns 0, which tells a script the command worked.
- `set_version_flag` reads `myproj::version_string` from the generated version header, whose template is in cpp/cmake.md and whose value comes from `project(... VERSION ...)`. Never a literal here.
- The subcommand callbacks run inside `app.parse`, which is why the library's own exceptions are caught around it rather than after. The exit codes are covered in the error handling fragment.
- The `<myproj/...>` includes are the lib-cli form. An application has no `include/`: its own headers sit in `src/` and `app/` includes them by name, `#include "errors.h"`, through the core's public include directory, while `<myproj/version.h>` is generated into an include tree in both tiers; see cmake-app.md.

## One registration function per subcommand

Each subcommand is a `Register<Name>(CLI::App &app)` function declared in `commands.h` and defined in its own file:

```cpp
// app/list.cpp
#include "commands.h"

#include <iostream>
#include <memory>
#include <string>

#include <myproj/archive.h>

void RegisterList(CLI::App &app) {
    struct Options {
        std::string target;
        bool detailed = false;
    };

    auto opts = std::make_shared<Options>();
    auto *sub = app.add_subcommand("list", "List the entries in an archive");

    sub->add_option("target", opts->target, "Archive to read")->required()->check(CLI::ExistingFile);
    sub->add_flag("-d,--detailed", opts->detailed, "Show size and timestamp for each entry");

    sub->callback([opts]() {
        myproj::ListEntries(opts->target, opts->detailed, std::cout);
    });
}
```

Three parts of that shape are load-bearing.

**An `Options` struct, not loose locals.** CLI11 binds each option to the address it is given, so the storage has to outlive registration. A struct keeps every value for one subcommand together, and the `shared_ptr` is what lets the callback own it after `Register` returns.

**A callback, not a dispatch chain.** The alternative is one `main` that declares every flag for every subcommand up front and ends in a chain of `if (app.got_subcommand(x))`. It is the shape a CLI grows into by accident, and it fails in three ways at once: the variables for every subcommand share one scope, so names grow prefixes to stay apart; `main` grows without limit; and each branch calls a function with a long positional argument list.

That last one is the real cost. A handler taking sixteen parameters, several of them adjacent `int64_t` flags, compiles perfectly with any two of them swapped:

```cpp
// Two adjacent parameters, silently interchangeable
return HandleCreate(target, path, output, sign, locale, profile, version,
                    stream_flags, sector_size, raw_chunk_size, file_flags1,
                    file_flags2, file_flags3, attr_flags, compression, next);
```

A named field cannot be swapped by accident, and the callback needs no argument list at all.

**The callback is thin.** It calls into `src/` and passes the stream, and does nothing else. A callback that opens the file and formats the output itself has put logic in the one place no unit test can link; see the output and testing fragments.

## Flags

- Long names are kebab-case and lowercase: `--dry-run`, `--keep-structure`, never `--dryRun`.
- Help text is a capitalised fragment with no trailing full stop, matching the help CLI11 generates beside it.
- A shorthand means one thing across the whole binary. Giving `-l` to `--long` under one subcommand and `--listfile` under another teaches a user something that then misfires, and nothing in the build can see it, because the two are registered on different subcommands.
- Add a shorthand only for a flag typed often enough to earn one. Adding one later costs nothing; removing one breaks whoever learned it.

Validate declaratively rather than in the callback. CLI11's built-in checks cover most cases, and a project-specific rule becomes a `CLI::Validator` in `app/validators.h` so the same rule reads the same way at every use:

```cpp
sub->add_option("target", opts->target, "Archive to read")->required()->check(CLI::ExistingFile);
sub->add_option("--locale", opts->locale, "Locale for the added files")->check(locale_valid);
```

A validator failure is a parse error, so it is reported by CLI11 with the usage text and never reaches the library.

Where a value can come from more than one place, the order is flag, then environment, then the built-in default. Resolve it in `app/` and pass the settled value down, so nothing in `src/` can tell which layer an argument came from. CLI11 reads the environment for you with `->envname("MYPROJ_TARGET")`.

## Windows option syntax

CLI11 defaults `allow_windows_style_options` to true on Windows, so `/foo` is parsed as an option there rather than as a path. An application with no `/x` style options of its own should turn it off:

```cpp
app.allow_windows_style_options(false);
```

The convention then costs a path shape and buys nothing. Leaving it on is what makes a POSIX-absolute path vanish into the parser; see the portable-input rule in the functional testing fragment.

## Shell completion

Every C++ CLI ships a `completion` subcommand that prints the script for a named shell to stdout, for bash, zsh, fish and PowerShell, and the user installs it. CLI11 generates none of them, so they are written by hand and kept in a top-level `completion/` directory, one per shell, embedded at configure time with `configure_file` so the binary needs no files beside it:

```
completion/
  myproj.bash
  myproj.zsh
  myproj.fish
  myproj.ps1
```

The generated header goes to the build tree, never into `src/` or `completion/`; see the tier fragment for the `configure_file` wiring. Declare the scripts with `CMAKE_CONFIGURE_DEPENDS` so editing one regenerates the header.

The scripts are embedded assets, so every CLI has a `check_embed` target and `check_all` includes it: `bash -n` over the bash script and the matching syntax check for each other shell the runner has. A completion script is code a user runs in their shell, and a syntax error in it compiles into the binary without complaint; see the tooling and Makefile targets fragments.

Installing the script is the user's job. Nothing in the release tooling writes to a user's shell configuration.
