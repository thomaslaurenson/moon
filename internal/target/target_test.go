package target

import (
	"strings"
	"testing"
)

func TestGlobForBundle(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		want string
	}{
		{name: "python-script", want: "**/*.py"},
		{name: "python-lib-code", want: "**/*.py"},
		{name: "python-lib", want: "**"},
		{name: "go-cli", want: "**"},
		{name: "go-cli-code", want: "**/*.go"},
		{name: "cpp-app", want: "**"},
		{name: "cpp-app-code", want: "**/*.{cpp,cc,h,hpp}"},
		{name: "bash-script", want: "**/*.sh"},
		{name: "bash-project", want: "**"},
		{name: "powershell-script", want: "**/*.ps1"},
		{name: "markdown", want: "**/*.md"},
		{name: "docker", want: "**/{Dockerfile,docker-compose.yml,docker-compose.yaml}"},
		{name: "totally-unknown-bundle", want: "**"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := GlobForBundle(tc.name); got != tc.want {
				t.Errorf("GlobForBundle(%q) = %q, want %q", tc.name, got, tc.want)
			}
		})
	}
}

func TestPlanClaude(t *testing.T) {
	t.Parallel()
	combined := []byte("# Python\n\n# Bash\n")
	files, err := Plan("claude", []Bundle{
		{Name: "python-script", Content: []byte("# Python\n")},
		{Name: "bash-script", Content: []byte("# Bash\n")},
	}, combined)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if len(files) != 1 || files[0].Path != "CLAUDE.md" {
		t.Fatalf("expected a single CLAUDE.md, got %v", files)
	}
	content := string(files[0].Content)
	if !strings.Contains(content, "# Python") || !strings.Contains(content, "# Bash") {
		t.Errorf("CLAUDE.md missing expected content:\n%s", content)
	}
}

func TestPlanAgents(t *testing.T) {
	t.Parallel()
	files, err := Plan("agents", []Bundle{{Name: "go-cli", Content: []byte("# Go\n")}}, []byte("# Go\n"))
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if len(files) != 1 || files[0].Path != "AGENTS.md" {
		t.Fatalf("expected a single AGENTS.md, got %v", files)
	}
}

func TestPlanCopilot(t *testing.T) {
	t.Parallel()
	files, err := Plan("copilot", []Bundle{
		{Name: "python-script", Content: []byte("# Python\n"), Glob: "**/*.py"},
		{Name: "go-cli", Content: []byte("# Go\n"), Glob: "**/*.go"},
	}, nil)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected one file per bundle, got %d", len(files))
	}
	wantPaths := map[string]bool{
		".github/instructions/python-script.instructions.md": true,
		".github/instructions/go-cli.instructions.md":        true,
	}
	for _, f := range files {
		if !wantPaths[f.Path] {
			t.Errorf("unexpected path: %s", f.Path)
		}
		if !strings.HasPrefix(string(f.Content), "---\napplyTo:") {
			t.Errorf("%s: missing applyTo frontmatter, got:\n%s", f.Path, f.Content)
		}
	}
}

func TestPlanUnknownTarget(t *testing.T) {
	t.Parallel()
	if _, err := Plan("nope", nil, nil); err == nil {
		t.Fatal("expected an error for an unknown target")
	}
}
