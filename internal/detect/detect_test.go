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
			name:        "pyproject without build-system is a tools project (not installable)",
			fsys:        fstest.MapFS{"pyproject.toml": {Data: []byte("[project]\nname = \"x\"\n")}, "client/run.py": {}},
			wantBundles: []string{"python-tools"},
		},
		{
			name:        "pyproject with build-system is a library",
			fsys:        fstest.MapFS{"pyproject.toml": {Data: []byte("[project]\nname = \"x\"\n\n[build-system]\nrequires = [\"hatchling\"]\n")}, "pkg/__init__.py": {}},
			wantBundles: []string{"python-lib"},
		},
		{
			name:        "pyproject with build-system and scripts is a lib-cli",
			fsys:        fstest.MapFS{"pyproject.toml": {Data: []byte("[project]\nname = \"x\"\n\n[project.scripts]\nmycli = \"x:main\"\n\n[build-system]\nrequires = [\"hatchling\"]\n")}, "pkg/__init__.py": {}},
			wantBundles: []string{"python-lib-cli"},
		},
		{
			name:        "loose python script with no pyproject",
			fsys:        fstest.MapFS{"tasks/backup.py": {}},
			wantBundles: []string{"python-script"},
		},
		{
			name:        "cpp project with main() still in src/ is an application",
			fsys:        fstest.MapFS{"CMakeLists.txt": {}, "src/main.cpp": {}},
			wantBundles: []string{"cpp-app"},
		},
		{
			name:        "cpp project with an app/ dir and no include/ is an application",
			fsys:        fstest.MapFS{"CMakeLists.txt": {}, "app/main.cpp": {}, "src/parser.cpp": {}},
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
			// examples/ is a demo built on request, not a shipped binary, so a
			// library that has one is still a library (see cpp/cmake-lib.md).
			name: "cpp library with examples/ but no app/ is still a library",
			fsys: fstest.MapFS{
				"CMakeLists.txt": {}, "include/gale/auth.h": {},
				"src/auth/auth.cpp": {}, "examples/gale_auth/main.cpp": {},
			},
			wantBundles: []string{"cpp-lib"},
		},
		{
			name: "cpp project with both include/ and app/ is a lib-cli",
			fsys: fstest.MapFS{
				"CMakeLists.txt": {}, "include/mylib/parser.h": {},
				"src/parser.cpp": {}, "app/main.cpp": {},
			},
			wantBundles: []string{"cpp-lib-cli"},
		},
		{
			// Only a root-level app/ marks a shipped binary; the walk must not be
			// fooled by the name appearing deeper in the tree.
			name: "cpp library with a nested app dir is not a lib-cli",
			fsys: fstest.MapFS{
				"CMakeLists.txt": {}, "include/mylib/parser.h": {},
				"src/app/dispatch.cpp": {},
			},
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
