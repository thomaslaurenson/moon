package cmd

import (
	"strings"
	"testing"
)

func TestVersionCmd(t *testing.T) {
	t.Parallel()
	stdout, _, err := run(t, testFS(), "version")
	if err != nil {
		t.Fatalf("version: %v", err)
	}
	if strings.TrimSpace(stdout) != Version {
		t.Errorf("stdout = %q, want %q", stdout, Version)
	}
}

func TestVersionFrom(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		injected string
		module   string
		want     string
	}{
		{name: "go install of a tag", injected: "dev", module: "v1.2.3", want: "1.2.3"},
		{name: "prerelease tag install", injected: "dev", module: "v1.2.3-rc.1", want: "1.2.3-rc.1"},
		{name: "no vcs information", injected: "dev", module: "(devel)", want: "dev"},
		{name: "empty module version", injected: "dev", module: "", want: "dev"},
		{name: "pseudo-version with no earlier tag", injected: "dev", module: "v0.0.0-20240101120000-abcdef123456", want: "dev"},
		{name: "pseudo-version above a tag", injected: "dev", module: "v1.2.4-0.20240101120000-abcdef123456", want: "dev"},
		{name: "dirty tagged build", injected: "dev", module: "v1.2.3+dirty", want: "dev"},
		{name: "ldflags always wins", injected: "1.9.9", module: "v1.2.3", want: "1.9.9"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := versionFrom(tc.injected, tc.module); got != tc.want {
				t.Errorf("versionFrom(%q, %q) = %q, want %q", tc.injected, tc.module, got, tc.want)
			}
		})
	}
}
