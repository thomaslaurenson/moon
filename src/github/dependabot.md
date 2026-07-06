# Dependabot

File `.github/dependabot.yml`. Replace `<username>` with the repo owner and `<ecosystem>` with the language ecosystem (`gomod` for Go, `pip` for Python).

```yaml
version: 2
updates:
  - package-ecosystem: github-actions
    directory: /
    schedule:
      interval: weekly
    assignees:
      - "<username>"

  - package-ecosystem: <ecosystem>
    directory: /
    schedule:
      interval: weekly
    labels:
      - "dependencies"
    groups:
      dev-dependencies:
        dependency-type: "development"
    assignees:
      - "<username>"
```

- GitHub Actions: weekly bumps. Language packages: weekly, dev dependencies grouped into a single PR.

## Security-only alternative

When the project wants security updates but no routine version-bump PRs, set `open-pull-requests-limit: 0` on that ecosystem's entry:

```yaml
  - package-ecosystem: <ecosystem>
    directory: /
    schedule:
      interval: weekly
    open-pull-requests-limit: 0
    assignees:
      - "<username>"
```

This only suppresses routine version-update PRs from this file. GitHub's separate Dependabot security-updates feature (a repo-level setting, independent of this file) still opens PRs for vulnerable dependencies regardless of this limit. One ecosystem entry covers both prod and dev dependencies in this mode, so no `dev-dependencies` group is needed.
