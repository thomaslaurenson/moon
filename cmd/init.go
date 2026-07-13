package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/thomaslaurenson/moon/internal/detect"
	"github.com/thomaslaurenson/moon/internal/target"
)

func (a *App) newInitCmd() *cobra.Command {
	var dir string
	var force, dryRun bool
	c := &cobra.Command{
		Use:   fmt.Sprintf("init <%s> [bundle...]", strings.Join(target.Names(), "|")),
		Short: "Populate a repo for a tool (claude, agents, copilot)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.runInit(cmd.OutOrStdout(), cmd.ErrOrStderr(), args[0], args[1:], dir, force, dryRun)
		},
	}
	c.Flags().StringVarP(&dir, "directory", "C", ".", "target directory")
	c.Flags().BoolVar(&force, "force", false, "overwrite existing files")
	c.Flags().BoolVar(&dryRun, "dry-run", false, "list files that would be written, without writing them")
	return c
}

func (a *App) runInit(out, errw io.Writer, targetName string, bundleNames []string, dir string, force, dryRun bool) error {
	root, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	if !insideGitRepo(root) {
		return fmt.Errorf("%s is not inside a git repository (no .git found); moon init requires one", root)
	}

	matches, err := a.resolveInitMatches(root, targetName, bundleNames)
	if err != nil {
		return err
	}

	names := make([]string, len(matches))
	bundles := make([]target.Bundle, len(matches))
	for i, m := range matches {
		data, err := a.e.Assemble(m.Bundle)
		if err != nil {
			return err
		}
		names[i] = m.Bundle
		bundles[i] = target.Bundle{Name: m.Bundle, Content: data, Glob: m.Glob}
	}
	combined, err := a.e.AssembleMany(names)
	if err != nil {
		return err
	}

	files, err := target.Plan(targetName, bundles, combined)
	if err != nil {
		return err
	}

	if !force {
		var existing []string
		for _, f := range files {
			if _, err := os.Stat(filepath.Join(root, f.Path)); err == nil {
				existing = append(existing, f.Path)
			}
		}
		if len(existing) > 0 {
			return fmt.Errorf("would overwrite existing file(s): %s (use --force)", strings.Join(existing, ", "))
		}
	}

	if dryRun {
		for _, f := range files {
			fmt.Fprintln(out, f.Path)
		}
		fmt.Fprintf(errw, "[*] dry run: %d file(s) would be written for target %s\n", len(files), targetName)
		return nil
	}
	for _, f := range files {
		dst := filepath.Join(root, f.Path)
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(dst, f.Content, 0o644); err != nil {
			return err
		}
		fmt.Fprintf(errw, "[*] wrote %s\n", dst)
	}
	fmt.Fprintf(errw, "[*] initialised %s (%d file(s)) for target %s\n", root, len(files), targetName)
	return nil
}

// resolveInitMatches determines which bundles an init run should use: explicit
// names if given (skipping detection entirely), otherwise language detection against
// the target directory's contents. It returns matches (bundle name + applyTo glob);
// the caller assembles their content.
func (a *App) resolveInitMatches(root, targetName string, bundleNames []string) ([]detect.Match, error) {
	if len(bundleNames) > 0 {
		matches := make([]detect.Match, 0, len(bundleNames))
		for _, name := range bundleNames {
			if !a.e.HasBundle(name) {
				return nil, fmt.Errorf("%s: not a known bundle (run moon list to see bundles)", name)
			}
			matches = append(matches, detect.Match{Bundle: name, Glob: target.GlobForBundle(name)})
		}
		return matches, nil
	}
	detected, err := detect.Detect(os.DirFS(root))
	if err != nil {
		return nil, fmt.Errorf("detecting project languages: %w", err)
	}
	if len(detected) == 0 {
		return nil, fmt.Errorf(
			"could not detect a language in %s; pass bundle names explicitly, e.g. moon init %s python-lib",
			root, targetName)
	}
	return detected, nil
}

// insideGitRepo reports whether dir or one of its ancestors contains a .git
// entry. It checks the filesystem directly rather than shelling out to git, so
// moon init works without the git binary being installed or on PATH.
func insideGitRepo(dir string) bool {
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return false
		}
		dir = parent
	}
}
