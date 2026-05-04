# GitHub Badges

## Contents

- [Design principles](#design-principles) — style=flat, grouping, no click-throughs
- [Go projects](#go-projects) — CI, release, language/quality badge groups
- [Python projects](#python-projects) — PyPI and licence badges
- [Pages / static sites](#pages--static-sites) — deploy badge only

## Design Principles

- `style=flat` on all shields.io badges - no other style variants
- Group badges semantically: CI → release → language/quality
- Separate groups with a blank line so GitHub renders a visual break
- Pairs within a group sit on the same line, separated by a single space
- No click-through links on badges unless the badge is for a package registry (PyPI, etc.)
- Code coverage is a static badge - update it manually on each release
- Both CI badges point to `tag.yml` - this reflects the last released build, not every push

## Go Projects

```markdown
![Build Status](https://img.shields.io/github/actions/workflow/status/{owner}/{repo}/tag.yml?style=flat) ![Test Status](https://img.shields.io/github/actions/workflow/status/{owner}/{repo}/tag.yml?style=flat&label=test)

![Release Version](https://img.shields.io/github/v/release/{owner}/{repo}?style=flat) ![Release downloads](https://img.shields.io/github/downloads/{owner}/{repo}/total?label=downloads)

![Go Version](https://img.shields.io/github/go-mod/go-version/{owner}/{repo}) ![Code Coverage](https://img.shields.io/badge/coverage-XX%25-blue)
```

Groups:
1. **CI** - build and test, both sourced from `tag.yml`
2. **Release** - latest version tag and total download count
3. **Language/quality** - Go version from `go.mod`, manually-maintained coverage percentage

Replace `XX` in the coverage badge with the actual percentage on each release.

## Python Projects

```markdown
[![Python versions](https://img.shields.io/pypi/pyversions/{package})](https://pypi.org/project/{package}/) [![License](https://img.shields.io/github/license/{owner}/{repo})](LICENSE)
```

- Use click-through links: Python versions badge links to PyPI, License badge links to the local `LICENSE` file
- If the project is not on PyPI, replace the Python versions badge with a static badge: `![Python](https://img.shields.io/badge/python-3.x%2B-blue)`

## Pages / Static Sites

```markdown
![Deploy](https://img.shields.io/github/actions/workflow/status/{owner}/{repo}/deploy.yml?style=flat)
```

- Single badge pointing to the deploy workflow
- No release or language badges needed
