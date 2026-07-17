package bundler

import (
	"os"
	"testing"
)

// repoFS returns an fs.FS rooted at the repository root (two levels up from this
// package), so the test exercises the real src/fragments and src/bundles trees
// rather than a synthetic fixture. This is the check that guards against a bundle
// referencing a fragment that does not exist, or a fragment being left in no bundle.
func repoFS(t *testing.T) *Engine {
	t.Helper()
	return New(os.DirFS("../.."))
}

// TestRealContentHasNoProblems fails if any shipped bundle references a missing
// fragment or forms an include cycle. It is the regression test for the class of
// bug where a fragment is renamed or removed but a bundle still points at it.
func TestRealContentHasNoProblems(t *testing.T) {
	t.Parallel()
	problems, _, err := repoFS(t).Check()
	if err != nil {
		t.Fatalf("Check on real content: %v", err)
	}
	if len(problems) > 0 {
		t.Errorf("real content has %d problem(s):", len(problems))
		for _, p := range problems {
			t.Errorf("  %s", p)
		}
	}
}

// standaloneFragments are fragments deliberately left out of every bundle, read
// on demand via "moon fragment show" instead. They are the exception: a fragment
// belongs in a bundle unless the task it covers is one a human does by hand, so
// that shipping it to an agent would only spend context.
var standaloneFragments = map[string]bool{
	// Branches are created manually, never by an agent working in a bundled repo.
	"github/branches.md": true,
}

// TestRealContentHasNoOrphans fails if a fragment under src/fragments is referenced
// by no bundle and is not declared standalone above. Orphans are not broken output,
// but in this repo a fragment is meant to belong to at least one bundle, so an
// undeclared orphan signals either a dropped bundle line or a fragment that should
// be deleted.
func TestRealContentHasNoOrphans(t *testing.T) {
	t.Parallel()
	_, orphans, err := repoFS(t).Check()
	if err != nil {
		t.Fatalf("Check on real content: %v", err)
	}
	var unexpected []string
	for _, o := range orphans {
		if !standaloneFragments[o] {
			unexpected = append(unexpected, o)
		}
	}
	if len(unexpected) > 0 {
		t.Errorf("real content has %d orphan fragment(s):", len(unexpected))
		for _, o := range unexpected {
			t.Errorf("  %s", o)
		}
	}
}

// TestStandaloneFragmentsExist fails if a fragment declared standalone above has
// been renamed or deleted, which would otherwise leave a stale entry silently
// excusing a future orphan of the same name.
func TestStandaloneFragmentsExist(t *testing.T) {
	t.Parallel()
	frags, err := repoFS(t).ListFragments()
	if err != nil {
		t.Fatalf("ListFragments: %v", err)
	}
	present := make(map[string]bool, len(frags))
	for _, f := range frags {
		present[f] = true
	}
	for name := range standaloneFragments {
		if !present[name] {
			t.Errorf("standaloneFragments names %q, which does not exist", name)
		}
	}
}

// TestEveryBundleAssembles confirms that every bundle assembles without error,
// which additionally exercises the up-front fragment-existence validation in
// Assemble that Check does not (Assemble reads each fragment; Check only stats).
func TestEveryBundleAssembles(t *testing.T) {
	t.Parallel()
	e := repoFS(t)
	names, err := e.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(names) == 0 {
		t.Fatal("List returned no bundles; repo tree not found")
	}
	for _, name := range names {
		if _, err := e.Assemble(name); err != nil {
			t.Errorf("Assemble(%q): %v", name, err)
		}
	}
}
