# Dependabot

## Config

File: `.github/dependabot.yml`

```yaml
version: 2
updates:
  - package-ecosystem: github-actions
    directory: /
    schedule:
      interval: weekly
    assignees:
      - "<username>"

  - package-ecosystem: <gomod|pip>
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

Replace `<username>` with the repo owner's GitHub handle. Replace `<gomod|pip>` with
the appropriate package ecosystem for the project.

## Rules

- GitHub Actions: weekly version bumps
- Go modules / Python packages: weekly updates, dev dependencies grouped into a single PR
- Use `gomod` for Go projects, `pip` for Python projects
