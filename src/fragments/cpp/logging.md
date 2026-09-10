# C++ output

A CLI writes to two streams and they carry different things. Getting the split wrong is what makes a tool awkward to use in a pipeline, and it is the most common output mistake there is.

## The two streams

- **stdout** carries exactly what the user asked for, and nothing else. It is the tool's return value: it will be piped, redirected to a file, and parsed by something that was not written yet.
- **stderr** carries everything else. Progress, warnings, and errors.

The test is simple. Redirect stdout to a file and the file should contain the answer and nothing more. A progress line, an "added 3 files" note, or a warning landing in that file means the split is wrong.

Which marker each message takes is settled in the core conventions, and the vocabulary there is the whole of it. A marker says what kind of message a line is, never which stream it goes to; those are separate decisions and this fragment owns the second.

## Library code takes a stream

Everything under `src/` writes to an `std::ostream &` it was handed. It never names `std::cout` or `std::cerr`, exactly as it never calls `exit()` and never decides what a failure means to a user; see the error handling fragment for the boundary this continues.

```cpp
// Good: the caller decides where this lands
void ListEntries(const Archive &archive, std::ostream &out);

// Bad: unusable from a GUI, a daemon, or a test
void ListEntries(const Archive &archive);
```

A free function takes the stream as a parameter. A class that emits over its lifetime takes it once in the constructor and stores a reference, rather than threading it through every method:

```cpp
class Extractor {
public:
    Extractor(Archive &archive, std::ostream &out, std::ostream &err)
        : archive_(archive), out_(out), err_(err) {}

private:
    Archive &archive_;
    std::ostream &out_;
    std::ostream &err_;
};
```

Take `std::ostream &`, never `std::ostream *` and never a template parameter. A reference cannot be null, so there is no absent case to handle, and the base class already covers every destination a caller might have: a file, a `std::ostringstream`, or `std::cout` itself.

`app/` is the only place the real streams are named, and it names them once:

```cpp
// app/main.cpp
Extractor extractor(archive, std::cout, std::cerr);
```

## Why this rather than a logger

This is the whole mechanism. There is no logging library, no severity levels, and no `--debug` flag, because a command line tool that prints its answer and its commentary has nothing left for a level to select. The standard library offers no logger to build on, and a hand-rolled one here would be machinery in front of two `operator<<` calls.

A tool that grows a genuine need for levels can add one later. Adding it before the need exists teaches a user a flag that changes nothing.

## What it buys

Injecting the stream is what makes output testable. A unit test captures into a `std::ostringstream` and asserts on the result, with no subprocess, no temporary file and no global state:

```cpp
TEST_CASE("ListEntries", "[list]") {
    std::ostringstream out;
    ListEntries(archive, out);
    REQUIRE_THAT(out.str(), Catch::Matchers::ContainsSubstring("readme.txt"));
}
```

Without it, the only way to see what a function printed is to spawn the binary and read its stdout, which moves an ordinary unit test into the functional layer and makes it slower, coarser and harder to read. That is the argument for the whole rule: the boundary is not bureaucracy, it is the difference between a test that links a function and a test that runs a process. See the testing fragment.

## Writing binary to stdout

A subcommand that dumps file contents writes bytes rather than text, and Windows breaks that by default: the C runtime opens stdout in text mode, so every `0x0A` on the way out becomes `0x0D 0x0A` and the output is silently corrupted. This is the same failure the style fragment describes for `std::ios::binary` on a file, arriving through the one stream that cannot be opened with a flag.

Switch the mode before writing, and only around the write:

```cpp
#ifdef _WIN32
#include <fcntl.h>
#include <io.h>
#endif

void WriteRaw(std::ostream &out, const char *data, std::streamsize size) {
#ifdef _WIN32
    if (&out == &std::cout) {
        _setmode(_fileno(stdout), _O_BINARY);
    }
#endif
    out.write(data, size);
}
```

The guard matters both ways. On POSIX there is no text mode and nothing to do, and the mode is a property of the file descriptor rather than of the `std::ostream`, so switching it when the caller passed an `std::ostringstream` would change the real stdout for the rest of the process.

CI proves nothing here, because a test comparing text output passes either way. It is caught by a test that writes a file containing `0x0A` and compares the bytes back, which is worth having wherever a project has a subcommand like this.
