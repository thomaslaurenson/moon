package cmd

import "github.com/spf13/cobra"

func (a *App) newShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <bundle|fragment>",
		Short: "Assemble a bundle, or print a single fragment, to stdout",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := a.resolveContent(args[0])
			if err != nil {
				return err
			}
			_, err = cmd.OutOrStdout().Write(data)
			return err
		},
	}
}
