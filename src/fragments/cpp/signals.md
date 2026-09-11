# C++ interrupts

What happens when a user presses Ctrl-C part-way through. Assumes the error handling and output fragments.

Applies to every tier. A library implements "Passing it down" and defines the `Interrupted` type; the handler, the re-raise and the Windows notes belong to the tier with an `app/` binary.

Without handling, an interrupt kills the process where it stands. Destructors do not run, so a half-written archive stays on disk, a temporary directory is never removed, and a file handle is closed by the kernel rather than flushed. The point of handling a signal is not politeness; it is that RAII cannot do its job if the process never unwinds.

## When a command needs it

A command needs interrupt handling when any of these is true:

- It writes a file that is invalid until complete.
- It can run for longer than a user is willing to wait.
- It holds a resource the operating system will not clean up.

A command that reads some input, prints an answer and exits in milliseconds needs none of this. There is nothing to interrupt and nothing to leave behind.

## The flag

A signal handler may do almost nothing. It runs between two arbitrary instructions, possibly on another thread, so allocating, locking, or writing to a stream from inside one is undefined behaviour. In particular it must never print: `std::cout` takes a lock, and taking a lock the interrupted code already holds deadlocks the process at the moment the user asked it to stop.

So the handler sets a flag and returns. Everything else happens in ordinary code that polls it:

```cpp
// app/main.cpp
#include <atomic>
#include <csignal>

namespace {

std::atomic<bool> interrupted{false};
static_assert(decltype(interrupted)::is_always_lock_free,
              "a signal handler may only touch a lock-free atomic");

extern "C" void OnInterrupt(int) { interrupted.store(true, std::memory_order_relaxed); }

} // namespace
```

The `static_assert` is what makes this defined rather than merely conventional. Writing an atomic from a handler is safe only when the type is lock-free; a non-lock-free atomic takes a lock internally and is back to the deadlock above. Every real platform satisfies it for `bool`, so the assertion costs nothing and documents the requirement.

This is a namespace-scope mutable variable, which the style fragment otherwise rules out. It qualifies because a signal handler takes no arguments and has nowhere else to write: there is no version of this that threads state through, which is exactly the test that section applies.

Install the handler in `main`, before the work starts:

```cpp
std::signal(SIGINT, OnInterrupt);
std::signal(SIGTERM, OnInterrupt);
```

## Passing it down

Library code cannot reach that flag and must not try. It takes the cancellation the same way it takes a stream, as a parameter:

```cpp
void ExtractAll(const Archive &archive, const fs::path &out,
                std::ostream &progress, const std::atomic<bool> &cancelled);
```

Take `const std::atomic<bool> &`, so the caller decides what cancellation means and a test can pass a flag it sets itself. A library that reads a global has the same problem as one that writes to `std::cout`: nothing can exercise it, and a second caller in the same process cannot have different behaviour.

Poll it at the top of the loop that does the work, not inside the innermost operation:

```cpp
for (const auto &entry : archive.Entries()) {
    if (cancelled.load(std::memory_order_relaxed)) {
        throw mylib::Interrupted{};
    }
    Extract(entry, out);
}
```

One check per iteration is the right granularity. A single entry finishes, the loop stops, and the work is left at a boundary rather than mid-write. Checking more often costs more than it buys; checking only before the loop means an interrupt does nothing once it starts.

`Interrupted` derives from the library's root exception like every other failure type, so it unwinds the stack and every destructor runs. That unwinding is the whole mechanism: the temporary file is removed, the archive handle is closed, and the tree is left as it was found.

## Leaving nothing behind

Polling makes the stop clean; it does not make the output correct. A command writing a file that is invalid until complete writes to a temporary path and renames on success:

```cpp
fs::path tmp = target;
tmp += ".partial";
WriteArchive(tmp, entries, cancelled);
fs::rename(tmp, target);
```

A rename on the same filesystem is atomic, so the target either holds the previous contents or the new ones and never a truncated mixture. An interrupt anywhere before the rename leaves only the temporary, which the writer's destructor removes.

## Reporting it

A user who pressed Ctrl-C does not need to be told. Print nothing, and do not report it as an error: it is the outcome they asked for.

Catch it above the other handlers in `main`, then restore the default disposition and re-raise, so the process dies of the signal rather than returning from `main`:

```cpp
} catch (const mylib::Interrupted &) {
    std::signal(SIGINT, SIG_DFL);
    std::raise(SIGINT);
    return 1; // not reached
}
```

Re-raising rather than returning a number is what makes the shell report the interrupt correctly. A shell derives 130 from the process having died of signal 2, and `$?` is only part of what it inspects: a `wait` status showing a normal exit with code 130 is not the same thing, and an interactive shell will not stop a loop for it. The project's own exit codes stay 0, 1 and 2 (see the error handling fragment), because this path never returns one.

## Windows

`std::signal` is supported by the Microsoft runtime, and SIGINT works, but two differences matter.

The handler runs on a **separate thread** for Ctrl-C, rather than interrupting the main one. Setting an atomic is still correct, which is another reason for the flag rather than anything more ambitious.

SIGTERM is never delivered by the system; nothing sends it. Registering it is harmless and keeps one code path, so do that rather than bracketing the call in `#ifdef`.
