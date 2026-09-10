# Changelog

## Unreleased

### Added

- Add C++ coverage and sanitizer targets, and the CI jobs that run them
- Add a get_version target, and build C++ examples in CI

### Changed

- Test C++ under both build types as well as both compilers
- Build and test each platform in one job instead of testing a downloaded artifact
- Describe integration test data by its contract rather than one project's sourcing
- Take file paths as std::filesystem::path throughout the C++ examples
- Note the runner cost multipliers and make the cheap platform set the default

### Fixed

- Fix the library version header, warning bar, and fuzz harness wiring
- Fix the member naming rule to apply the trailing underscore by access
- Make every C++ make target runnable from a clean checkout

### Removed

- Remove macOS from the C++ release path

## 0.4.2 - 2026-09-10

### Added

- Add a shared output marker vocabulary for messages a command line tool prints

### Removed

- Remove the binary name prefix from fatal error lines in the Go and bash fragments

## 0.4.1 - 2026-08-25

### Added

- Add a review-findings bundle for running a review and reporting what it found
- Add a review-fixes bundle for working through findings one at a time

## 0.4.0 - 2026-08-24

### Added

- Add a bundle expand subcommand, replacing the bundle show list flag

### Changed

- Make a bare invocation print usage to stderr and exit non-zero
- Render make help with section headings, ordered by everyday use
- Require the debug flag only for CLIs that emit diagnostics
- Quote the failing input in error messages

### Fixed

- Fix version reporting for go install builds and local builds near the dev tag
- Fix the dev release check to treat API failures as failures

## 0.3.0 - 2026-08-11

### Added

- Add comment conventions covering history, earned length, and pre-empting wrong fixes
- Add C++ portability rules for include hygiene, binary I/O, and path types
- Add container image labels and ghcr publishing to the C++ application tier
- Expand the bash workflow, testing, and style fragments, add bash-project badges

### Changed

- Tie C++ CI depth to the trigger, test under both GCC and clang
- Define the C++ library workflow set, make build and prerelease depend on the tier
- Split the C++ build directory by configuration, define the warning bar once

### Fixed

- Configure clang-tidy against clang so its lint output stops being fiction
- Correct the fetch-depth, GH_REPO, gpipe Go, and get_changelog header rules

### Removed

- Remove install_clang_tools, resolving clang tool paths per platform instead

## 0.2.6 - 2026-08-06

### Updated

- Move to new gpipe action

## 0.2.5 - 2026-07-25

### Fixed

- Parallel build for cpp
- Tidied cpp testing approach

## 0.2.4 - 2026-07-22

### Added

- Add a non-ASCII character check to the check command

### Changed

- Make the third-party include example in the cpp fragments generic

### Removed

- Remove the branches fragment

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
