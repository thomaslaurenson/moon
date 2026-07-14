# Markdown style

Markdown-specific style for all `.md` files. Assumes the core conventions (characters, spelling, comment rules).

## Paragraphs

- Write each paragraph as a single unbroken line in the source. Never insert hard line breaks mid-paragraph.
- Separate paragraphs with a single blank line. Never use two or more consecutive blank lines.

## Headings

- ATX-style only (`#`, `##`, `###`). Never Setext underlines (`====`, `----`).
- One blank line before and after each heading.
- Never skip heading levels (no `##` straight to `####`).
- Sentence case only; no title case.

## Lists

- Use `-` for unordered items. Never `*` or `+`.
- No blank lines between items in a tight list; add them only when items contain multiple paragraphs.
- Indent nested lists by 2 spaces.

## Code blocks

- Always fenced (```), never indented. Always include a language hint; use `text` when none applies.
- No line-length limit inside code blocks.
- For omitted or placeholder code use `...` (three ASCII dots).

## Tables

- Always include a header row.
- Use `|---|---|` separators with no padding or column alignment.

## Horizontal rules and inline formatting

- Do not use `---` as a section separator; use headings.
- `**bold**` for key terms or warnings (not `__bold__`); `*italic*` for light stress (not `_italic_`).
- Backticks for all code, command names, file paths, flags, and environment variables inline in prose.
