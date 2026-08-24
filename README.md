# moon

![Build Status](https://img.shields.io/github/actions/workflow/status/thomaslaurenson/moon/tag.yml?style=flat&logo=github) ![Test Status](https://img.shields.io/github/actions/workflow/status/thomaslaurenson/moon/tag.yml?style=flat&label=test&logo=github)

![Release Version](https://img.shields.io/github/v/release/thomaslaurenson/moon?style=flat&logo=github) ![Release downloads](https://img.shields.io/github/downloads/thomaslaurenson/moon/total?label=downloads&logo=github)

![Go Version](https://img.shields.io/github/go-mod/go-version/thomaslaurenson/moon?logo=go) ![Code Coverage](https://img.shields.io/badge/Coverage-90.3%25-blue?logo=go)

To the moon! A self-contained binary that composes AI agent instructions from markdown fragments.

## What

- moon assembles markdown bundles from fragments to make dynamic specs for AI agent instructions
- A **fragment** is a single markdown file (`src/fragments`)
- A **bundle** is a named composition of fragments (`src/bundles`)
- All markdown is embedded into the binary at build time, so the compiled `moon` needs no files alongside it at runtime

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
moon bundle list --json        # machine-readable output (also on: fragment list)
moon bundle show <name>        # print an assembled bundle to stdout
moon bundle expand <name>      # list the fragments a bundle expands to
moon fragment list [filter]    # list fragment paths (optionally filtered)
moon fragment show <path>      # print a single fragment to stdout
moon init <target> [bundle...] # populate a repo for claude, agents, or copilot
```
