# Dependabot

File `.github/dependabot.yml`. Replace `<username>` with the repo owner and `<ecosystem>` with the language ecosystem: `gomod` for Go, `uv` for Python. C++ has no language entry; see below. The `<ecosystem>` value stays a placeholder here because this fragment is shared across languages; substitute the concrete value for the project's language.

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
    assignees:
      - "<username>"
```

- GitHub Actions: weekly bumps. Language packages: weekly, one pull request per dependency.
- Do not add a `groups:` entry keyed on `dependency-type: "development"`. GitHub supports that grouping for `bundler`, `composer`, `mix`, `maven`, `npm` and `pip` only. None of the ecosystems these specs use is on that list, so the group matches nothing, produces no error, and reads for ever after as though dev dependencies were being handled.
- C++ carries the `github-actions` entry alone. Every C++ dependency is a git submodule under `extern/` pinned to an immutable reference (see cpp/cmake), and the `gitsubmodule` ecosystem follows the commits on a submodule's branch, not its releases, so it proposes a bump every time upstream merges anything and never says which bump is a release. A submodule moves by hand, to the next upstream release worth taking, or to a reviewed commit where upstream does not tag releases; see cpp/tooling.
- For Python the ecosystem is always `uv`, never `pip`. uv is the default for every Python project in these specs, and the `uv` ecosystem reads both `pyproject.toml` and `uv.lock`, so bumps regenerate the lockfile and CI stays green. The `pip` ecosystem updates `pyproject.toml` but leaves `uv.lock` stale, so reserve it for a genuinely legacy, non-uv Python project only.

A project that ships a Dockerfile adds an entry for the `docker` ecosystem. The Docker fragment pins the base image to a minor tag, and nothing else moves that pin; Dependabot then proposes the newer tag the same way it bumps an action:

```yaml
  - package-ecosystem: docker
    directory: /
    schedule:
      interval: weekly
    assignees:
      - "<username>"
```

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

This only suppresses routine version-update PRs from this file. GitHub's separate Dependabot security-updates feature (a repo-level setting, independent of this file) still opens PRs for vulnerable dependencies regardless of this limit. One ecosystem entry covers every dependency in this mode.
