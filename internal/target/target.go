// Package target computes the files an "moon init" target (claude, copilot, ...)
// writes for a set of resolved bundles. Plan is pure: it never touches disk, which
// keeps it fully unit-testable. The caller (cmd) is responsible for actually
// writing the returned files.
package target

import (
	"bytes"
	"fmt"
	"strings"
)

// Bundle is a resolved bundle ready to be written for a target: its name, its
// assembled markdown content, and a best-guess applyTo glob for tools (like
// GitHub Copilot) that attach instructions to matching files automatically.
type Bundle struct {
	Name    string
	Content []byte
	Glob    string
}

// PlannedFile is a file a target would write, relative to the repo root.
type PlannedFile struct {
	Path    string
	Content []byte
}

// Names returns the known init target names, sorted.
func Names() []string {
	return []string{"agents", "claude", "copilot"}
}

// Plan computes the files a target would write. combined is the dedup-merged
// content of all selected bundles (used by the single-file targets); bundles carries
// per-bundle content and globs (used by copilot, whose files are scoped per bundle).
func Plan(name string, bundles []Bundle, combined []byte) ([]PlannedFile, error) {
	switch name {
	case "claude":
		return []PlannedFile{{Path: "CLAUDE.md", Content: combined}}, nil
	case "agents":
		return []PlannedFile{{Path: "AGENTS.md", Content: combined}}, nil
	case "copilot":
		return planCopilot(bundles), nil
	default:
		return nil, fmt.Errorf("unknown init target: %s (known: %s)", name, strings.Join(Names(), ", "))
	}
}

// planCopilot writes one path-specific instructions file per bundle under
// .github/instructions/, each carrying an applyTo glob so Copilot attaches it
// automatically to matching files without the user referencing it by hand.
func planCopilot(bundles []Bundle) []PlannedFile {
	files := make([]PlannedFile, 0, len(bundles))
	for _, b := range bundles {
		glob := b.Glob
		if glob == "" {
			glob = "**"
		}
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "---\napplyTo: \"%s\"\n---\n\n", glob)
		buf.Write(b.Content)
		files = append(files, PlannedFile{
			Path:    fmt.Sprintf(".github/instructions/%s.instructions.md", b.Name),
			Content: buf.Bytes(),
		})
	}
	return files
}
