# Dependabot

## Contents

- [Config](#config) — canonical dependabot.yml for all projects
- [Rules](#rules) — scheduling and limits

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
    open-pull-requests-limit: 0
    assignees:
      - "<username>"
```

Replace `<username>` with the repo owner. Replace `<gomod|pip>` with the appropriate
package ecosystem for the project.

## Rules

- GitHub Actions: weekly version bumps.
- Go modules / Python packages: security updates only (`open-pull-requests-limit: 0`).
