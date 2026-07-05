package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is moon's version, injected at build time via
// -ldflags "-X github.com/thomaslaurenson/moon/cmd.Version=...". Defaults to "dev"
// for a plain `go build` or `go run`.
var Version = "dev"

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the moon version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), Version)
			return nil
		},
	}
}
