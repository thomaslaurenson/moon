package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

// newBundleCmd groups the verbs that operate on bundles: named compositions of
// fragments, defined under src/bundles.
func (a *App) newBundleCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "bundle",
		Short: "List or assemble bundles (named compositions of fragments)",
	}
	c.AddCommand(a.newBundleListCmd(), a.newBundleShowCmd())
	return c
}

func (a *App) newBundleListCmd() *cobra.Command {
	var long, asJSON bool
	c := &cobra.Command{
		Use:   "list",
		Short: "List available bundles",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return a.bundleList(cmd.OutOrStdout(), long, asJSON)
		},
	}
	c.Flags().BoolVarP(&long, "long", "l", false, "show each bundle's one-line description")
	c.Flags().BoolVar(&asJSON, "json", false, "output as structured JSON")
	return c
}

// bundleList writes the available bundle names to out; long and asJSON add
// per-bundle descriptions as aligned columns or structured JSON respectively.
func (a *App) bundleList(out io.Writer, long, asJSON bool) error {
	names, err := a.e.List()
	if err != nil {
		return err
	}

	if !long && !asJSON {
		for _, n := range names {
			fmt.Fprintln(out, n)
		}
		return nil
	}

	type entry struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	entries := make([]entry, 0, len(names))
	for _, n := range names {
		desc, err := a.e.Description(n)
		if err != nil {
			return err
		}
		entries = append(entries, entry{Name: n, Description: desc})
	}

	if asJSON {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(entries)
	}
	width := 0
	for _, e := range entries {
		if len(e.Name) > width {
			width = len(e.Name)
		}
	}
	for _, e := range entries {
		fmt.Fprintf(out, "%-*s  %s\n", width, e.Name, e.Description)
	}
	return nil
}

func (a *App) newBundleShowCmd() *cobra.Command {
	var listOnly bool
	c := &cobra.Command{
		Use:               "show <name>",
		Short:             "Assemble and print a bundle, or with --list show its fragments",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: a.completeBundles,
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.bundleShow(cmd.OutOrStdout(), cmd.ErrOrStderr(), args[0], listOnly)
		},
	}
	c.Flags().BoolVarP(&listOnly, "list", "l", false, "list the fragments the bundle expands to, without their content")
	return c
}

// bundleShow assembles a bundle to out. With listOnly it instead prints the
// ordered fragment paths the bundle expands to (with @include resolved), one per
// line, and a count to errw.
func (a *App) bundleShow(out, errw io.Writer, name string, listOnly bool) error {
	if !a.e.HasBundle(name) {
		return fmt.Errorf("%s: not a known bundle (run moon bundle list to see them)", name)
	}
	if listOnly {
		frags, err := a.e.Expand(name)
		if err != nil {
			return err
		}
		for _, f := range frags {
			fmt.Fprintln(out, f)
		}
		fmt.Fprintf(errw, "(%d fragments)\n", len(frags))
		return nil
	}
	data, err := a.e.Assemble(name)
	if err != nil {
		return err
	}
	_, err = out.Write(data)
	return err
}
