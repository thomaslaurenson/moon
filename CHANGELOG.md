# Changelog

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
