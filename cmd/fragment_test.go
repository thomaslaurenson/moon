package cmd

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestFragmentList(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		check   func(t *testing.T, stdout string)
	}{
		{
			name: "default lists every fragment",
			args: []string{"fragment", "list"},
			check: func(t *testing.T, stdout string) {
				if stdout != "_core\npython/style\n" {
					t.Errorf("stdout = %q, want all fragment names sorted", stdout)
				}
			},
		},
		{
			name: "filter keeps matching paths",
			args: []string{"fragment", "list", "python"},
			check: func(t *testing.T, stdout string) {
				if stdout != "python/style\n" {
					t.Errorf("stdout = %q, want only the python fragment", stdout)
				}
			},
		},
		{
			name: "json emits a parseable array",
			args: []string{"fragment", "list", "--json"},
			check: func(t *testing.T, stdout string) {
				var paths []string
				if err := json.Unmarshal([]byte(stdout), &paths); err != nil {
					t.Fatalf("stdout is not JSON: %v", err)
				}
				if len(paths) != 2 {
					t.Errorf("paths = %v, want both fragments", paths)
				}
			},
		},
		{
			name:    "a second argument is an error with empty stdout",
			args:    []string{"fragment", "list", "a", "b"},
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			stdout, _, err := run(t, testFS(), tc.args...)
			if tc.wantErr {
				if err == nil {
					t.Error("want error, got nil")
				}
				if stdout != "" {
					t.Errorf("stdout = %q, want empty on failure", stdout)
				}
				return
			}
			if err != nil {
				t.Fatalf("run(%v): %v", tc.args, err)
			}
			tc.check(t, stdout)
		})
	}
}

func TestFragmentShow(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		want    string
	}{
		{
			name: "prints the fragment with provenance",
			args: []string{"fragment", "show", "python/style"},
			want: "# Style",
		},
		{
			name:    "unknown fragment is an error with empty stdout",
			args:    []string{"fragment", "show", "ghost"},
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			stdout, _, err := run(t, testFS(), tc.args...)
			if tc.wantErr {
				if err == nil {
					t.Error("want error, got nil")
				}
				if stdout != "" {
					t.Errorf("stdout = %q, want empty on failure", stdout)
				}
				return
			}
			if err != nil {
				t.Fatalf("run(%v): %v", tc.args, err)
			}
			if !strings.Contains(stdout, tc.want) || !strings.Contains(stdout, "Fragment: src/fragments/python/style.md") {
				t.Errorf("stdout = %q, want content and provenance header", stdout)
			}
		})
	}
}
