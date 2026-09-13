package cmd

import (
	"errors"
	"strings"
	"testing"
	"testing/fstest"
)

func TestCheck(t *testing.T) {
	t.Parallel()

	t.Run("healthy tree passes with a summary on stderr", func(t *testing.T) {
		t.Parallel()
		stdout, stderr, err := run(t, testFS(), "check")
		if err != nil {
			t.Fatalf("check: %v", err)
		}
		if stdout != "" {
			t.Errorf("stdout = %q, want empty (check reports on stderr)", stdout)
		}
		if !strings.Contains(stderr, "checked 2 bundle(s)") {
			t.Errorf("stderr = %q, want a summary line", stderr)
		}
	})

	t.Run("missing fragment fails silently with exit code 1", func(t *testing.T) {
		t.Parallel()
		fsys := testFS()
		fsys["src/bundles/broken"] = &fstest.MapFile{Data: []byte("python/ghost.md\n")}

		stdout, stderr, err := run(t, fsys, "check")
		var ec *ExitCodeError
		if !errors.As(err, &ec) || ec.Code != 1 {
			t.Errorf("check error = %v, want ExitCodeError{1}", err)
		}
		if stdout != "" {
			t.Errorf("stdout = %q, want empty", stdout)
		}
		if !strings.Contains(stderr, "[!]") || !strings.Contains(stderr, "missing fragment") {
			t.Errorf("stderr = %q, want a marked problem line", stderr)
		}
	})

	t.Run("orphan fragment warns but passes", func(t *testing.T) {
		t.Parallel()
		fsys := testFS()
		fsys["src/fragments/orphan.md"] = &fstest.MapFile{Data: []byte("# Orphan\n")}

		_, stderr, err := run(t, fsys, "check")
		if err != nil {
			t.Fatalf("check: %v", err)
		}
		if !strings.Contains(stderr, "[!] orphan fragment") {
			t.Errorf("stderr = %q, want an orphan warning", stderr)
		}
	})
}
