# Branch names

Format `type/short-description`, all lowercase kebab-case, 3-5 words, no issue numbers.

| Type | When |
|---|---|
| `feature/` | New capabilities, subcommands, flags |
| `refactor/` | Internal restructuring, no behaviour change |
| `fix/` | Bug fixes |
| `update/` | Dependency bumps or toolchain upgrades |

```
feature/add-install-script
refactor/noun-first-cli-and-layout
fix/local-deployment
```

Pick the type from what the change does to the project, not from how it was written: a bug fix that happens to restructure a package is still `fix/`.

The branch name is where a PR title comes from: `feature/add-install-script` becomes `Feature/add install script (#13)`. Keeping the description short and readable in kebab-case is what makes that rewrite mechanical. See the commit and PR conventions fragment.
