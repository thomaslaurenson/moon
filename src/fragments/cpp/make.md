# C++ Makefile targets

Targets common to every C++ project; see the Makefile conventions fragment for the structure they sit in, the help target and the `##@` sections. The rules these recipes implement live elsewhere: the CMake fragment owns the build directories and the clang pin, the testing fragments own the layers, and the tooling fragment owns embedded assets. This fragment is where every target is written down, so a workflow calling `make <target>` finds it defined in exactly one place; `get_changelog` is the one exception, being the same in every language, and lives in the Makefile conventions fragment. A workflow calling a target no fragment defines is a scaffolding bug that only surfaces on a real release, in the job that publishes it.

## Variables

Declared once, at the top, before the first target:

```makefile
# Project-owned C++ directories, in whichever of them this tier actually has
CPP_DIRS      := $(wildcard include src app examples test)
CPP_LINT_DIRS := $(wildcard src app)
JOBS          ?= $(shell nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 4)
BUILD_TYPE    ?= Debug
CMAKE_ARGS    ?=
LINT_DIR      ?= build/lint
FUZZ_TIME     ?= 60
INTEGRATION_DATA ?= $(MYPROJ_INTEGRATION_DATA)

CLANG_VERSION ?= 18
CLANG_CC      ?= $(shell command -v clang-$(CLANG_VERSION) || echo clang)
CLANG_FORMAT  ?= $(shell command -v clang-format-$(CLANG_VERSION) || echo clang-format)
CLANG_CXX     ?= $(shell command -v clang++-$(CLANG_VERSION) || echo clang++)
CLANG_TIDY    ?= $(shell command -v clang-tidy-$(CLANG_VERSION) || echo clang-tidy)
LLVM_PROFDATA ?= $(shell command -v llvm-profdata-$(CLANG_VERSION) || echo llvm-profdata)
LLVM_COV      ?= $(shell command -v llvm-cov-$(CLANG_VERSION) || echo llvm-cov)
GCC_INSTALL_DIR := $(shell f=$$(gcc -print-libgcc-file-name 2>/dev/null) && dirname "$$f")
```

- `CPP_DIRS` and `CPP_LINT_DIRS` take their directory list from `wildcard`, so one Makefile covers every tier: a library has no `app/`, an application has no `include/`, and the expansion simply omits what is absent rather than failing. `examples/` is formatted but not tidied, because its programs are only configured when the examples option is on, so a lint configure has no compile commands for them; `test/` is formatted but not tidied for the reason the CMake fragment gives.
- `JOBS` is declared once and every `cmake --build` and `ctest` below reuses it.
- `BUILD_TYPE` defaults to `Debug`, the everyday configuration, and is overridable so a CI job can test the shipped one with `make configure BUILD_TYPE=Release`. It is a variable rather than a `CMAKE_ARGS` flag because it is the one setting a developer changes often enough to deserve a name.
- `CMAKE_ARGS` passes extra `-D` flags through to `cmake` (for example CI's `-DCMAKE_CXX_COMPILER=clang++-18`, or `-DMYPROJ_BINARY_PATH_OVERRIDE=...`); it is empty for a normal local configure.
- `INTEGRATION_DATA` defaults to the environment variable named after the CMake cache variable, so a machine that already exports it needs nothing on the command line.
- The clang tools are resolved rather than named, because how the pinned major version is installed differs per platform (see Clang tooling in the CMake fragment): Debian and Ubuntu install versioned binaries such as `clang-format-18`, while Homebrew and the LLVM Windows installer provide an unversioned `clang-format` from a versioned install. Prefer the versioned name, fall back to the plain one, and let either be overridden from the command line (`make format CLANG_FORMAT=/opt/homebrew/opt/llvm/bin/clang-format`). Falling back to the bare name rather than failing keeps the failure legible: an absent tool reports `clang-format: command not found`, which is clearer than a Make-level error about an empty variable.
- `llvm-profdata` and `llvm-cov` are resolved here with the rest, rather than beside the coverage target that uses them, so every clang tool the project shells out to is named in one block. They are packaged and versioned exactly like `clang-format`, so they need the same fallback.
- `GCC_INSTALL_DIR` is empty on a machine with no GCC, and `configure_lint` passes the flag only when it is set; see LINT. The `&&` is what keeps it empty: `dirname` of an empty string is `.`, so a form that runs `dirname` unconditionally passes `--gcc-install-dir=.` wherever GCC is absent.
- `CLANG_CC` is used only by a project that compiles C, which adds `-DCMAKE_C_COMPILER=$(CLANG_CC)` beside the C++ compiler in `configure_lint`, `configure_coverage` and `configure_fuzz`; see Projects that compile C in the CMake fragment. It is resolved here regardless, so a C dependency added later is one flag per configure rather than a new variable.

## BUILD

```makefile
##@ BUILD

.PHONY: configure
configure: ## Configure the cmake build
	cmake -B build/dev \
	  -DCMAKE_BUILD_TYPE=$(BUILD_TYPE) \
	  -DCMAKE_EXPORT_COMPILE_COMMANDS=ON \
	  $(CMAKE_ARGS)

.PHONY: build
build: configure ## Build the project
	cmake --build build/dev --parallel $(JOBS)
```

- `configure` is `.PHONY` and always runs, rather than being a rule on `build/dev/CMakeCache.txt`. Keying it to the cache file looks like a saving and is a trap: `make configure CMAKE_ARGS=-DFOO=ON` then does nothing at all on a tree that already configured, silently ignoring the flags, and the reconfigure it avoids takes under a second.
- `configure` stays compiler-neutral. CI builds under both GCC and clang by passing `-DCMAKE_CXX_COMPILER` through `CMAKE_ARGS`, and the configurations below that need clang pin it themselves.
- `build` depends on `configure`, and every test target depends on `build`, so any entry point works from a fresh clone. CMake keeps cached `-D` values across a reconfigure, so the repeat does not discard the compiler or `-Werror` a CI job set with `make configure CMAKE_ARGS=...`.

A project that needs a configuration of its own, such as a 32-bit build in `build/32`, adds a `configure_<name>` and `test_<name>` pair with the shape of `configure_asan` and `test_asan` below: its own directory under `build/`, named for what makes it different (see Build directory in the CMake fragment), and the unit layer run there.

## TEST

```makefile
##@ TEST

.PHONY: test
test: build ## Run Catch2 unit tests
	ctest --test-dir build/dev --output-on-failure --parallel $(JOBS) -L unit

.PHONY: test_verbose
test_verbose: build ## Run unit tests with verbose Catch2 output
	ctest --test-dir build/dev --verbose -L unit

.PHONY: test_functional
test_functional: build ## Run Catch2 functional tests against the built binary
	ctest --test-dir build/dev --output-on-failure --parallel $(JOBS) -L functional

.PHONY: test_all
test_all: build ## Run every test layer built into the current configure
	ctest --test-dir build/dev --output-on-failure --parallel $(JOBS)
```

- `test` runs the unit layer alone and is the everyday target; see cpp/testing for why that layer and no other. `-L` selects by label, never `-R`, for the reason given there.
- `test_functional` exists only in a tier that ships a binary. It depends on `build` because the layer spawns the compiled binary, and a stale or absent one is a failure with a confusing message rather than a test result.
- `test_all` runs whatever the current configure contains, which is the unit layer alone unless another was configured in. It is defined once, here, so a tier with a functional layer and a tier without see the same target.
- Every test target depends on `build`, which depends on `configure`, so `make ci` runs on a clean checkout. Without that chain `ctest` fails on a missing `build/dev` with a message about the directory rather than about the missing build.

### Integration

```makefile
.PHONY: configure_integration
configure_integration: ## Configure build/dev with integration tests (requires: INTEGRATION_DATA)
	@if [ -z "$(INTEGRATION_DATA)" ]; then \
	  echo "Error: set INTEGRATION_DATA=/path/to/dataset or MYPROJ_INTEGRATION_DATA" >&2; exit 1; \
	fi
	cmake -B build/dev \
	  -DCMAKE_BUILD_TYPE=$(BUILD_TYPE) \
	  -DCMAKE_EXPORT_COMPILE_COMMANDS=ON \
	  -DMYPROJ_INTEGRATION=ON \
	  -DMYPROJ_INTEGRATION_DATA="$(INTEGRATION_DATA)" \
	  $(CMAKE_ARGS)

.PHONY: test_integration
test_integration: configure_integration ## Build and run the integration layer
	cmake --build build/dev --parallel $(JOBS)
	ctest --test-dir build/dev --output-on-failure -L integration
```

- The layer is off by default, so `test_integration` depends on `configure_integration` rather than on `build`: a plain `configure` never turns it on, and `ctest -L integration` against a tree with no such tests reports that nothing was found rather than running anything.
- The guard on `INTEGRATION_DATA` belongs in a project whose variable has no default. A project that does default the variable drops it, because the condition can never be true and a check that cannot fire is one more thing to read; see cpp/testing-integration.
- No `--parallel`: the layer's tests share real inputs the machine already has, and two of them touching the same file at once is a failure the layer did not exist to find.
- `fetch_integration_data`: download the inputs into `$(INTEGRATION_DATA)`, where that can happen without a human deciding anything. Offered, not required: a project whose inputs are not downloadable omits it and the README carries the answer instead (see cpp/testing-integration).

### Sanitizers

```makefile
.PHONY: configure_asan
configure_asan: ## Configure build/asan with Address + UB sanitizers
	cmake -B build/asan \
	  -DCMAKE_BUILD_TYPE=Debug \
	  -DMYPROJ_ASAN=ON \
	  -DMYPROJ_BUILD_TESTING=ON \
	  $(CMAKE_ARGS)

.PHONY: test_asan
test_asan: configure_asan ## Build and run the unit tests under sanitizers
	cmake --build build/asan --parallel $(JOBS)
	ctest --test-dir build/asan --output-on-failure --parallel $(JOBS) -L unit
```

A sanitized build changes code generation, so it gets its own directory and cannot share `build/dev`. `test_asan` runs the unit layer only: that layer needs no external data or server, so it is the one that can run anywhere, and sanitizer findings in it point at the project's own code rather than at a fixture. Without these targets the option is reachable only through a raw `cmake -D` invocation, which the Makefile exists to prevent. `test.yml` calls `test_asan` as a job of its own; see cpp/workflows.

### Coverage

```makefile
.PHONY: configure_coverage
configure_coverage: ## Configure build/coverage with clang source-based coverage
	cmake -B build/coverage \
	  -DCMAKE_BUILD_TYPE=Debug \
	  -DMYPROJ_BUILD_TESTING=ON \
	  -DCMAKE_CXX_COMPILER=$(CLANG_CXX) \
	  -DCMAKE_CXX_FLAGS="-fprofile-instr-generate -fcoverage-mapping" \
	  -DCMAKE_EXE_LINKER_FLAGS="-fprofile-instr-generate" \
	  $(CMAKE_ARGS)

.PHONY: test_coverage
test_coverage: configure_coverage ## Report unit-test coverage
	cmake --build build/coverage --parallel $(JOBS) --target myproj_unit_tests
	LLVM_PROFILE_FILE=build/coverage/unit.profraw ./build/coverage/bin/myproj_unit_tests
	$(LLVM_PROFDATA) merge -o build/coverage/unit.profdata build/coverage/unit.profraw
	$(LLVM_COV) report ./build/coverage/bin/myproj_unit_tests \
	  -instr-profile=build/coverage/unit.profdata \
	  --ignore-filename-regex="extern/|test/"
```

- Its own `build/coverage` directory, because the instrumentation changes code generation, and the compiler is pinned to clang whatever the everyday build uses. The unit layer alone, for the reason cpp/testing gives under Coverage.
- `--ignore-filename-regex` keeps vendored code and the tests themselves out of the report.
- The report goes to stdout and nothing publishes it: the `TOTAL` line is the number copied by hand into the coverage badge on each release (see cpp/badges).

### Fuzzing

```makefile
.PHONY: configure_fuzz
configure_fuzz: ## Configure build/fuzz with libFuzzer harnesses (requires Clang)
	cmake -B build/fuzz \
	  -DCMAKE_BUILD_TYPE=Debug \
	  -DMYPROJ_BUILD_FUZZERS=ON \
	  -DCMAKE_CXX_COMPILER=$(CLANG_CXX) \
	  $(CMAKE_ARGS)

.PHONY: build_fuzz
build_fuzz: configure_fuzz ## Configure and build the fuzz harnesses
	cmake --build build/fuzz --parallel $(JOBS)

.PHONY: fuzz
fuzz: build_fuzz ## Run one harness for FUZZ_TIME seconds (requires: NAME=archive)
	@if [ -z "$(NAME)" ]; then echo "Error: set NAME=archive" >&2; exit 1; fi
	@corpus=test/fuzz/corpus/$(NAME); \
	  [ -d "$$corpus" ] || corpus=; \
	  ./build/fuzz/bin/myproj_fuzz_$(NAME) -max_total_time=$(FUZZ_TIME) $$corpus
```

- `NAME` is the bare harness name, the same string `add_fuzzer` takes, so `make fuzz NAME=archive` runs `myproj_fuzz_archive` against `test/fuzz/corpus/archive`.
- The corpus directory is passed only when it exists. libFuzzer treats a corpus path it was given as mandatory and exits 1 with `ERROR: The required directory ... does not exist`, so hardcoding the path makes a harness unrunnable in a project that has not seeded one, which cpp/testing-fuzz allows. Passing nothing is the supported way to run without a corpus, and libFuzzer then generates from scratch.
- Its own `build/fuzz` directory, because `-fsanitize=fuzzer` is clang-only and a build tree cannot change compiler after its first configure; see cpp/testing-fuzz.

## LINT

```makefile
##@ LINT

.PHONY: format
format: ## Format all source files with clang-format
	find $(CPP_DIRS) \( -name "*.cpp" -o -name "*.h" \) | xargs $(CLANG_FORMAT) -i

.PHONY: check_format
check_format: ## Check formatting without modifying files
	find $(CPP_DIRS) \( -name "*.cpp" -o -name "*.h" \) | xargs $(CLANG_FORMAT) --dry-run --Werror

.PHONY: configure_lint
configure_lint: ## Configure $(LINT_DIR) with clang++, so clang-tidy can parse the sources
	cmake -B $(LINT_DIR) \
	  -DCMAKE_BUILD_TYPE=Debug \
	  -DCMAKE_EXPORT_COMPILE_COMMANDS=ON \
	  -DCMAKE_CXX_COMPILER=$(CLANG_CXX) \
	  $(if $(GCC_INSTALL_DIR),-DCMAKE_CXX_FLAGS="--gcc-install-dir=$(GCC_INSTALL_DIR)") \
	  $(CMAKE_ARGS)

.PHONY: check_lint
check_lint: configure_lint ## Run clang-tidy static analysis
	$(CLANG_TIDY) --quiet -p $(LINT_DIR) \
	--header-filter="$(CURDIR)/(include|src|app)/.*" $$(find $(CPP_LINT_DIRS) -name "*.cpp") 2>&1 \
	| grep -v " warnings generated"; \
	exit $${PIPESTATUS[0]}

.PHONY: check_embed
check_embed: ## Syntax-check the embedded completion scripts with every shell present
	bash -n completion/myproj.bash
	@if command -v zsh >/dev/null; then zsh -n completion/myproj.zsh; \
	  else echo "[*] zsh not installed, skipped"; fi
	@if command -v fish >/dev/null; then fish --no-execute completion/myproj.fish; \
	  else echo "[*] fish not installed, skipped"; fi
	@if command -v pwsh >/dev/null; then pwsh -NoProfile -Command \
	  '[scriptblock]::Create((Get-Content -Raw completion/myproj.ps1)) | Out-Null'; \
	  else echo "[*] pwsh not installed, skipped"; fi
```

- `configure_lint` writes its own directory with the compiler pinned to clang, because clang-tidy resolves headers through the compiler that produced `compile_commands.json` and a GCC-configured tree leaves it emitting diagnostics from a broken AST; see Configuring for clang-tidy in the CMake fragment. It only configures, never builds, so it produces `compile_commands.json` without a second compile of the project.
- `--gcc-install-dir` tells clang which libstdc++ to use when the two toolchains are installed side by side, which is the normal state on a Linux runner and on most developer machines. It is passed only when `gcc` is present to ask: on a machine with no GCC the shell call yields an empty string, and `--gcc-install-dir=` with nothing after it is rejected by clang, so an unconditional flag would break `make check_lint` everywhere GCC is not installed.
- `check_lint` depends on `configure_lint`, so it needs no separate `make configure` first and reads `$(LINT_DIR)/compile_commands.json` rather than the everyday build's.
- `--quiet` suppresses the "Suppressed N warnings" summary and hint lines.
- `find` covers every implementation file in those directories, including nested subdirectories; a bare `src/*.cpp` glob would miss anything below the top level.
- `include/` is formatted but not tidied directly: its headers carry no `.cpp` of their own, and clang-tidy reaches them through the `--header-filter` when it analyses the `src/` files that include them.
- `--header-filter="$(CURDIR)/(include|src|app)/.*"` limits diagnostic output to project headers; `extern/` headers are already excluded as system headers (see Including extern/ headers in the CMake fragment) but this provides belt-and-suspenders coverage.
- `grep -v " warnings generated"` strips the per-file progress counter, which counts all warnings before any filtering and is always misleading when third-party headers are present; `exit $${PIPESTATUS[0]}` preserves clang-tidy's exit code through the pipe.
- `check_embed` parses each completion script with its own shell, because the compiler proves only that the file was read (see the tooling fragment). `bash` is unconditional, since the Makefile already needs it; the other three run where the shell is installed and say so where it is not, so a developer without fish sees a skip rather than a failure. A library that embeds a resource of its own writes the matching check the same way, and a project that embeds nothing omits the target.

## GET

```makefile
##@ GET

.PHONY: get_version
get_version: ## Print the project version from CMakeLists.txt (fails if absent)
	@awk '\
	  /cmake_minimum_required/ { next } \
	  match($$0, /VERSION[ \t]+[0-9]+\.[0-9]+\.[0-9]+/) { \
	    v = substr($$0, RSTART, RLENGTH); sub(/VERSION[ \t]+/, "", v); \
	    print v; found = 1; exit } \
	  END { if (!found) exit 1 }' CMakeLists.txt
```

- `get_version` reads the version out of `project(... VERSION X.Y.Z)`, which cpp/style makes the single place a version is declared. It skips the `cmake_minimum_required` line first, because that also says `VERSION` and comes earlier in the file; a three-component minimum such as `3.21.0` would otherwise be reported as the project version. Nothing in CI calls it, and it is worth having anyway: it is what lets you check that the tag about to be pushed matches what the build will report, which is the mismatch nobody notices until a release is out. It uses only POSIX `awk`, so it behaves the same under gawk, mawk and busybox.
- `get_changelog` is the shared recipe in the Makefile conventions fragment, and `release.yml` calls it directly (see cpp/workflows); a release that reaches that step without the target fails after the artefacts are already built.

## CI

```makefile
##@ CI

.PHONY: check_all
check_all: check_format check_lint ## Run every static check

.PHONY: ci
ci: check_all test ## Run the checks CI runs

.PHONY: clean
clean: ## Remove build output and release artefacts
	rm -rf build dist install.sh install.ps1 checksums.txt checksums.txt.sigstore.json
```

- `check_all`: `check_format check_lint`, plus `check_embed` where the project embeds anything. This is the only place the static checks are listed: `lint.yml` calls it and `ci` composes it, so there is no second copy to fall out of step.
- `ci`: `check_all test`, plus `test_functional` in a tier that ships a binary. It composes the aggregates the workflows call and has no recipe of its own, as the Makefile conventions fragment requires; it cannot mirror CI's two-compiler matrix and does not try.
- `clean` removes `build` entirely, not `$(LINT_DIR)` or `build/dev`. There are several build directories (`dev`, `lint`, `asan`, `coverage`, `fuzz`) and a clean that leaves the others behind is the one that gets debugged at the wrong moment. `dist/` and the four files gpipe writes into the repository root belong here too in a tier that releases with gpipe (see the gpipe fragment for why each is build output); a plain library never runs gpipe and removes `build` alone, since `clean` only removes what the project can rebuild.

## Which targets a tier has

- Every tier: `configure`, `build`, `test`, `test_verbose`, `test_all`, `configure_asan`, `test_asan`, `configure_coverage`, `test_coverage`, `format`, `check_format`, `configure_lint`, `check_lint`, `get_version`, `get_changelog`, `check_all`, `ci` and `clean`.
- A tier that ships a binary adds `test_functional` to `ci` and `check_embed` to `check_all`: every CLI embeds its completion scripts, so every CLI has something to check (see the CLI scaffolding fragment).
- A project with the layer adds `configure_integration`, `test_integration` and, where the inputs are downloadable, `fetch_integration_data`; one with fuzz harnesses adds `configure_fuzz`, `build_fuzz` and `fuzz`.
- A library that embeds a resource of its own adds `check_embed` the same way.
