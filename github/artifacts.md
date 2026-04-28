# Clean LLM Artifacts — Copilot Instruction File

> **Usage**: Paste the contents of the relevant section below into your
> `.github/copilot-instructions.md`, VS Code user instructions, or any
> Copilot/AI chat context window before pasting code to be cleaned.

---

## Global Rules for Retroactive Code Cleanup

When asked to clean, refactor, or remove AI artifacts from an existing file, you must strictly adhere to the following operational constraints:

- **Full-Fidelity Output**: When returning cleaned code, output the entire requested block or file. Never use placeholders like `// ... rest of code remains the same ...`, `pass`, or truncate the output in any way. 
- **Strict Scope Constriction**: Do not alter any underlying business logic, variable names, function signatures, or architectural patterns. Your only job is to remove AI artifacts, normalize comments, and fix punctuation. The code must execute exactly as it did before.
- **Whitespace Normalization**: After removing block comments, decorative dividers, or inline comments, ensure the resulting code does not have excessive blank lines. Condense multiple blank lines into a single blank line to maintain standard formatting.
- **Protect Human Comments**: You must distinguish between human context and AI narration. 
  - *Delete* AI comments that merely narrate syntax (e.g., `// Initialize counter to 0`, `// Check if user is null`, `// Print to console`).
  - *Strictly Preserve* human comments that explain the "why", business logic, or edge cases (e.g., `// Initialize to 0 because the legacy API expects a 0-indexed fallback`).

---

## General Text Cleanup (Prose, Comments, Docs)

When editing or rewriting textual content, comments, or docstrings, apply these cleanup rules:

- **The "No Dash" Strategy**: 
  - *In Prose/Comments:* Never use em dashes (`—`) or en dashes (`–`). If a comment or docstring contains them, rewrite the sentence naturally using commas, parentheses, or break it into two separate, concise sentences. Do not use standard hyphens (`-`) as makeshift separators in paragraph text.
  - *In Code:* You must absolutely preserve all standard hyphens (`-`) used in actual code syntax (such as minus signs, decrements, pointers, or command-line flags).
- **Standard Formatting**: Remove any Unicode "smart" quotes (`“` `”` `‘` `’`) and replace with straight ASCII quotes (`"` `'`).
- **Eliminate AI Filler**: Remove conversational filler phrases typical of AI-generated text, such as:
  - "It's worth noting that..."
  - "It's important to remember..."
  - "In conclusion..." / "To summarize..."
  - "Certainly!" / "Absolutely!" / "Of course!"
  - "I hope this helps"
  - "Feel free to..."
- **Direct Tone**: Write code and documentation as final, production-ready output. Avoid unnecessarily formal or ornate sentence structures. Prefer direct, plain language. Do not use bullet points where a simple sentence or two would do.

---

## Anti-AI Code Quirks (All Languages)

Aggressively remove the following common AI coding artifacts:

- **Decorative Dividers**: Delete all ASCII dividers of any kind (`---`, `===`, `***`, `###` repeated).
- **Over-Explaining Standard Libraries**: Remove comments that explain basic built-in functions.
- **Defensive Programming Overkill**: Remove unnecessary `try/catch` blocks or overly aggressive null checks if they are not idiomatic to the surrounding codebase or explicitly requested.
- **Step Narration**: Delete "step" comments that just describe the next line (`// Step 1: Open file`).
- **Trailing Block Comments**: Remove closing brace comments (`// end if`, `// end for`) unless the block is genuinely very long (>50 lines).
- **Placeholder Clutter**: Write code without injecting redundant `TODO` or `FIXME` comments.

---

## Language-Specific Cleanup

### C / C++
- Remove all decorative block comment dividers (e.g., `// -----------------------`, `// =======================`).
- Keep standard Doxygen formatting `/** ... */` if it exists, but remove redundant structural boilerplate inside them.

### Python
- Convert heavy block comments (`# -----------------------`) to standard docstrings `""" ... """` if they describe a function/class, or delete them if they are just visual clutter.
- Do not add conversational comments like `# Here we...` or `# Next, we will...`.
- Do not over-annotate type hints with redundant inline comments.

---

## Quick Regex Reference (for find-and-replace in VS Code)

| Pattern | What it catches | Replace with |
|---|---|---|
| `—` | Em dash (U+2014) | *(Rewrite manually or use `,`)* |
| `–` | En dash (U+2013) | *(Rewrite manually or use `,`)* |
| `\u2018\|\u2019` | Smart single quotes | `'` |
| `\u201C\|\u201D` | Smart double quotes | `"` |
| `\/\/ [-=*#]{4,}` | Decorative C++ dividers | *(delete line)* |
| `# [-=*#]{4,}` | Decorative Python dividers | *(delete line)* |

*In VS Code, open **Find & Replace** (`Ctrl+H` / `Cmd+H`), enable **regex mode** (`Alt+R`), and use the patterns above.*

---

## Tip: Persistent Copilot Instructions

To apply these rules automatically to all Copilot suggestions, save the relevant sections of this document into a `.github/copilot-instructions.md` file at the root of your repository. Copilot Chat will automatically ingest these rules and use them as context for every prompt.
