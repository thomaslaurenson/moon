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
