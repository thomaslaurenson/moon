# moon

![Build Status](https://img.shields.io/github/actions/workflow/status/thomaslaurenson/moon/tag.yml?style=flat&logo=github) ![Test Status](https://img.shields.io/github/actions/workflow/status/thomaslaurenson/moon/tag.yml?style=flat&label=test&logo=github)

![Release Version](https://img.shields.io/github/v/release/thomaslaurenson/moon?style=flat&logo=github) ![Release downloads](https://img.shields.io/github/downloads/thomaslaurenson/moon/total?label=downloads&logo=github)

![Go Version](https://img.shields.io/github/go-mod/go-version/thomaslaurenson/moon?logo=go) ![Code Coverage](https://img.shields.io/badge/Coverage-91%25-blue?logo=go)

To the moon! A self-contained binary that composes AI agent instructions from markdown fragments.

## What

moon assembles **bundles** (named recipes in `bundles/`) from markdown **fragments** (files in `src/`). The whole `src/` and `bundles/` tree is embedded into the binary at build time, so the compiled `moon` needs no files alongside it at runtime.

## Install

```sh
make build    # produces ./moon for the current platform
make ci       # fmt check, vet, race tests
```

Cross-platform release binaries are built with GoReleaser.

## Usage

```sh
moon list --long             # see every bundle with a one-line description
moon show <bundle>           # print an assembled bundle to stdout
moon init <target> <bundle>  # populate a repo for claude, agents, or copilot
```

Run `moon help` for the full command reference, including `build`, `recipe`, and `check`.

## Editing

Edit fragments in `src/` and recipes in `bundles/`, then rebuild (`make build`) to pick up changes. Run `moon check` to validate every recipe.
