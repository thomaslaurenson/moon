# C++ fuzz testing

libFuzzer harnesses that drive a parser with hostile input, looking for crashes, hangs and undefined behaviour. Assumes cpp/testing.md and the tier fragment.

A fuzz harness is a test, so it lives under `test/fuzz/` with every other test, never in a `fuzz/` directory of its own at the repository root. It links the library target exactly as a unit test does; what makes it different is that it has no assertions and no expected output. It feeds bytes in and lets the sanitizers decide whether something went wrong.

## When to fuzz

Fuzz any code that parses input it did not create: a file format, a network packet, a decompressor. That is the code where a malformed input is a security problem rather than a wrong answer, and it is exactly the code a unit test covers least well, because a unit test only ever supplies inputs its author thought of.

One harness per parser entry point. Do not write a harness for pure logic with no untrusted input; there is nothing for it to find.

## Structure

```
test/
  fuzz/
    CMakeLists.txt
    fuzz_archive.cpp    # drives the archive reader
    fuzz_record.cpp     # drives the record reader
    corpus/             # optional seed inputs, one directory per harness
      archive/
```

## Option

The root `CMakeLists.txt` declares a project-scoped option, default `OFF`, and fails loudly if it is on without Clang:

```cmake
option(MYPROJ_BUILD_FUZZERS "Build libFuzzer harnesses (requires Clang)" OFF)

if(MYPROJ_BUILD_FUZZERS)
    if(NOT CMAKE_CXX_COMPILER_ID MATCHES "Clang")
        message(FATAL_ERROR "MYPROJ_BUILD_FUZZERS requires Clang (-fsanitize=fuzzer)")
    endif()
endif()
```

Never use a bare `BUILD_FUZZERS`; it is as collision-prone as `BUILD_TESTING`, and a vendored dependency with the same idea will pick it up. Default `OFF` keeps the fuzzing runtime out of a normal build entirely.

`-fsanitize=fuzzer` is a Clang feature. Checking the compiler at configure time turns a confusing link error into a sentence that says what to do.

## Harness target

Each harness links its library and the libFuzzer runtime. The sanitizer flags are set on the harness itself and are deliberately independent of the project's global sanitizer option: a fuzzer without ASan finds crashes but not the memory errors that precede them, so it is always instrumented, even in a plain configure.

```cmake
# test/fuzz/CMakeLists.txt
#
# libFuzzer harnesses (Clang only). Built with -DMYPROJ_BUILD_FUZZERS=ON.
# Developer and CI tools, never part of the shipped library.

set(FUZZ_FLAGS -g -O1 -fsanitize=fuzzer,address,undefined -fno-omit-frame-pointer)

function(add_fuzzer name)
    add_executable(myproj_fuzz_${name} fuzz_${name}.cpp)
    target_link_libraries(myproj_fuzz_${name} PRIVATE myproj::myproj myproj::warnings)
    target_compile_options(myproj_fuzz_${name} PRIVATE ${FUZZ_FLAGS})
    target_link_options(myproj_fuzz_${name} PRIVATE ${FUZZ_FLAGS})
endfunction()

add_fuzzer(archive)
add_fuzzer(record)
```

`test/CMakeLists.txt` adds this directory, guarded on the same option:

```cmake
# test/CMakeLists.txt, after the unit and functional targets
if(MYPROJ_BUILD_FUZZERS)
    add_subdirectory(fuzz)
endif()
```

Without that line the harnesses are never configured and `make build_fuzz` builds nothing, with no error to explain it: `test/fuzz/CMakeLists.txt` is simply a file CMake never reads. The guard is on the tier fragment's `test/CMakeLists.txt` rather than the root, because the root adds `test/` as a whole and the fuzz layer is a layer of the test tree like any other.

The function takes the bare harness name and builds both the target name and the source name from it, so `add_fuzzer(archive)` compiles `fuzz_archive.cpp` into `myproj_fuzz_archive`. The prefix is not optional: a target called `fuzz_archive` breaks the rule that every target name carries the project prefix, and a harness is exactly as capable of colliding in a superbuild as a module is. See Target names in cpp/cmake.md.

Harnesses link the warning bar like any other target the project owns; see Warnings in cpp/cmake.md.

A harness is not registered with `catch_discover_tests`: it runs forever by design and is not a pass/fail test case. It is driven from the Makefile instead.

## Writing a harness

`LLVMFuzzerTestOneInput` takes a buffer and hands it to the parser. The body must not assert on the result: any input is legal input to a parser, and returning an error is a correct outcome. The only failures a harness reports are the ones the sanitizers and the runtime detect for it: a segfault, a leak, a read past the end, a hang.

```cpp
// test/fuzz/fuzz_archive.cpp
#include <cstddef>
#include <cstdint>
#include <myproj/archive/archive.h>
#include <myproj/errors.h>

extern "C" int LLVMFuzzerTestOneInput(const uint8_t *data, size_t size) {
    try {
        auto archive = myproj::archive::Archive::FromMemory(data, size);
        for (const auto &entry : archive.Entries()) {
            (void)archive.Read(entry);
        }
    } catch (const myproj::Error &) {
        // Rejecting malformed input is correct behaviour, not a finding
    }
    return 0;
}
```

Catch only the library's own exception root, never `...` or `std::exception`. A harness that swallows everything hides the `std::bad_alloc` from a bogus 4GB length field, which is precisely the bug worth finding. This is the same rule the other layers follow; see the error handling fragment.

Never let a harness write to disk or print. It runs millions of times.

## Corpus and findings

Seed a harness with real inputs where you have them: a corpus of valid files makes the fuzzer spend its time on interesting mutations instead of rediscovering the magic number. Keep seeds small; libFuzzer prefers many small inputs to a few large ones.

When a harness finds something, it writes the offending input to a `crash-<hash>` file. Commit that file to `test/data/` and write a unit test that reads it, before fixing the bug. The fuzzer found it once; the unit test is what stops it coming back.

## Running it

`configure_fuzz`, `build_fuzz` and `fuzz` in the Makefile targets fragment configure `build/fuzz` with clang, build the harnesses, and run one of them for `FUZZ_TIME` seconds. `make fuzz NAME=archive` runs `myproj_fuzz_archive` against `test/fuzz/corpus/archive`, with `NAME` the same bare name `add_fuzzer` takes, and the corpus passed only when the directory exists.

Fuzzing gets its own `build/fuzz` directory, and this is the clearest case for the rule in cpp/cmake.md. `-fsanitize=fuzzer` is Clang-only (see Option above), and CMake cannot change a build tree's compiler after the first configure without discarding the cache. Pointing `configure_fuzz` at `build/dev` would therefore reconfigure and rebuild everything, not just the harnesses, destroy whatever `build/dev` previously held, and charge the same cost again on the next ordinary configure. Two directories cost one `.gitignore` entry that already exists, and `rm -rf build` still cleans both.
