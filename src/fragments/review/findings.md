# Review findings

How to run a review and report what it found. The other half of the job, turning those findings into commits, is the review-fixes bundle.

## Agree the scope first

Two things are settled before any code is read: what is in scope, and what it is being judged against.

- Scope is a set of files, a diff, a package, or a whole repository. Name it back before starting, because "review this" and "review what I changed" are different jobs with different answers.
- The standard is the project's own spec, not the reviewer's taste. Load the language's code-authoring bundle (`go-cli-code`, `python-lib-code`, `cpp-app-code`) and review against that. Where the project has a full-project bundle, the build files, CI workflows and release config are in scope against it too.
- Prose, comments and documentation are reviewed against the core conventions and the markdown style in this bundle, on the same footing as the code.
- Where no standard is named, say so and review against the ordinary practice of the language. Inventing house rules on the spot produces findings the user never agreed to be held to.

## A review writes nothing

A review is read-only. Not one file is written, for any reason.

The tempting exceptions are why this is worth stating: a typo in a comment, a stray blank line, a one-line fix that takes less time than writing it up. Fix any of them and the diff carries a change the user never saw in a list, buried in the reading they actually asked for.

- Formatters and linters run in check mode only (`gofmt -l`, `make check_format`, `go vet`). Never in write mode.
- Building and running the test suite is reading. Do it, and prefer it to guessing.
- The findings table is the entire output of a review.

## Verify before reporting

Run the check rather than predicting it. A finding that says a file will not compile, when it compiles, costs more trust than the finding was ever worth, and it puts every other finding in the list in doubt.

Where a claim can be settled by a command, settle it and report what the command said. Where it cannot, say that it was reasoned about rather than run.

## The findings table

The output is a plan, not an essay: a short phase key, then one table.

```markdown
### Phases

1. Entry point and wiring, which everything below depends on
2. Boundary: writers, environment, process ownership
3. Independent: build, CI, docs

| # | Phase | Severity | Finding | Location |
|---|---|---|---|---|
| 1 | 1 | high | Package-level Execute wrapper blocks writer injection | main.go:15 |
| 2 | 2 | high | Mutable package streams and a ResetForTesting helper | internal/ui/ui.go:29 |
| 3 | 3 | low | Both CI badges point at the same workflow | README.md:3 |
```

- Number every finding. The fix loop cites them by number, so the numbers are the interface between the two halves.
- The finding cell is ten words at most, and names the fault rather than the remedy.
- Severity is high, medium or low. Three levels only: a finer scale invites argument about the labels without telling the reader anything more.
- Order phases so dependent work follows what it depends on, and put everything independent in the last one. Phase is the primary sort and severity breaks ties within it, so dependency ordering always wins.
- Cite `file:line`. Where a finding has no single location, name the area instead.

## Hard limits

The table is the report. These are limits rather than targets, and going over is a failure even when every line of it is true:

- No prose paragraphs before the phase key or between rows.
- No code blocks, no diffs, no proposed fixes. The remedy belongs to the fix loop's turn, not this one.
- A closing note of two lines at most: what was verified, and anything notable that passed.
- Deliberate deviations go in a second table, dropping the severity column, so code that departs from the standard on purpose is not ranked as a fault.

Everything held back is available the moment it is asked for. A reader who wants the reasoning asks for the finding by number.
