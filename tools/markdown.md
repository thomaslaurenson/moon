# Markdown Style Guide

Style conventions for all Markdown files in this repo.

## Characters

- No em dash (`—`). Rewrite using a comma, parentheses, or two sentences.
- No smart or curly quotes (`"` `"` `'` `'`). Always use straight ASCII quotes (`"` `'`).
- No non-ASCII characters in prose, code, or headings. This includes Unicode arrows (`→`, `←`), tick marks (`✓`, `✗`), en dashes (`–`), and any other non-ASCII glyph. Use plain ASCII equivalents or rewrite the sentence.
- Use British English spellings. `Initialise` not `Initialize`. `Colour` not `Color`.

## Paragraphs

- Write each paragraph as a single unbroken line in the source file. Never insert hard line breaks mid-paragraph.
- Use a single blank line to separate paragraphs. Never use two or more consecutive blank lines.

## Headings

- Use ATX-style headings only (`#`, `##`, `###`). Never use Setext-style underlines (`====` or `----`).
- Always place one blank line before a heading and one blank line after it.
- Never skip heading levels. Do not jump from `##` to `####`.
- Sentence case only: capitalise the first word and proper nouns. No title case.

## Lists

- Use `-` for unordered list items. Never use `*` or `+`.
- No blank lines between items in a tight list. Add a blank line between items only when the items themselves contain multiple paragraphs.
- Indent nested lists by 2 spaces.

## Code Blocks

- Always use fenced code blocks (` ``` `). Never use indented code blocks.
- Always include a language hint on the opening fence. For blocks with no applicable language, use `text`.
- No line-length limit inside code blocks. Wrap only when the language itself requires it.
- When showing omitted or placeholder code, use `...` (three ASCII dots). Never use `…` (U+2026).

## Tables

- Always include a header row.
- Use `|---|---|` separators with no padding or column alignment. Do not pad columns with extra spaces to align values.

## Horizontal Rules

- Do not use `---` as a section separator. Use headings to separate sections.

## Inline Formatting

- Use `**bold**` for emphasis on key terms or warnings. Do not use `__bold__`.
- Use `*italic*` for light stress or foreign terms. Do not use `_italic_`.
- Use backticks for all code, command names, file paths, flag names, and environment variables inline in prose.
