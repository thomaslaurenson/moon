# Review findings

How to run a review and report what it found. The other half of the job, turning those findings into commits, is the review-fixes bundle.

## Agree the scope first

Two things are settled before any code is read: what is in scope, and what it is being judged against.

- Scope is a set of files, a diff, a package, or a whole repository. Name it back before starting, because "review this" and "review what I changed" are different jobs with different answers.
- The standard is usually a bundle. Load the language's own code-authoring bundle (`go-cli-code`, `python-lib-code`, `cpp-app-code`) and review against that, rather than against whatever the reviewer happens to believe.
- Where no standard is named, say so and review against the ordinary practice of the language. Inventing house rules on the spot produces findings the user never agreed to be held to.

## A review changes nothing

A review produces a list, not a diff. No edits, no formatting, no "while I was in there".

The reason is not caution. A finding the user has not seen cannot have been agreed, so an agent that fixes as it reads hands back a diff nobody chose, mixed in with the reading it was actually asked for. The list is what makes each change a decision.

Running the project's own checks is reading rather than editing, so `go vet`, `make check_format` and the test suite are all fair game. A formatter in write mode is not.

## Verify before reporting

Run the check rather than predicting it. A finding that says a file will not compile, when it compiles, costs more trust than the finding was ever worth, and it puts every other finding in the list in doubt.

Where a claim can be settled by a command, settle it and report what the command said. Where it cannot, say that it was reasoned about rather than run. The distinction matters to the reader deciding which findings to take on faith.

## Reporting

- Number every finding. The fix loop cites them by number, so an unnumbered list makes the next conversation harder than it needs to be.
- Group by area, most significant first. Wiring and correctness before naming and layout.
- Cite `file:line` for anything with a location.
- A finding is two or three sentences: what is wrong, where, and what it costs. Detail is provided when asked for, not pre-emptively.
- Separate deliberate deviations from misses. Code that departs from the standard on purpose, and says why in a comment, has not made a mistake. Report it as a deviation for confirmation rather than as a fault, and say why the deviation looks justified.
- Say briefly what passed. A list containing only failures gives no sense of proportion, and the reader cannot tell a project with six problems from one with six hundred.

## Length

A review written from the files walks the diff and reports everything it sees. A review written from the standard names the gaps between the two and stops. The second is shorter and worth more.
