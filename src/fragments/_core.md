# Core conventions

These rules apply to every file, every language, every task.

These instructions are assembled from general fragments followed by more specific ones. Where a later, more specific section conflicts with an earlier, general one, the later section wins. For example, a language fragment that narrows or overrides a general CI or tooling rule takes precedence over the shared rule it refines.

## Characters

- No em dash (U+2014) or en dash (U+2013). Rewrite using a comma, parentheses, or two sentences. A plain ASCII hyphen-minus is fine.
- No smart or curly quotes. Always use straight ASCII quotes (`"` `'`).
- No non-ASCII characters in code, comments, or prose. This includes Unicode arrows (use `->` and `<-`), tick and cross marks (use `yes`/`no`), en dashes, and decorative glyphs. Use plain ASCII or rewrite the sentence. Exception: user-facing localised string values (translations shown to end users) use the target language's correct characters; the ban is on non-ASCII in the code and prose around them, not on the translated text itself.
- No decorative dividers in code or comments (`// ---`, `// ===`, `# ---`, `# ===`, or similar). Delete them.

## Comments

A comment has to earn its place, and the test is usefulness rather than length. A comment that saves a reader real work is worth twenty lines; one that restates the code is not worth one.

- **Comment what the code cannot say.** If a competent reader of the language can get it from the code in a few seconds, delete it. Step-narration is the common case and is always noise: `# Open the file`, `# Loop through results`.
- **Describe the present.** State how the code behaves now, never how it used to behave or why it changed. Version control holds the history and is better at it. The one exception is behaviour still reachable today: a compatibility shim or a deprecated path is present behaviour, not history.
- **Pre-empt the plausible wrong fix.** Where code deliberately rejects a simpler approach a reader would reasonably try, name the approach and say why it fails. This is the highest-value comment there is, and the one most often deleted as "too long" by someone who has not yet made the mistake it prevents.
- Contrasting with an alternative is useful when the alternative is hypothetical ("match the field exactly rather than a substring grep") and noise when it is the code's own past ("this used to use a substring grep"). The wording barely differs; the test is whether a reader might otherwise try it.
- Single-line comments start with the comment character and a single space, capitalise the first word, and take no trailing full stop. A comment of multiple sentences uses normal punctuation; continuation lines need not be capitalised.
- Do not inject `TODO` or `FIXME` comments unless they refer to a real, known issue.

Doc comments are a different job. They are the interface, rendered by tooling, so restating a signature is their purpose rather than a violation. See the language fragment for the applicable convention.

## Spelling

- Write all natural-language prose, comments, and documentation in British English (initialise, colour, optimise, centre).
- Retain American English only for proper nouns, trademarks, brand names, and established technical identifiers: database schemas and fields, code variables, APIs, and third-party libraries where American spelling is already established.

## Output markers

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

A marker says what kind of message it is, never which stream it goes to. Those are separate decisions, and the language fragment governs the stream.

### When markers do not apply

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
