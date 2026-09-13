// Package detect infers which moon bundles apply to a project by looking for
// marker files (go.mod, pyproject.toml, and so on). Detection only ever picks a
// bundle's default tier; callers that want a different tier (or a bundle detect
// can't infer) should pass bundle names explicitly instead of relying on this
// package. Go, C++, and Python each distinguish their tiers using cheap
// structural signals: a main.go anywhere (Go binary), root-level include/ and
// app/ directories (a C++ public API and a C++ shipped binary respectively), or
// a [build-system] table in pyproject.toml (installable Python package,
// optionally with [project.scripts] for a console script). These are heuristics;
// when one is wrong, an explicit bundle name always wins.
package detect
