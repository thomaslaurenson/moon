# Bash Testing Standards

Standards and conventions for testing Bash scripts and sourced files.

## Tools

Three tools are used in combination, each covering a different layer:

- `bash -n`: syntax check, verifies the script parses without errors
- `shellcheck`: static analysis, catches common bugs and unsafe patterns
- `bats`: functional tests, verifies runtime behaviour

All three must pass before a change is considered complete.

## Syntax Check

Run `bash -n` on every script and sourced file as the first lint step. It is fast and catches parse errors before any other tool runs:

```makefile
lint:
	@printf 'bash -n  src/app.bash ... '
	@bash -n src/app.bash \
	  && printf 'ok\n' \
	  || { printf 'fail\n'; exit 1; }
```

## ShellCheck

Run `shellcheck` on every script and sourced file. ShellCheck infers the shell dialect from the shebang line; no `-s` flag is needed when `#!/usr/bin/env bash` is present. Only specify `-s bash` explicitly when a file has no shebang, such as a fragment intended to be sourced.

```makefile
lint:
	@printf 'shellcheck  src/app.bash ... '
	@shellcheck src/app.bash \
	  && printf 'ok\n' \
	  || { printf 'fail\n'; exit 1; }
```

### Disabling checks

Suppress a ShellCheck warning only when it is a deliberate false positive. Always add an inline comment explaining why the suppression is justified:

```bash
# shellcheck disable=SC2086 - word splitting is intentional here, values are
# validated against ^[A-Za-z_][A-Za-z0-9_]*$ and contain no IFS characters
printf '%s\n' $varlist
```

Never suppress `SC2086` (unquoted variables), `SC2046` (unquoted command substitution), or `SC2048` (unquoted array expansion) without a specific, documented reason.

## Bats

Use [bats-core](https://github.com/bats-core/bats-core) for functional tests. Pin bats as a git submodule under `test/extern/bats` so the version is controlled and no system install is required.

### Version requirement

Every test file must declare the minimum bats version at the top:

```bash
bats_require_minimum_version 1.7.0
```

### Structure

```
test/
  extern/
    bats/              # bats-core submodule
  fixtures/            # static input files used by tests
  helpers/             # mock executables and shared utilities
  app_bash.bats        # tests for src/app.bash
  init_sh.bats         # tests for sourced init scripts
```

- One `.bats` file per source file under test
- Fixtures live in `test/fixtures/`; never generate test data inside a test
- Mock executables live in `test/helpers/`; named after the command they replace

### setup

Every test file must define a `setup` function that configures the environment before each test. Set `REPO_ROOT` relative to `BATS_TEST_DIRNAME` so tests run correctly regardless of the working directory:

```bash
# Configure the environment before each test.
#
# Environment:
#   REPO_ROOT  - absolute path to the repository root, derived from BATS_TEST_DIRNAME
#   DATA_DIR   - path to fixture data used by tests
#   APP_CMD    - path to the mock command helper
#   SCRIPT     - path to the script under test
setup() {
  REPO_ROOT="$(cd "$BATS_TEST_DIRNAME/.." && pwd)"
  export DATA_DIR="$REPO_ROOT/test/fixtures/data"
  export APP_CMD="$REPO_ROOT/test/helpers/mock_cmd"
  SCRIPT="$REPO_ROOT/src/app.bash"
}
```

### Test naming

Test names use the pattern `group: description` where the group identifies the subcommand, function, or feature under test:

```bash
@test "list: exits 0" { ... }
@test "list: returns entries in sorted order" { ... }
@test "create: rejects invalid entry name" { ... }
@test "create: output confirms entry was written" { ... }
```

Names are descriptive enough that a failure message identifies the problem without reading the test body.

### Assertions

Use `run` to capture exit status and output, then assert `$status` and `$output` explicitly:

```bash
@test "list: exits 0" {
  run bash "$SCRIPT" list
  (( status == 0 ))
}

@test "create: rejects invalid entry name" {
  run bash "$SCRIPT" create bad_name
  (( status != 0 ))
  [[ "$output" =~ "invalid entry name" ]]
}
```

Always assert both `$status` and relevant `$output` content. Never assert only one when both are meaningful.

### Mocks

Replace external dependencies with mock executables placed in `test/helpers/`. Add the helpers directory to `PATH` in `setup` or use `env PATH=... bash ...` inline for tests that need to control PATH precisely:

```bash
@test "run: uses selector when no entry argument is given" {
  local tmpbin="$BATS_TEST_TMPDIR/bin"
  mkdir -p "$tmpbin"
  ln -s "$REPO_ROOT/test/helpers/mock_selector" "$tmpbin/selector"
  run env "PATH=$tmpbin:$PATH" "MOCK_OUTPUT=myentry" bash "$SCRIPT" run myentry
  (( status == 0 ))
  [[ "$output" == "ok" ]]
}
```

Mock executables must have a header comment describing what they replace and what environment variables control their behaviour.

### Running tests

Run the full test suite via the Makefile:

```makefile
.PHONY: test
test: ## Run bats test suite
	test/extern/bats/bin/bats test/
```

## What to Test

- Every subcommand or public function has at least one happy-path test
- Every error path has a test that asserts the non-zero exit status and relevant error message content
- Edge cases that are explicitly handled in the code (CRLF input, symlinks, spaces in paths, injection attempts) each have a dedicated test
- Tests for sourced files verify that variables are set and unset correctly in the current shell, not just that commands exit 0
