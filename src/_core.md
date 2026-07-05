# Core Conventions

These rules apply to every file, every language, every task.

## Characters

- No em dash (`-`). Rewrite using a comma, parentheses, or two sentences.
- No smart or curly quotes. Always use straight ASCII quotes (`"` `'`).
- No non-ASCII characters in code, comments, or prose. This includes Unicode arrows (use `->` and `<-`), tick and cross marks (use `yes`/`no`), en dashes, and decorative glyphs. Use plain ASCII or rewrite the sentence.
- No decorative dividers in code or comments (`// ---`, `// ===`, `# ---`, `# ===`, or similar). Delete them.

## Comments

- Do not write step-narration comments that restate the next line of code. Bad: `# Open the file`, `# Loop through results`.
- Preserve comments that explain the why (business logic, architecture). Aggressively delete and refactor comments that narrate the what.
- Single-line comments start with the comment character and a single space, capitalise the first word, and take no trailing full stop. A comment of multiple sentences uses normal punctuation; continuation lines need not be capitalised.
- Do not inject `TODO` or `FIXME` comments unless they refer to a real, known issue.

## Spelling

- Write all natural-language prose, comments, and documentation in British English (initialise, colour, optimise, centre).
- Retain American English only for proper nouns, trademarks, brand names, and established technical identifiers: database schemas and fields, code variables, APIs, and third-party libraries where American spelling is already established.
