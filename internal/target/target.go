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

// Plan computes the files a target would write for the given bundles.
func Plan(name string, bundles []Bundle) ([]PlannedFile, error) {
	switch name {
	case "claude":
		return planSingleFile("CLAUDE.md", bundles), nil
	case "agents":
		return planSingleFile("AGENTS.md", bundles), nil
	case "copilot":
		return planCopilot(bundles), nil
	default:
		return nil, fmt.Errorf("unknown init target: %s (known: %s)", name, strings.Join(Names(), ", "))
	}
}

// planSingleFile concatenates every bundle into one always-on instructions file,
// the convention used by CLAUDE.md and AGENTS.md.
func planSingleFile(filename string, bundles []Bundle) []PlannedFile {
	var buf bytes.Buffer
	for i, b := range bundles {
		if i > 0 {
			buf.WriteByte('\n')
		}
		buf.Write(b.Content)
	}
	return []PlannedFile{{Path: filename, Content: buf.Bytes()}}
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
