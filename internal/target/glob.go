package target

import "strings"

var globByPrefix = map[string]string{
	"python":     "**/*.py",
	"go":         "**/*.go",
	"cpp":        "**/*.{cpp,cc,h,hpp}",
	"bash":       "**/*.sh",
	"powershell": "**/*.ps1",
	"wow":        "**/*.lua",
}

var globByName = map[string]string{
	"markdown": "**/*.md",
	"docker":   "**/{Dockerfile,docker-compose.yml,docker-compose.yaml}",
}

// scopedSuffixes mark bundles whose instructions are about a single language's
// source files (code-authoring rules, single scripts). Only these get a narrow
// language glob; a full-project bundle carries repo-wide rules (Makefile, CI,
// dependabot) that should attach to every file, so it falls back to "**".
var scopedSuffixes = []string{"-code", "-script", "-lua"}

// GlobForBundle returns a best-guess applyTo glob for a bundle name, used to scope
// a Copilot instructions file. Single-purpose bundles (markdown, docker) and
// language-scoped bundles (names ending in -code, -script, -lua) get a narrow glob;
// full-project bundles get "**" so their repo-wide rules apply everywhere. Falls
// back to "**" when the language can't be inferred, which is safe (broader than
// ideal) rather than wrong.
func GlobForBundle(name string) string {
	if g, ok := globByName[name]; ok {
		return g
	}
	scoped := false
	for _, s := range scopedSuffixes {
		if strings.HasSuffix(name, s) {
			scoped = true
			break
		}
	}
	if !scoped {
		return "**"
	}
	prefix, _, found := strings.Cut(name, "-")
	if !found {
		prefix = name
	}
	if g, ok := globByPrefix[prefix]; ok {
		return g
	}
	return "**"
}
