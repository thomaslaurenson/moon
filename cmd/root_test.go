package cmd

import (
	"bytes"
	"errors"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"
)

// testFS returns a minimal fragment and bundle tree for functional tests, so
// they exercise the command layer without depending on the shipped content.
func testFS() fstest.MapFS {
	return fstest.MapFS{
		"src/fragments/_core.md":        {Data: []byte("# Core\n")},
		"src/fragments/python/style.md": {Data: []byte("# Style\n")},
		"src/bundles/py-code":           {Data: []byte("# Python code rules\n_core.md\npython/style.md\n")},
		"src/bundles/py-app":            {Data: []byte("@include py-code\n")},
	}
}

// run builds a fresh command tree over fsys and executes it with args. A fresh
// tree per call keeps flag state from leaking between cases.
func run(t *testing.T, fsys fs.FS, args ...string) (stdout, stderr string, err error) {
	t.Helper()

	var out, errOut bytes.Buffer
	root := NewRootCmd(fsys, &out, &errOut)
	root.SetArgs(args)
	err = root.Execute()

	return out.String(), errOut.String(), err
}

func TestBareRoot(t *testing.T) {
	t.Parallel()
	stdout, stderr, err := run(t, testFS())

	var ec *ExitCodeError
	if !errors.As(err, &ec) || ec.Code != 1 {
		t.Errorf("bare invocation error = %v, want ExitCodeError{1}", err)
	}
	if !strings.Contains(stderr, "Usage:") {
		t.Errorf("bare invocation stderr = %q, want usage text", stderr)
	}
	if stdout != "" {
		t.Errorf("bare invocation stdout = %q, want empty", stdout)
	}
}

func TestUnknownCommand(t *testing.T) {
	t.Parallel()
	stdout, _, err := run(t, testFS(), "bogus")

	if err == nil {
		t.Error("unknown command returned nil error")
	}
	if stdout != "" {
		t.Errorf("unknown command stdout = %q, want empty", stdout)
	}
}
