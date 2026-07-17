# C++ error handling

How a failure travels from library code to a user. Assumes the C++ style fragment.

These rules are the concrete form of the `src/`-versus-`app/` split: they are what makes a CLI wrapper thin, rather than merely short.

## The boundary

Library code (everything under `src/`) reports failure and stops there. It never decides what a failure means to a user:

- Never call `exit()`, `abort()`, or `std::terminate()`. The caller owns the process lifetime.
- Never write to `stdout` or `stderr`. The caller owns the streams; a consumer may be a GUI, a daemon, or another library with no terminal attached.
- Never phrase a message for a human. Carry the facts (the path, the underlying error code, what was attempted) and let the caller do the wording.

Application code (`app/`) is where a failure becomes an exit code and a line on `stderr`. That translation lives in exactly one place: a `try`/`catch` around the real work in `main()`. An app that catches library exceptions in three different functions and prints from each has lost the boundary.

## Exception hierarchy

Exceptions are the mechanism. `std::expected` is C++23 and the target standard here is C++20 (17 on older projects), so it is not available; do not hand-roll a substitute.

Define a single root exception so a consumer can catch everything the library throws with one handler, and derive specific types so they can catch narrowly when they care. Root the hierarchy at `std::runtime_error`, which gives `what()` for free:

```cpp
// include/mylib/errors.h
#pragma once
#include <stdexcept>
#include <string>

namespace mylib {

/// Base for every exception thrown by mylib
class Error : public std::runtime_error {
public:
    explicit Error(const std::string &message) : std::runtime_error(message) {}
};

/// Base for failures reading or writing an archive
class ArchiveError : public Error {
public:
    explicit ArchiveError(const std::string &message) : Error(message) {}
};

/// Thrown when an archive cannot be opened
class ArchiveOpenError : public ArchiveError {
public:
    ArchiveOpenError(std::string path, int error_code)
        : ArchiveError("could not open archive: " + path),
          path_(std::move(path)), error_code_(error_code) {}

    const std::string &path() const { return path_; }
    int error_code() const { return error_code_; }

private:
    std::string path_;
    int error_code_;
};

}  // namespace mylib
```

- Every exception class gets a Doxygen comment saying when it is thrown; see the Doxygen fragment.
- Public API functions throw the library's own types, never a bare `std::runtime_error`, `std::invalid_argument`, or a third-party library's exception. Catch a dependency's exception at the boundary and rethrow as your own with `std::throw_with_nested` where the original matters.
- Carry structured data as members (`path()`, `error_code()`), not just a formatted string. A caller that wants to retry needs the path, not prose.
- Exception types live in `include/<lib>/errors.h` so a consumer imports them from one place.

## What is not an exception

An exception is for a failure the caller did not expect. A result the caller asks about routinely is a return value:

```cpp
// Good - absence is a normal answer to this question
std::optional<Entry> FindEntry(const std::string &name) const;

// Bad - throwing for a routine miss forces try/catch into normal control flow
Entry FindEntry(const std::string &name) const;  // throws EntryNotFoundError
```

Use `std::optional` for "may legitimately be absent", a `bool` return for "worked or did not, and the reason is obvious", and an exception when the reason matters and the caller cannot reasonably continue.

Never use exceptions for control flow across a loop body; the cost is real and the intent is unclear.

## Catching in the app

`main()` catches the library root, prints to `stderr`, and returns an exit code. Nothing below `main()` prints:

```cpp
// app/main.cpp
#include <iostream>
#include <mylib/errors.h>

int main(int argc, char **argv) {
    try {
        // parse arguments, call into mylib
        return 0;
    } catch (const mylib::Error &e) {
        std::cerr << "error: " << e.what() << "\n";
        return 1;
    } catch (const std::exception &e) {
        std::cerr << "unexpected error: " << e.what() << "\n";
        return 2;
    }
}
```

Exit codes are part of the CLI contract and are asserted by functional tests (see the functional testing fragment). Use `0` for success, `1` for an expected failure the user can act on, and `2` for a bug or an unhandled condition. A project needing finer-grained codes documents them in the README and keeps them stable across releases.
