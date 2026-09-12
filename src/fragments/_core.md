# Core conventions

These rules apply to every file, every language, every task.

These instructions are assembled with shared fragments before the language fragments that refine them. Where two sections conflict, the more specific one wins: a language fragment that narrows or overrides a general CI or tooling rule takes precedence over the shared rule it refines.

## Characters

- No non-ASCII characters in code, comments, or prose. The usual offenders: an em dash (U+2014) or en dash (U+2013) becomes a comma, parentheses, or two sentences; smart quotes become straight `"` and `'`; arrows become `->` and `<-`; tick and cross marks become `yes` and `no`. A plain ASCII hyphen-minus is fine. Exception: user-facing localised string values (translations shown to end users) use the target language's correct characters; the ban is on non-ASCII in the code and prose around them, not on the translated text itself.
- No decorative dividers in code or comments (`// ---`, `// ===`, `# ---`, `# ===`, or similar). Delete them.

## Comments

A comment has to earn its place, and the test is usefulness rather than length. A comment that saves a reader real work is worth twenty lines; one that restates the code is not worth one.

- **Comment what the code cannot say.** If a competent reader of the language can get it from the code in a few seconds, delete it. Step-narration is the common case and is always noise: `# Open the file`, `# Loop through results`.
- **Describe the present.** State how the code behaves now, never how it used to behave or why it changed. Version control holds the history and is better at it. The one exception is behaviour still reachable today: a compatibility shim or a deprecated path is present behaviour, not history.
- **Pre-empt the plausible wrong fix.** Where code deliberately rejects a simpler approach a reader would reasonably try, name the approach and say why it fails ("match the field exactly rather than a substring grep"). This is the highest-value comment there is, and the one most often deleted as "too long" by someone who has not yet made the mistake it prevents. The test is whether a reader might otherwise try it: the same contrast against the code's own past ("this used to use a substring grep") is history, and goes.
- Single-line comments start with the comment character and a single space, capitalise the first word, and take no trailing full stop. A comment of multiple sentences uses normal punctuation; continuation lines need not be capitalised.
- Do not inject `TODO` or `FIXME` comments unless they refer to a real, known issue.

Doc comments are a different job. They are the interface, rendered by tooling, so restating a signature is their purpose rather than a violation. See the language fragment for the applicable convention.

## Spelling

- Write all natural-language prose, comments, and documentation in British English (initialise, colour, optimise, centre).
- Retain American English only for proper nouns, trademarks, brand names, and established technical identifiers: database schemas and fields, code variables, APIs, and third-party libraries where American spelling is already established.
