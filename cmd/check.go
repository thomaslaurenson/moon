package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

func (a *App) newCheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "Validate every bundle: missing fragments, include cycles, orphans",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ok, err := a.check(cmd.ErrOrStderr())
			if err != nil {
				return err
			}
			if !ok {
				return ErrSilent
			}
			return nil
		},
	}
}

// check validates every bundle, printing FAIL/WARN lines and a summary to errw.
// It reports ok=false (rather than an error) when validation itself succeeded
// but found problems, since that's a normal check-failed outcome, not a fault.
func (a *App) check(errw io.Writer) (ok bool, err error) {
	problems, orphans, err := a.e.Check()
	if err != nil {
		return false, err
	}
	names, err := a.e.List()
	if err != nil {
		return false, err
	}
	for _, p := range problems {
		fmt.Fprintf(errw, "  FAIL %s\n", p)
	}
	for _, o := range orphans {
		fmt.Fprintf(errw, "  WARN orphan fragment (in no bundle): %s\n", o)
	}
	fmt.Fprintf(errw, "[*] checked %d bundle(s): %d problem(s), %d orphan(s)\n", len(names), len(problems), len(orphans))
	return len(problems) == 0, nil
}
