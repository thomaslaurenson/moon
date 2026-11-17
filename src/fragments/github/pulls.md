# Pull requests

How a pushed branch becomes one commit on `main`, and how that commit becomes a release. Branch names and commit messages are in the git fragment.

## Title

Format `Type/short description`: the branch name with the type capitalised and the hyphens turned to spaces. The type vocabulary is the branch's, in the git fragment. GitHub adds the pull request number itself; the title does not carry it.

```text
Feature/add install script
Fix/local deployment
Refactor/cpp cleanup
```

## Body

Empty. The title says what the change is, the diff says how, and the changelog says what a user will notice; a body would be a fourth account of the same change and the first to go stale. No summary, no test plan, no generated-by line, no co-author line.

```text
git push -u origin <type>/<short-description>
gh pr create --title "<Type>/<short description>" --body ""
```

Pass `--body ""` explicitly, and never `--fill`, which copies the branch's commit messages into the body.

## Checks

Merge only on green, and wait for every check rather than merging on the first few:

```text
gh pr checks <number> --watch --fail-fast
```

A red check is fixed on the branch with another commit, never merged past.

## Merge

Squash, with the commit message given explicitly and the branch deleted:

```text
gh pr merge <number> --squash --subject "<commit message>" --body "" --delete-branch
git fetch --prune
```

- The squash commit is the only one that reaches `main`; the branch's own commits go with the branch. `--subject` is therefore the commit message that lasts, and it follows the git fragment's rules: one line, past tense, no trailing period, the changes comma-separated where there are several.
- Both flags are needed. Without `--subject`, GitHub uses the pull request title and number (`Update/go 1.27 (#18)`), which is not a commit message. Without `--body ""`, it writes a bullet list of every commit on the branch, with a `Co-authored-by` trailer for anyone else who committed to it.
- `--delete-branch` deletes the branch on GitHub and locally, then checks out and pulls `main`, but leaves the remote-tracking ref behind; `git fetch --prune` clears it.

## After the merge

Watch the run on the merged commit before building anything on it:

```text
gh run list --branch main --commit "$(git rev-parse HEAD)"
gh run watch <run-id> --exit-status
```

The `main` run is the first to execute what only `main` runs, such as a prerelease build, so a tooling upgrade that passed its pull request can still fail here. The run takes a few seconds to appear after the merge.

## Releasing

A release is a changelog entry merged like any other change, then a tag on the commit that carries it.

- The entry is the last commit on the branch that makes the release, `Added the X.Y.Z changelog entry`; its format is in the changelog fragment. Dependency and toolchain bumps alone are a patch release.
- Tag only a commit whose `main` run is green. The tag is annotated and signed, it names the commit rather than whatever is checked out, and only the tag is pushed:

```text
git tag -s vX.Y.Z -m "Release vX.Y.Z" <sha>
git push origin vX.Y.Z
```

- Watch the tag's run. `--branch` takes the tag name; `--commit` would also match the `main` run on the same commit:

```text
gh run list --branch vX.Y.Z
gh run watch <run-id> --exit-status
```

- Then check what it published rather than trusting a green run: the release notes match the changelog entry, every expected artefact is attached, and where the language fragment names a check to run against a downloaded artefact, it passes.
