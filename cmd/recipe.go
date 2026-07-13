package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

func (a *App) newRecipeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "recipe <bundle|fragment>",
		Short: "Show the fragments a bundle expands to, without their content",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.recipe(cmd.OutOrStdout(), cmd.ErrOrStderr(), args[0])
		},
	}
}

func (a *App) recipe(out, errw io.Writer, name string) error {
	switch {
	case a.e.HasBundle(name):
		frags, err := a.e.Recipe(name)
		if err != nil {
			return err
		}
		fmt.Fprintln(out, name)
		for _, f := range frags {
			fmt.Fprintf(out, "  %s\n", f)
		}
		fmt.Fprintf(errw, "(%d fragments)\n", len(frags))
		return nil
	case a.e.HasFragment(name):
		fmt.Fprintln(out, name)
		fmt.Fprintln(errw, "(a single fragment, not a bundle)")
		return nil
	default:
		return fmt.Errorf("%s: not a known bundle or fragment (run moon list to see bundles)", name)
	}
}
