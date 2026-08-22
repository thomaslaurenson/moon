# Badge conventions

Language-agnostic badge principles. The per-language badge block lives in the relevant language fragment.

- `style=flat` on all shields.io badges; no other variants.
- `logo=github` on CI and release badges; the language logo (`go`, `python`, `cplusplus`) on language/quality badges.
- Group semantically: CI, then release, then language/quality. Separate groups with a blank line; pairs within a group share a line separated by a single space.
- No click-through links except package-registry badges (PyPI, etc.).
- Coverage is a static badge, updated manually on each release.
- The two CI badges point at different workflows: `tag.yml` for the last released build, `main.yml` for the current state of main. Two badges rendering the same workflow can never disagree, which makes one of the labels a lie.
- The dynamic shields.io badges below (`github/actions/workflow/status`, `github/v-release`, `github/downloads`) call GitHub's unauthenticated public API and will not render for a private repository. Private repos use static badges for CI, release version, and downloads instead, updated by hand alongside the coverage badge.

The CI and release rows are identical for every project. Use them as the first two groups, then append the language/quality row from the language fragment:

```markdown
![Release Build](https://img.shields.io/github/actions/workflow/status/{owner}/{repo}/tag.yml?style=flat&label=release&logo=github) ![Main Build](https://img.shields.io/github/actions/workflow/status/{owner}/{repo}/main.yml?style=flat&label=main&logo=github)

![Release Version](https://img.shields.io/github/v/release/{owner}/{repo}?style=flat&logo=github) ![Release downloads](https://img.shields.io/github/downloads/{owner}/{repo}/total?label=downloads&logo=github)
```
