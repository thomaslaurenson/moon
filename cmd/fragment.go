package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
)

// newFragmentCmd groups the verbs that operate on individual fragments: the
// atomic markdown files under src/fragments that bundles are composed from.
func (a *App) newFragmentCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "fragment",
		Short: "List or print individual fragments (single markdown files)",
	}
	c.AddCommand(a.newFragmentListCmd(), a.newFragmentShowCmd())
	return c
}

func (a *App) newFragmentListCmd() *cobra.Command {
	var asJSON bool
	c := &cobra.Command{
		Use:   "list [filter]",
		Short: "List fragment paths, optionally filtered by a substring",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var filter string
			if len(args) == 1 {
				filter = args[0]
			}
			return a.fragmentList(cmd.OutOrStdout(), filter, asJSON)
		},
	}
	c.Flags().BoolVar(&asJSON, "json", false, "output as structured JSON")
	return c
}

// fragmentList writes fragment paths to out, keeping only those containing filter
// when it is non-empty. asJSON emits a JSON array instead of one path per line.
func (a *App) fragmentList(out io.Writer, filter string, asJSON bool) error {
	paths, err := a.e.ListFragments()
	if err != nil {
		return err
	}
	if filter != "" {
		var kept []string
		for _, p := range paths {
			if strings.Contains(p, filter) {
				kept = append(kept, p)
			}
		}
		paths = kept
	}
	if asJSON {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(paths)
	}
	for _, p := range paths {
		fmt.Fprintln(out, p)
	}
	return nil
}

func (a *App) newFragmentShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:               "show <path>",
		Short:             "Print a single fragment to stdout",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: a.completeFragments,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !a.e.HasFragment(args[0]) {
				return fmt.Errorf("%s: not a known fragment (run moon fragment list to see them)", args[0])
			}
			data, err := a.e.Fragment(args[0])
			if err != nil {
				return err
			}
			_, err = cmd.OutOrStdout().Write(data)
			return err
		},
	}
}
