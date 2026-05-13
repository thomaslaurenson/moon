# GitHub Badges

## Design Principles

- `style=flat` on all shields.io badges - no other style variants
- `logo=github` on all CI and release badges; use the language logo (`go`, `cplusplus`, `python`) on language/quality badges
- Group badges semantically: CI → release → language/quality
- Separate groups with a blank line so GitHub renders a visual break
- Pairs within a group sit on the same line, separated by a single space
- No click-through links on badges unless the badge is for a package registry (PyPI, etc.)
- Code coverage is a static badge - update it manually on each release
- Both CI badges point to `tag.yml` - this reflects the last released build, not every push

## Go Projects

```markdown
![Build Status](https://img.shields.io/github/actions/workflow/status/{owner}/{repo}/tag.yml?style=flat&logo=github) ![Test Status](https://img.shields.io/github/actions/workflow/status/{owner}/{repo}/tag.yml?style=flat&label=test&logo=github)

![Release Version](https://img.shields.io/github/v/release/{owner}/{repo}?style=flat&logo=github) ![Release downloads](https://img.shields.io/github/downloads/{owner}/{repo}/total?label=downloads&logo=github)

![Go Version](https://img.shields.io/github/go-mod/go-version/{owner}/{repo}?logo=go) ![Code Coverage](https://img.shields.io/badge/Coverage-XX%25-blue?logo=go)
```

Groups:
1. **CI** - build and test, both sourced from `tag.yml`
2. **Release** - latest version tag and total download count
3. **Language/quality** - Go version from `go.mod`, manually-maintained coverage percentage

Replace `XX` in the coverage badge with the actual percentage on each release.

## Python Projects

```markdown
![Build Status](https://img.shields.io/github/actions/workflow/status/{owner}/{repo}/tag.yml?style=flat&logo=github) ![Test Status](https://img.shields.io/github/actions/workflow/status/{owner}/{repo}/tag.yml?style=flat&label=test&logo=github)

![Release Version](https://img.shields.io/github/v/release/{owner}/{repo}?style=flat&logo=github) ![Release downloads](https://img.shields.io/github/downloads/{owner}/{repo}/total?label=downloads&logo=github)

![Python Version](https://img.shields.io/pypi/pyversions/{package}?logo=python) ![Code Coverage](https://img.shields.io/badge/Coverage-XX%25-blue?logo=python)
```

Groups:
1. **CI** - build and test, both sourced from `tag.yml`
2. **Release** - latest version tag and total download count
3. **Language/quality** - Python version from PyPI, manually-maintained coverage percentage

Notes:
- If the project is not on PyPI, replace the Python version badge with a static badge: `![Python Version](https://img.shields.io/badge/python-3.x%2B-blue?logo=python)`
- Replace `XX` in the coverage badge with the actual percentage on each release.

## C++ Projects

```markdown
![Build Status](https://img.shields.io/github/actions/workflow/status/{owner}/{repo}/tag.yml?style=flat&logo=github) ![Test Status](https://img.shields.io/github/actions/workflow/status/{owner}/{repo}/tag.yml?style=flat&label=test&logo=github)

![Release Version](https://img.shields.io/github/v/release/{owner}/{repo}?style=flat&logo=github) ![Release downloads](https://img.shields.io/github/downloads/{owner}/{repo}/total?label=downloads&logo=github)

![C++ Version](https://img.shields.io/badge/Version-XX-blue?logo=cplusplus) ![Code Coverage](https://img.shields.io/badge/Coverage-XX%25-blue?logo=cplusplus)
```

Groups:
1. **CI** - build and test, both sourced from `tag.yml`
2. **Release** - latest version tag and total download count
3. **Language/quality** - C++ version from root `CMakeLists.txt`, manually-maintained coverage percentage

Notes:
- Replace `XX` in the version badge with the actual version from root `CMakeLists.txt`
- Replace `XX` in the coverage badge with the actual percentage on each release.

## Pages / Static Sites

```markdown
![Deploy](https://img.shields.io/github/actions/workflow/status/{owner}/{repo}/deploy.yml?style=flat)
```

- Single badge pointing to the deploy workflow
- No release or language badges needed
