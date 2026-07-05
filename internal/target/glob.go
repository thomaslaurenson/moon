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

// GlobForBundle returns a best-guess applyTo glob for a bundle name, used when a
// bundle is chosen explicitly rather than through language detection. It falls
// back to "**" (all files) if the bundle's language can't be inferred from its
// name, which is safe (broader than ideal) rather than wrong.
func GlobForBundle(name string) string {
	if g, ok := globByName[name]; ok {
		return g
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
