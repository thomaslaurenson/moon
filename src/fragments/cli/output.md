# Command line output

Markers for the messages a command line program prints for a person to read. Assumes the core conventions.

A command line program that prints messages for a person to read prefixes each one with a marker naming what kind of message it is. The vocabulary is the same in every language, so that somebody moving between tools reads the same signals rather than learning each one separately.

| Marker | Meaning |
|---|---|
| `[*]` | Information, progress, or a result |
| `[+]` | Something was added |
| `[-]` | Something was removed |
| `[~]` | Something was changed in place |
| `[!]` | A warning or an error |
| `[?]` | A question, with an answer expected |

`[+]`, `[-]` and `[~]` describe a change to whatever the command is operating on. `[*]` is everything else, and is the one to reach for when no other fits.

A marker says what kind of message it is, never which stream it goes to. Those are separate decisions, and the language output fragment governs the stream.

## When markers do not apply

The rule covers the running commentary, not the result, and there are more places it stays out of than it goes.

- **Structured output carries no marker.** A table, a list of names, or anything a caller is expected to parse is the program's answer rather than a message about it, and a prefix on those lines corrupts them for whatever reads them next.
- **Only a command line program.** A library prints nothing at all, and a service writes for a log aggregator rather than a person.
- **Only where a person is reading.** Output shaped for another program is not commentary, whatever stream it goes to.

Where none of the markers fits what a line is doing, that is a sign the line is output rather than commentary. Print it bare. A vocabulary applied to everything stops carrying meaning, so it is better for a program to mark some of its lines well than all of them badly.

```text
[*] Reading archive: patch.mpq
[+] Adding file: readme.txt
[!] Skipped unreadable file: locked.dat
[?] Overwrite the existing archive? [y/N]:

NAME        SIZE  MODIFIED
readme.txt  1.2K  2026-01-04
locked.dat   840  2026-01-04
```

The listing at the end is the answer somebody asked for and takes no marker. Everything above it is the program talking about its work, and every line of that says which kind of talking it is.
