// Package detect infers which moon bundles apply to a project by looking for
// marker files (go.mod, pyproject.toml, and so on). Detection only ever picks a
// bundle's default tier; callers that want a different tier (or a bundle detect
// can't infer) should pass bundle names explicitly instead of relying on this
// package. Go and C++ additionally distinguish an application from a library
// tier using a single cheap structural signal each (a main.go anywhere, or a
// root-level include/ directory); this is a heuristic; when it's wrong, an
// explicit bundle name always wins.
package detect

import (
	"io/fs"
	"strings"

	"github.com/thomaslaurenson/moon/internal/target"
)

// Match is one detected language: the default bundle for it, and the applyTo
// glob tools like Copilot should attach it to.
type Match struct {
	Bundle string
	Glob   string
}

// skipDirs are never descended into: they're either not source, or (for dist)
// moon's own generated-output convention.
var skipDirs = map[string]bool{
	".git": true, "node_modules": true, "vendor": true, "dist": true, ".venv": true,
}

type presence struct {
	goMod, cmakeLists, toc, ps1, pyproject, py, sh bool
	mainGo                                         bool // a main.go anywhere suggests a Go binary, not a library
	includeDir                                     bool // an include/ dir at the root is this repo's C++ library marker
}

// Detect walks fsys (rooted at a project directory) and returns the bundles
// whose language markers were found, one match per detected language, in a
// stable order. An empty result means detection found nothing to go on.
func Detect(fsys fs.FS) ([]Match, error) {
	var p presence
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != "." && skipDirs[d.Name()] {
				return fs.SkipDir
			}
			if path == "include" {
				p.includeDir = true
			}
			return nil
		}
		name := d.Name()
		switch {
		case name == "go.mod":
			p.goMod = true
		case name == "main.go":
			p.mainGo = true
		case name == "CMakeLists.txt":
			p.cmakeLists = true
		case strings.HasSuffix(name, ".toc"):
			p.toc = true
		case strings.HasSuffix(name, ".ps1"):
			p.ps1 = true
		case name == "pyproject.toml":
			p.pyproject = true
		case strings.HasSuffix(name, ".py"):
			p.py = true
		case strings.HasSuffix(name, ".sh"):
			p.sh = true
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var matches []Match
	add := func(bundle string) {
		matches = append(matches, Match{Bundle: bundle, Glob: target.GlobForBundle(bundle)})
	}

	if p.goMod {
		// A repo-wide absence of main.go is the strongest cheap signal that this
		// is a library, not a binary: nothing to run means nothing to be a CLI.
		if p.mainGo {
			add("go-cli")
		} else {
			add("go-lib")
		}
	}
	if p.cmakeLists {
		// include/ is this repo's own convention for a library's public API
		// surface (see cpp/cmake-lib.md); its presence is a reliable signal.
		if p.includeDir {
			add("cpp-lib")
		} else {
			add("cpp-app")
		}
	}
	if p.toc {
		add("wow-addon")
	}
	if p.ps1 {
		add("powershell-script")
	}
	switch {
	case p.pyproject:
		add("python-app")
	case p.py:
		add("python-script")
	}
	if p.sh {
		add("bash-script")
	}
	return matches, nil
}
