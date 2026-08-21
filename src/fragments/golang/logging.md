# Go output and logging

A CLI writes to two streams and they carry different things. Getting the split wrong is what makes a tool awkward to use in a pipeline, and it is the most common output mistake there is.

## The two streams

- **stdout** carries exactly what the user asked for, and nothing else. It is the tool's return value: it will be piped, redirected to a file, and parsed by something that was not written yet.
- **stderr** carries everything else. Progress, warnings, debug output, and errors.

The test is simple. Redirect stdout to a file and the file should contain the answer and nothing more. A progress line, a "wrote 3 bundles" note, or a warning landing in that file means the split is wrong.

A progress indicator belongs on stderr, and should be suppressed when stderr is not a terminal: nobody reads a spinner from a log file.

## Writers, not globals

Packages under `internal/` never touch `os.Stdout` or `os.Stderr`. They take an `io.Writer` and write to it, exactly as they take a context and an error path rather than owning the process.

Cobra supplies the writers, so a command reads them from itself rather than reaching for the globals:

```go
func newShowCmd() *cobra.Command {
    return &cobra.Command{
        Use:  "show <name>",
        Args: cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            return render(cmd.OutOrStdout(), args[0])
        },
    }
}
```

`OutOrStdout` and `ErrOrStderr` honour whatever `SetOut` and `SetErr` were given, so a test captures output into a `bytes.Buffer` with no global state and no temporary file. `fmt.Println` inside a command defeats this: it writes to the real stdout whatever the caller asked for, and no test can see it.

## slog for diagnostics

Diagnostics go through `log/slog` from the standard library. No third-party logging package (see the tooling fragment), and not the older `log` package, which has no levels.

- Use the **text** handler, never JSON. These are terminal tools read by a person, not services feeding a log aggregator.
- The handler writes to **stderr**.
- Build the logger once, in `cmd/`, and pass it down as an ordinary dependency.

```go
func newLogger(w io.Writer, debug bool) *slog.Logger {
    level := slog.LevelWarn
    if debug {
        level = slog.LevelDebug
    }
    return slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: level}))
}
```

Do not call `slog.Info` and friends, or `slog.Default`, from `internal/`. Those read a package-level logger, which is the process-global state the style fragment rules out, and it makes a test's output depend on what some other test configured.

Pass attributes rather than interpolating them:

```go
// Good: the fields stay separate and quoting is handled
logger.Debug("expanded bundle", slog.String("name", name), slog.Int("fragments", n))

// Bad
logger.Debug(fmt.Sprintf("expanded bundle %s with %d fragments", name, n))
```

## The debug flag

Every CLI has a persistent `--debug` flag on the root command.

- Default level is **Warn**. A successful run prints nothing to stderr, which is what makes the tool usable in a script.
- `--debug` sets the level to **Debug**.

Two levels are what a user can actually select, so write for two. Anything that is not a warning is Debug: if it were something the user needed in normal operation, it would be part of the answer and belong on stdout.

Warn means the command carried on but something was off. It is not a substitute for returning an error.

## Never log and return

One failure, one report. If a function returns an error, the top level prints it (see the errors fragment); logging it as well shows the user the same problem twice, at two levels of detail, and the logged copy is usually the poorer one because it lacks the context that wrapping added on the way up.

Log about a failure only where you **handle** it and do not propagate it. A cleanup error you have deliberately decided to swallow is the clearest case:

```go
if err := os.Remove(tmp); err != nil {
    logger.Warn("could not remove temporary file", slog.String("path", tmp), slog.Any("error", err))
}
```

That is a warning precisely because nothing above will ever hear about it.
