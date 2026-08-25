# Working through review findings

How an agreed list of changes becomes commits. Written for review findings, and it applies unchanged to any agreed list: a plan, a backlog, a set of review comments.

## Start from the findings table

The input is the table the review produced: numbered findings grouped into phases, in the shape the review-findings bundle specifies. The phase key is the order. Confirm it and work to it rather than deriving a new one, because the phases already encode which findings depend on which.

- Work phases in sequence, and rows in order within a phase.
- Open every turn by citing the finding's number, so the user can match it against the table without hunting for it.
- Where there is no table, agree an order once before starting and then treat it exactly as a phase key.
- Close each handover with what is left, as a table in the review's own columns, so the user reads one shape throughout. Drop completed rows rather than marking them done; the git log already records those.
- Say so when a phase completes. That is the point at which reordering what remains is cheap, and the last chance to do it before the next phase depends on it.

## One at a time

Take one finding per turn. Never batch several into a single proposal, and never implement ahead of the discussion.

Batching costs twice. The user is asked to react to several unrelated proposals at once, so the feedback arrives partial and it is unclear which part it applies to; and the resulting commit cannot be reverted one finding at a time, which is the whole reason for reviewing them separately.

## The shape of a turn

Three parts, then stop:

1. The issue, in one or two sentences.
2. The proposed solution, in no more than five lines.
3. A request for feedback, in one line.

These are hard limits, not targets to aim near. A proposal turn that runs over has failed the rule even when every line in it is true and useful.

What is banned from a proposal turn, whatever the temptation:

- Code blocks, diffs and signatures. Code belongs in the implementation, not the pitch. The commit handover is the one turn that requires a block; see below.
- The reasoning behind the proposal, the alternatives weighed, the trade-offs considered.
- Restating the finding, the standard it came from, or why the standard says it.
- Listing what is deliberately out of scope, or what a later finding will cover.
- Naming every affected file or call site. A count is enough.

All of it is available the moment the user asks for it, and none of it belongs in the opening turn. The user is deciding whether to approve a direction, and the material that led to the direction buries the direction itself. Ten findings at this length is a document nobody reads; ten findings at three lines each is a conversation.

Assume the user knows the codebase and has read the review. Write for someone who needs reminding which finding this is, not teaching what it means.

## Asking

Ask an open question in prose. No numbered options, no multiple choice, no option-selection interface.

An option list quietly narrows the answer to whatever was thought of while writing it, and the useful reply to a proposal is often none of the offered choices. Asking in prose leaves room for "yes but not that way", which is the answer that improves the work.

## Implementing

Implement the finding under discussion and nothing else. Anything noticed along the way goes back on the list as a new finding; it does not get fixed in passing.

An unrelated change in the diff is one the user never agreed to, it is invisible in a commit named after something else, and it is reliably the one that breaks something later.

## Verify before handing over

The project's aggregate check target passes before any commit command is offered. Where the project has one, that is `make ci`; otherwise it is the tests and the static checks it does have.

Report what was run and what it said. Do not assert that everything is fine. If the checks cannot be made to pass, say so, say what failed, and stop there rather than handing over a commit that buries a failure.

## Committing

The user runs git. The agent writes the commands and prints them:

```sh
git add <paths>
git commit -m "Fixed the thing that was wrong"
```

- One commit per finding.
- The message follows the project's commit conventions. Where a commit conventions fragment is in scope it governs; otherwise match the existing `git log`.
- Name the paths explicitly. Never `git add -A` or `git add .`, which stage whatever else happened to be sitting in the tree.
- Never run a command that writes history or reaches a remote: `add`, `commit`, `push`, `tag`, `reset`, `rebase`. Read-only inspection is how the agent checks its own work, so `status`, `diff` and `log` stay free.

Then the next finding, and the same loop again.
