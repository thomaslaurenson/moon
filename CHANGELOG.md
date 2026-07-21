# Changelog

## 0.2.3 - 2026-07-21

### Changed

- Made CPP testing generic

## 0.2.2 - 2026-07-20

### Added

- Add a get_changelog target and document the v-prefixed tag versus bare changelog header
- Make PyPI publishing optional over a GitHub-release baseline for Python libraries
- Default the Python version badge to a static requires-python badge
- Clarify dependency groups versus extras across the Python fragments

### Changed

- Rename the python-app bundle to python-tools and restructure it around domain-named script directories
- Use standard-library logging across all Python tiers
- Make PyPI publishing optional over a GitHub-release baseline for Python libraries
- Default the Python version badge to a static requires-python badge
- Clarify dependency groups versus extras across the Python fragments
- Print the full coverage table in test_coverage and document the total line as the badge source

### Removed

- Remove the Python structlog logging fragment

## 0.2.1 - 2026-07-17

### Added

- Add the cpp-lib-cli and cpp-lib-cli-code bundles for a library that ships a CLI binary
- Add cpp fragments for error handling, Doxygen, integration testing, and fuzz testing

### Changed

- Restructure the cpp tiers around a public API in include and a shipped binary in app
- Move main into app so src always builds a library target that tests link
- Restructure cpp testing around unit, integration, functional, and fuzz layers
- Give the cpp code bundles per-tier rules, matching the Python tiers
- Add warning, sanitizer, and multi-module library conventions to the cpp cmake fragments
- Split the git conventions fragment into separate commits and branches fragments

### Fixed

- Fix path, variable scope, testing option, and target naming issues across the cpp cmake fragments
- Fix ctest layer selection and skip handling in the cpp testing fragments
- Harden the Go dev prerelease delete against a flaky GitHub API

## 0.2.0 - 2026-07-14

### Added

- Add fragment list and fragment show commands for discovering and printing individual fragments
- Add shell completion for bundle and fragment names via the completion command

### Changed

- Restructure the CLI around two nouns, fragment and bundle, each with list and show subcommands
- Rename list to bundle list, and show to bundle show
- Fold the recipe command into bundle show --list
- Move fragments to src/fragments and bundle definitions to src/bundles

### Removed

- Remove the build command

## 0.1.2 - 2026-07-13

### Added

- Detect Python library and CLI tiers from pyproject build-system and scripts
- Add AssembleMany with fragment deduplication for multi-bundle init
- Add recipe validation to CI

### Changed

- Move build output to dist/bundles, reject unknown directives
- Adopt PEP 735 dependency groups, expand structlog and CI workflow fragments

### Fixed

- Fix cmake, SavedVariables, and fmt_check recipe issues across language fragments

## 0.1.1 - 2026-07-06

### Changed

- Refactor Python modules to better support lib and app projects

### Updated

- Improve Dependabot configuration for security-only updates

## 0.1.0 - 2026-07-06

### Added

- Initial release
