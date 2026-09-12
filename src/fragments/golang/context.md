# Go contexts

A `context.Context` carries one thing that matters here: a signal that the work should stop. It is created at the entry point, passed down the call chain, and every function and goroutine holding it learns at the same moment that the user pressed Ctrl-C or a deadline passed. Without it an interrupt kills the process mid-operation, leaving a half-written file, a mount still held, or an orphaned child process.

## When a command needs one

A command needs a context when any of these is true:

- It can run for longer than a user is willing to wait.
- It starts a goroutine.
- It runs a subprocess or talks to the network.

A command that does none of them does not get a context, and adding one is ceremony rather than correctness. A tool that reads some files, assembles output, and exits in milliseconds has nothing to cancel.

## Creating it

`main` owns the context, because `main` owns the process. Build it from the signal handler so an interrupt becomes ordinary cancellation:

```go
func main() {
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    root := cmd.NewRootCmd(os.Stdout, os.Stderr)
    if err := root.ExecuteContext(ctx); err != nil {
        // See the errors fragment for the rest of this block
    }
}
```

`ExecuteContext` hands the context to cobra, and a command handler takes it back with `cmd.Context()`. Never call `context.Background()` below `main`; a context created halfway down is one nothing can cancel.

## Plumbing

- `ctx` is the first parameter and is named `ctx`: `func Scan(ctx context.Context, root string) error`.
- Never store a context in a struct field. It describes one call, not the lifetime of a value.
- Never pass a nil context. `context.TODO()` exists to mark a call site that has not been plumbed through yet, and is a note to finish the job rather than a permanent answer.
- Do not add `ctx` to a function that cannot act on it. Plumbing that ends in a function ignoring the parameter is noise.

## Using it

Prefer handing the context to something that already understands it over checking it by hand:

```go
c := exec.CommandContext(ctx, "sh", "-c", script)
req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
```

Check it directly only in a loop the standard library cannot see into:

```go
for _, target := range targets {
    if err := ctx.Err(); err != nil {
        return err
    }
    // ... scan target
}
```

Give anything on a network a deadline as well as a cancel, since a hung connection is not an interrupted one:

```go
ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
defer cancel()
```

Always `defer cancel()`. Skipping it leaks the timer until the parent context is done.

## Goroutines

Every goroutine a command starts takes the context and has a defined way to exit. Fire-and-forget is the failure mode: the process ends while the goroutine is mid-write, and nothing waits for it or reports what it was doing.

```go
errs := make(chan error, len(targets))

var wg sync.WaitGroup
for _, t := range targets {
    wg.Add(1)
    go func() {
        defer wg.Done()
        if err := scan(ctx, t); err != nil {
            errs <- err
        }
    }()
}
wg.Wait()
close(errs)

var failures []error
for err := range errs {
    failures = append(failures, err)
}
return errors.Join(failures...)
```

The buffer has to hold every send the loop can make. Size it below that, or leave the channel unbuffered with nothing reading, and the first goroutine to fail blocks on the send, `wg.Done` never runs, and `wg.Wait` never returns: the command hangs rather than reporting. Close after `wg.Wait` so the drain terminates, and `errors.Join` returns nil when nothing failed.

A goroutine with no exit path other than the process ending is a leak, whether or not it is reported as one.

## Cleanup after cancellation

What has to be undone is project-specific: releasing a mount, killing a child, restoring terminal state. What is not project-specific is the trap. Once the context is cancelled, every call taking it fails immediately, so cleanup that needs a context of its own must not reuse the cancelled one:

```go
// The parent is already cancelled, so unmounting with it would fail at once
cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
defer cancel()
return unmount(cleanupCtx, target)
```

Cleanup must be safe to run twice: an interrupt during shutdown is normal, not exceptional.

## Reporting an interrupt

A user who pressed Ctrl-C does not need an error message about it. Do not print `context canceled` as a failure; it is the requested outcome. Exit non-zero without a message, using the shell convention of 128 plus the signal number, so 130 for SIGINT.

Ask the context, not the returned error. A cancelled operation usually reports its own symptom rather than the cancellation: `exec.CommandContext` returns `signal: killed`, and an interrupted read returns whatever it was part-way through. Matching on the error therefore misses the common cases, and the interrupt is reported as a failure with a confusing message. Only the context knows why the work stopped.

The check belongs in `main`, because that is where the cancellable context was created, and it comes before the other exit-code decisions (see the errors fragment):

```go
if err := root.ExecuteContext(ctx); err != nil {
    if errors.Is(ctx.Err(), context.Canceled) {
        os.Exit(130)
    }
    // ... the remaining exit-code decisions
}
```

`signal.NotifyContext` does not report which signal arrived, so 130 covers SIGTERM as well as SIGINT. A command that has to tell them apart keeps a channel-based handler for the distinction (see below).

A deadline is different. `context.DeadlineExceeded` means the work did not finish in the time allowed, which the user did not ask for and should be told about, so it travels up as an ordinary error and `main` prints it. Wrap it where the deadline was set, so the message names the operation that ran out of time instead of reporting a bare `context deadline exceeded`.

## Signals beyond shutdown

`signal.NotifyContext` covers shutdown, which is what most commands need. It cannot express a signal that means something other than stop, so a command handling terminal resize (`SIGWINCH`) or coordinating a process group keeps a channel-based `signal.Notify` for those, alongside the context for shutdown. The two are complementary rather than alternatives.
