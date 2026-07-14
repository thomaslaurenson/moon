# moon

![Build Status](https://img.shields.io/github/actions/workflow/status/thomaslaurenson/moon/tag.yml?style=flat&logo=github) ![Test Status](https://img.shields.io/github/actions/workflow/status/thomaslaurenson/moon/tag.yml?style=flat&label=test&logo=github)

![Release Version](https://img.shields.io/github/v/release/thomaslaurenson/moon?style=flat&logo=github) ![Release downloads](https://img.shields.io/github/downloads/thomaslaurenson/moon/total?label=downloads&logo=github)

![Go Version](https://img.shields.io/github/go-mod/go-version/thomaslaurenson/moon?logo=go) ![Code Coverage](https://img.shields.io/badge/Coverage-91%25-blue?logo=go)

To the moon! A self-contained binary that composes AI agent instructions from markdown fragments.

## What

A **fragment** is a single markdown file (`src/fragments`). A **bundle** is a named composition of fragments (`src/bundles`). moon assembles bundles from fragments. The whole `src/` tree is embedded into the binary at build time, so the compiled `moon` needs no files alongside it at runtime.

## Installation

Download a pre-built binary from the [releases page](https://github.com/thomaslaurenson/moon/releases). For easier install, use the bash installer script:

```sh
curl -fsSL https://github.com/thomaslaurenson/moon/releases/latest/download/install.sh | bash
```

Or the PowerShell installer script if on Windows:

```ps
irm https://github.com/thomaslaurenson/moon/releases/latest/download/install.ps1 | iex
```

Install from source:

```sh
go install github.com/thomaslaurenson/moon@latest
```

## Usage

```sh
moon bundle list --long        # see every bundle with a one-line description
moon bundle show <name>        # print an assembled bundle to stdout
moon bundle show <name> -l     # list the fragments a bundle expands to
moon fragment list [filter]    # list fragment paths (optionally filtered)
moon fragment show <path>      # print a single fragment to stdout
moon init <target> [bundle...] # populate a repo for claude, agents, or copilot
```

Run `moon help` for the full command reference, including `check`. Shell completion for bundle and fragment names is available via `moon completion <shell>` (bash, zsh, fish, powershell).

## Editing

Edit fragments in `src/fragments` and bundle definitions in `src/bundles`, then rebuild (`make build`) to pick up changes. Run `moon check` (or `make ci`, which includes it) to validate every bundle before committing.
