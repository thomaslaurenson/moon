package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

func (a *App) newListCmd() *cobra.Command {
	var long, asJSON bool
	c := &cobra.Command{
		Use:   "list",
		Short: "List available bundles",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return a.list(cmd.OutOrStdout(), long, asJSON)
		},
	}
	c.Flags().BoolVarP(&long, "long", "l", false, "show each bundle's one-line description")
	c.Flags().BoolVar(&asJSON, "json", false, "output as structured JSON")
	return c
}

// list writes the available bundle names to out; long and asJSON add per-bundle
// descriptions as aligned columns or structured JSON respectively.
func (a *App) list(out io.Writer, long, asJSON bool) error {
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
