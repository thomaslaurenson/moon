# Go functional testing

Asserts the CLI contract: that a flag reaches the code it names, that output lands on the right stream, and that a failure is reported rather than swallowed. It is the same layer as the C++ functional fragment, without the subprocess. Cobra hands the command tree and its writers over as values, so the test runs in-process and needs no compiled binary, no subprocess helper, and no platform handling.

## What it is for

Unit tests prove a function behaves. Functional tests prove the wiring around it is correct, which is a separate question and the one users hit first.

The characteristic bug is invisible to every unit test:

```go
c.Flags().BoolVarP(&long, "long", "l", false, "...")
c.Flags().BoolVar(&asJSON, "json", false, "...")

RunE: func(cmd *cobra.Command, _ []string) error {
    return a.bundleList(cmd.OutOrStdout(), long, asJSON)
},
```

Two adjacent `bool` parameters passed positionally. Swap them and it still compiles, every unit test of `bundleList` still passes because they call it directly, and `--long` starts emitting JSON. Only a test that goes through the command layer can see it.

## Shape

Tests live beside the code they exercise, in the `cmd` package. A helper builds the tree and captures both streams:

```go
func run(t *testing.T, fsys fs.FS, args ...string) (stdout, stderr string, err error) {
    t.Helper()

    var out, errOut bytes.Buffer
    root := NewRootCmd(fsys, &out, &errOut)
    root.SetArgs(args)
    err = root.Execute()

    return out.String(), errOut.String(), err
}
```

Build a fresh tree per case rather than reusing one. Flag values persist on a command after `Execute`, so a shared tree leaks state between subtests and reintroduces the ordering problems the constructor exists to prevent.

## What to test

One table-driven test per command, with a row for each of:

- the default output, with no flags
- **each flag**, asserting it changes the output in the way its name claims
- **the error path**: a missing or invalid argument returns an error and writes nothing to stdout

Not every combination of flags. The point is proving each one is connected, not exploring the matrix.

## Asserting

Assert on properties, not exact bytes. "parses as JSON", "has two columns", "mentions the bundle name" all survive a help-text tweak; a full expected string does not, and turns every wording change into a dozen failures.

Errors are returned, not printed. The root sets `SilenceErrors` (see the scaffolding fragment), so a failing command surfaces through the returned `error` and stderr carries only diagnostics. Assert on `err` for failure, and assert stdout is empty: a command that fails halfway and leaves partial output on stdout has broken the contract that stdout is the answer.

## Placement and coverage

- Files sit in `cmd/`, named for the command they mirror: `cmd/bundle_test.go`, not one file per source file.
- They run under `make test` with no build tag and no build step, so they run on every pull request.
- Coverage still measures `./internal/...` only. `cmd/` stays excluded deliberately: one functional test through a `RunE` covers most of a command file, so counting it would raise the figure without saying anything about how well the logic underneath is tested.
