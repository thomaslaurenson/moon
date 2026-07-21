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
option(MYLIB_BUILD_FUZZERS "Build libFuzzer harnesses (requires Clang)" OFF)

if(MYLIB_BUILD_FUZZERS)
    if(NOT CMAKE_CXX_COMPILER_ID MATCHES "Clang")
        message(FATAL_ERROR "MYLIB_BUILD_FUZZERS requires Clang (-fsanitize=fuzzer)")
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
# libFuzzer harnesses (Clang only). Built with -DMYLIB_BUILD_FUZZERS=ON.
# Developer and CI tools, never part of the shipped library.

set(FUZZ_FLAGS -g -O1 -fsanitize=fuzzer,address,undefined -fno-omit-frame-pointer)

function(add_fuzzer name)
    add_executable(${name} ${name}.cpp)
    target_link_libraries(${name} PRIVATE mylib::mylib)
    target_compile_options(${name} PRIVATE ${FUZZ_FLAGS})
    target_link_options(${name} PRIVATE ${FUZZ_FLAGS})
endfunction()

add_fuzzer(fuzz_archive)
add_fuzzer(fuzz_record)
```

A harness is not registered with `catch_discover_tests`: it runs forever by design and is not a pass/fail test case. It is driven from the Makefile instead.

## Writing a harness

`LLVMFuzzerTestOneInput` takes a buffer and hands it to the parser. The body must not assert on the result: any input is legal input to a parser, and returning an error is a correct outcome. The only failures a harness reports are the ones the sanitizers and the runtime detect for it: a segfault, a leak, a read past the end, a hang.

```cpp
// test/fuzz/fuzz_archive.cpp
#include <cstddef>
#include <cstdint>
#include <mylib/errors.h>
#include <mylib/archive/archive.h>

extern "C" int LLVMFuzzerTestOneInput(const uint8_t *data, size_t size) {
    try {
        auto archive = mylib::archive::Archive::FromMemory(data, size);
        for (const auto &entry : archive.Entries()) {
            (void)archive.Read(entry);
        }
    } catch (const mylib::Error &) {
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

## Makefile targets

```makefile
CLANG_VERSION ?= 18
FUZZ_TIME     ?= 60

.PHONY: configure_fuzz
configure_fuzz: ## Configure with libFuzzer harnesses (requires Clang)
	cmake -B build-fuzz \
	  -DCMAKE_BUILD_TYPE=Debug \
	  -DMYLIB_BUILD_FUZZERS=ON \
	  -DCMAKE_CXX_COMPILER=clang++-$(CLANG_VERSION)

.PHONY: build_fuzz
build_fuzz: configure_fuzz ## Configure and build the fuzz harnesses
	cmake --build build-fuzz --parallel $(JOBS)

.PHONY: fuzz
fuzz: ## Run one harness for FUZZ_TIME seconds (requires: NAME=fuzz_archive)
	@if [ -z "$(NAME)" ]; then echo "Error: set NAME=fuzz_archive" >&2; exit 1; fi
	./build-fuzz/bin/$(NAME) -max_total_time=$(FUZZ_TIME) test/fuzz/corpus/$(subst fuzz_,,$(NAME))
```

The fuzz build gets its own `build-fuzz/` directory rather than sharing `build/`. This is the one place a second build directory is justified: the harnesses need Clang and a sanitizer runtime that the normal build has no reason to carry, and reconfiguring `build/` back and forth would rebuild the world each time.
