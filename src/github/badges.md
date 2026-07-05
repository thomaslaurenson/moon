# Badge Conventions

Language-agnostic badge principles. The per-language badge block lives in the relevant language fragment.

- `style=flat` on all shields.io badges; no other variants.
- `logo=github` on CI and release badges; the language logo (`go`, `python`, `cplusplus`) on language/quality badges.
- Group semantically: CI, then release, then language/quality. Separate groups with a blank line; pairs within a group share a line separated by a single space.
- No click-through links except package-registry badges (PyPI, etc.).
- Coverage is a static badge, updated manually on each release.
- Both CI badges point to `tag.yml`, reflecting the last released build.

The CI and release rows are identical for every project. Use them as the first two groups, then append the language/quality row from the language fragment:

```markdown
![Build Status](https://img.shields.io/github/actions/workflow/status/{owner}/{repo}/tag.yml?style=flat&logo=github) ![Test Status](https://img.shields.io/github/actions/workflow/status/{owner}/{repo}/tag.yml?style=flat&label=test&logo=github)

![Release Version](https://img.shields.io/github/v/release/{owner}/{repo}?style=flat&logo=github) ![Release downloads](https://img.shields.io/github/downloads/{owner}/{repo}/total?label=downloads&logo=github)
```
