package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func (a *App) newBuildCmd() *cobra.Command {
	var out string
	c := &cobra.Command{
		Use:   "build [name...]",
		Short: "Write bundles/fragments to dist/bundles/ (default: all bundles)",
		RunE: func(cmd *cobra.Command, args []string) error {
			names := args
			if len(names) == 0 {
				var err error
				if names, err = a.e.List(); err != nil {
					return err
				}
			}
			return a.writeItems(cmd.ErrOrStderr(), out, names)
		},
	}
	c.Flags().StringVarP(&out, "output", "o", "dist/bundles", "output directory")
	return c
}

// writeItems assembles every named bundle or fragment before writing anything to
// disk, so a failure partway through never leaves a partially populated output
// directory. A bundle is written to <out>/<name>.md; a fragment is written to
// <out>/<path>, mirroring its location under src/ (so e.g. "python/style.md"
// lands at <out>/python/style.md, and same-named fragments in different
// languages never collide).
func (a *App) writeItems(errw io.Writer, out string, names []string) error {
	type result struct {
		path string
		data []byte
	}
	results := make([]result, 0, len(names))
	for _, n := range names {
		switch {
		case a.e.HasBundle(n):
			data, err := a.e.Assemble(n)
			if err != nil {
				return err
			}
			results = append(results, result{path: n + ".md", data: data})
		case a.e.HasFragment(n):
			data, err := a.e.Fragment(n)
			if err != nil {
				return err
			}
			results = append(results, result{path: n, data: data})
		default:
			return fmt.Errorf("%s: not a known bundle or fragment (run moon list to see bundles)", n)
		}
	}

	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	for _, r := range results {
		dst := filepath.Join(out, r.path)
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(dst, r.data, 0o644); err != nil {
			return err
		}
		fmt.Fprintf(errw, "[*] wrote %s\n", dst)
	}
	fmt.Fprintf(errw, "[*] built %d item(s) into %s/\n", len(results), out)
	return nil
}
