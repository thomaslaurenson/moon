package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// gitDir returns a temporary directory containing a .git entry, satisfying the
// repository check without shelling out to git.
func gitDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestInit(t *testing.T) {
	t.Parallel()

	t.Run("dry run lists files without writing", func(t *testing.T) {
		t.Parallel()
		dir := gitDir(t)
		stdout, _, err := run(t, testFS(), "init", "claude", "py-code", "--directory", dir, "--dry-run")
		if err != nil {
			t.Fatalf("init --dry-run: %v", err)
		}
		if strings.TrimSpace(stdout) != "CLAUDE.md" {
			t.Errorf("stdout = %q, want the planned file list", stdout)
		}
		if _, err := os.Stat(filepath.Join(dir, "CLAUDE.md")); err == nil {
			t.Error("dry run wrote CLAUDE.md")
		}
	})

	t.Run("writes the assembled bundle", func(t *testing.T) {
		t.Parallel()
		dir := gitDir(t)
		_, stderr, err := run(t, testFS(), "init", "claude", "py-code", "--directory", dir)
		if err != nil {
			t.Fatalf("init: %v", err)
		}
		data, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
		if err != nil {
			t.Fatalf("reading CLAUDE.md: %v", err)
		}
		if !strings.Contains(string(data), "# Core") {
			t.Errorf("CLAUDE.md = %q, want assembled content", data)
		}
		if !strings.Contains(stderr, "initialised") {
			t.Errorf("stderr = %q, want a completion note", stderr)
		}
	})

	t.Run("existing file needs force", func(t *testing.T) {
		t.Parallel()
		dir := gitDir(t)
		if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("old"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, _, err := run(t, testFS(), "init", "claude", "py-code", "--directory", dir); err == nil {
			t.Error("overwrite without --force did not error")
		}
		if _, _, err := run(t, testFS(), "init", "claude", "py-code", "--directory", dir, "--force"); err != nil {
			t.Errorf("init --force: %v", err)
		}
	})

	t.Run("outside a git repository is an error", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		if _, _, err := run(t, testFS(), "init", "claude", "py-code", "--directory", dir); err == nil {
			t.Error("init outside a git repository did not error")
		}
	})

	t.Run("unknown target and bundle are errors", func(t *testing.T) {
		t.Parallel()
		dir := gitDir(t)
		if _, _, err := run(t, testFS(), "init", "bogus", "py-code", "--directory", dir); err == nil {
			t.Error("unknown target did not error")
		}
		if _, _, err := run(t, testFS(), "init", "claude", "nope", "--directory", dir); err == nil {
			t.Error("unknown bundle did not error")
		}
	})
}
