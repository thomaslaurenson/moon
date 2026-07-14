// Package bundler resolves bundle definitions and assembles instruction bundles
// from a fragment tree.
//
// A bundle definition (src/bundles/<name>) is an ordered list of fragment paths
// relative to src/fragments. Blank lines and content after '#' are ignored. A line
// "@include <bundle>" expands another bundle in place, so bundles share a common base.
package bundler

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

const (
	fragmentsDir = "src/fragments"
	bundlesDir   = "src/bundles"
)

// Sentinel errors, so callers can distinguish failure kinds with errors.Is rather
// than matching on message text.
var (
	// ErrUnknownBundle means the named bundle does not exist in src/bundles.
	ErrUnknownBundle = errors.New("unknown bundle")
	// ErrIncludeCycle means a bundle's @include chain refers back to itself.
	ErrIncludeCycle = errors.New("include cycle detected")
	// ErrMissingFragment means a bundle references a fragment absent from src/fragments.
	ErrMissingFragment = errors.New("missing fragment")
)

// Engine assembles bundles from a filesystem containing src/fragments and src/bundles.
type Engine struct {
	fsys fs.FS
}

// New returns an Engine backed by fsys.
func New(fsys fs.FS) *Engine {
	return &Engine{fsys: fsys}
}

// List returns the available bundle names, sorted.
func (e *Engine) List() ([]string, error) {
	entries, err := fs.ReadDir(e.fsys, bundlesDir)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", bundlesDir, err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	return names, nil
}

// ListFragments returns every fragment path under src/fragments, sorted. Paths are
// relative to src/fragments (the same strings Expand prints and a bundle definition
// references), so a caller can feed one straight back to Fragment or Show.
func (e *Engine) ListFragments() ([]string, error) {
	var paths []string
	err := fs.WalkDir(e.fsys, fragmentsDir, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || !strings.HasSuffix(p, ".md") {
			return nil
		}
		paths = append(paths, strings.TrimPrefix(p, fragmentsDir+"/"))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walking %s: %w", fragmentsDir, err)
	}
	sort.Strings(paths)
	return paths, nil
}

// Description returns a bundle's leading comment block (the '#' lines at the top of
// its definition, before the first real entry), joined into one string. Every bundle
// in this repo's convention starts with one; an empty string means the definition has
// none, not that it's invalid.
func (e *Engine) Description(name string) (string, error) {
	data, err := fs.ReadFile(e.fsys, bundlesDir+"/"+name)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", fmt.Errorf("%s: %w", name, ErrUnknownBundle)
		}
		return "", fmt.Errorf("reading bundle %s: %w", name, err)
	}
	var lines []string
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if !strings.HasPrefix(trimmed, "#") {
			break
		}
		lines = append(lines, strings.TrimSpace(strings.TrimPrefix(trimmed, "#")))
	}
	return strings.Join(lines, " "), nil
}

// Expand returns the ordered fragment paths a bundle expands to, resolving @include.
func (e *Engine) Expand(name string) ([]string, error) {
	return e.resolve(name, nil)
}

// HasBundle reports whether a bundle definition with this name exists.
func (e *Engine) HasBundle(name string) bool {
	return e.isFile(bundlesDir + "/" + name)
}

// HasFragment reports whether a fragment exists at this path under src/fragments.
func (e *Engine) HasFragment(path string) bool {
	return e.isFile(fragmentsDir + "/" + path)
}

func (e *Engine) isFile(p string) bool {
	info, err := fs.Stat(e.fsys, p)
	return err == nil && !info.IsDir()
}

// Fragment returns the raw content of a single fragment, prefixed with a minimal
// header identifying where it came from. Unlike Assemble, it performs no bundle
// resolution: the path must be an exact fragment path relative to src/fragments (the
// same strings Expand or ListFragments print).
func (e *Engine) Fragment(path string) ([]byte, error) {
	data, err := fs.ReadFile(e.fsys, fragmentsDir+"/"+path)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, ErrMissingFragment)
	}
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "<!-- Fragment: %s/%s -->\n\n", fragmentsDir, path)
	buf.Write(data)
	return buf.Bytes(), nil
}

func (e *Engine) resolve(bundle string, seen []string) ([]string, error) {
	data, err := fs.ReadFile(e.fsys, bundlesDir+"/"+bundle)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", bundle, ErrUnknownBundle)
	}
	for _, s := range seen {
		if s == bundle {
			return nil, fmt.Errorf("%s: %w", bundle, ErrIncludeCycle)
		}
	}
	seen = append(seen, bundle)

	var frags []string
	for _, line := range strings.Split(string(data), "\n") {
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if line == "@include" || strings.HasPrefix(line, "@include ") || strings.HasPrefix(line, "@include\t") {
			target := strings.TrimSpace(strings.TrimPrefix(line, "@include"))
			if target == "" {
				return nil, fmt.Errorf("%s: @include needs a target bundle", bundle)
			}
			sub, err := e.resolve(target, seen)
			if err != nil {
				return nil, err
			}
			frags = append(frags, sub...)
			continue
		}
		if strings.HasPrefix(line, "@") {
			return nil, fmt.Errorf("%s: unknown directive %q (only @include is recognised)", bundle, line)
		}
		frags = append(frags, line)
	}
	return frags, nil
}

// Assemble returns the full assembled content for a bundle. It validates that every
// fragment exists before emitting, so a broken bundle never produces partial output.
// Duplicate fragments (from a diamond @include) are emitted once, at first occurrence.
func (e *Engine) Assemble(name string) ([]byte, error) {
	frags, err := e.resolve(name, nil)
	if err != nil {
		return nil, err
	}
	header := fmt.Sprintf("<!-- Generated by moon from src/bundles/%s. Do not edit; edit src/fragments and rerun. -->", name)
	return e.emit(header, frags, name)
}

// AssembleMany assembles several bundles into one document, deduplicating fragments
// across the whole set so shared bases (_core.md, github/*, tools/*) appear once
// rather than repeating per bundle. Fragments keep first-occurrence order. It is the
// basis for single-file init targets (CLAUDE.md, AGENTS.md) that combine bundles.
func (e *Engine) AssembleMany(names []string) ([]byte, error) {
	var all []string
	for _, name := range names {
		frags, err := e.resolve(name, nil)
		if err != nil {
			return nil, err
		}
		all = append(all, frags...)
	}
	header := fmt.Sprintf("<!-- Generated by moon from src/bundles/%s. Do not edit; edit src/fragments and rerun. -->", strings.Join(names, ", "))
	return e.emit(header, all, strings.Join(names, ", "))
}

// emit validates and writes the given fragments under a header. It deduplicates
// (first occurrence wins), separates fragments with exactly one blank line, and
// prefixes each with a provenance comment so a reader (or a debugging maintainer)
// can see which src/fragments file a passage came from. Validation happens up front,
// so a missing fragment never yields partial output. The origin string names the
// bundle(s) for error messages.
func (e *Engine) emit(header string, frags []string, origin string) ([]byte, error) {
	seen := make(map[string]bool, len(frags))
	ordered := make([]string, 0, len(frags))
	for _, f := range frags {
		if seen[f] {
			continue
		}
		seen[f] = true
		ordered = append(ordered, f)
	}
	for _, f := range ordered {
		if _, err := fs.Stat(e.fsys, fragmentsDir+"/"+f); err != nil {
			return nil, fmt.Errorf("%s/%s (in bundle %s): %w", fragmentsDir, f, origin, ErrMissingFragment)
		}
	}
	var buf bytes.Buffer
	buf.WriteString(header)
	buf.WriteString("\n\n")
	for _, f := range ordered {
		data, err := fs.ReadFile(e.fsys, fragmentsDir+"/"+f)
		if err != nil {
			return nil, fmt.Errorf("reading fragment %s: %w", f, err)
		}
		fmt.Fprintf(&buf, "<!-- %s/%s -->\n\n", fragmentsDir, f)
		buf.Write(bytes.TrimRight(data, "\n"))
		buf.WriteString("\n\n")
	}
	// End with exactly one trailing newline.
	out := append(bytes.TrimRight(buf.Bytes(), "\n"), '\n')
	return out, nil
}

// Check validates every bundle. It returns problems (missing fragments or include
// cycles) and orphans (fragments present in src/fragments but referenced by no
// bundle). The returned strings are for human display; callers that need to branch on
// failure kind should call Expand or Assemble directly and use errors.Is instead.
func (e *Engine) Check() (problems, orphans []string, err error) {
	names, err := e.List()
	if err != nil {
		return nil, nil, err
	}
	referenced := make(map[string]bool)
	for _, name := range names {
		frags, rerr := e.resolve(name, nil)
		if rerr != nil {
			problems = append(problems, fmt.Sprintf("%s: %v", name, rerr))
			continue
		}
		for _, f := range frags {
			referenced[f] = true
			if _, serr := fs.Stat(e.fsys, fragmentsDir+"/"+f); serr != nil {
				problems = append(problems, fmt.Sprintf("%s: missing fragment %s", name, f))
			}
		}
	}
	err = fs.WalkDir(e.fsys, fragmentsDir, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || !strings.HasSuffix(p, ".md") {
			return nil
		}
		if rel := strings.TrimPrefix(p, fragmentsDir+"/"); !referenced[rel] {
			orphans = append(orphans, rel)
		}
		return nil
	})
	if err != nil {
		return problems, orphans, err
	}
	sort.Strings(problems)
	sort.Strings(orphans)
	return problems, orphans, nil
}
