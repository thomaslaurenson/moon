package detect

import (
	"testing"
	"testing/fstest"
)

func TestDetect(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		fsys        fstest.MapFS
		wantBundles []string
	}{
		{
			name:        "go project",
			fsys:        fstest.MapFS{"go.mod": {Data: []byte("module x\n")}, "main.go": {}},
			wantBundles: []string{"go-cli"},
		},
		{
			name:        "python project with pyproject prefers app tier over script",
			fsys:        fstest.MapFS{"pyproject.toml": {}, "pkg/__init__.py": {}},
			wantBundles: []string{"python-app"},
		},
		{
			name:        "loose python script with no pyproject",
			fsys:        fstest.MapFS{"tasks/backup.py": {}},
			wantBundles: []string{"python-script"},
		},
		{
			name:        "cpp project",
			fsys:        fstest.MapFS{"CMakeLists.txt": {}, "src/main.cpp": {}},
			wantBundles: []string{"cpp-app"},
		},
		{
			name:        "go project with no main.go anywhere is a library",
			fsys:        fstest.MapFS{"go.mod": {Data: []byte("module x\n")}, "parser.go": {}},
			wantBundles: []string{"go-lib"},
		},
		{
			name:        "cpp project with a root include/ dir is a library",
			fsys:        fstest.MapFS{"CMakeLists.txt": {}, "include/mylib/parser.h": {}, "src/parser.cpp": {}},
			wantBundles: []string{"cpp-lib"},
		},
		{
			name:        "wow addon",
			fsys:        fstest.MapFS{"MyAddon.toc": {}, "MyAddon.lua": {}},
			wantBundles: []string{"wow-addon"},
		},
		{
			name:        "bash project",
			fsys:        fstest.MapFS{"install.sh": {}},
			wantBundles: []string{"bash-script"},
		},
		{
			name:        "empty repo detects nothing",
			fsys:        fstest.MapFS{"README.md": {}},
			wantBundles: nil,
		},
		{
			name:        "multi-language repo detects both",
			fsys:        fstest.MapFS{"go.mod": {}, "main.go": {}, "scripts/build.sh": {}},
			wantBundles: []string{"go-cli", "bash-script"},
		},
		{
			name: "marker files inside skipped dirs are ignored",
			fsys: fstest.MapFS{
				"README.md":                   {},
				"node_modules/pkg/go.mod":     {}, // must NOT trigger go-cli
				"vendor/thing/CMakeLists.txt": {}, // must NOT trigger cpp-app
			},
			wantBundles: nil,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			matches, err := Detect(tc.fsys)
			if err != nil {
				t.Fatalf("Detect: %v", err)
			}
			var got []string
			for _, m := range matches {
				got = append(got, m.Bundle)
				if m.Glob == "" {
					t.Errorf("match %s has empty glob", m.Bundle)
				}
			}
			if len(got) != len(tc.wantBundles) {
				t.Fatalf("Detect() bundles = %v, want %v", got, tc.wantBundles)
			}
			for i := range got {
				if got[i] != tc.wantBundles[i] {
					t.Errorf("Detect() bundles = %v, want %v", got, tc.wantBundles)
					break
				}
			}
		})
	}
}
