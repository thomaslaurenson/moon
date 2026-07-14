// Package cmd wires up moon's subcommands with Cobra. It handles argument
// parsing and dispatch only; the actual bundling logic lives in internal/bundler.
package cmd

import (
	"errors"
	"io"
	"io/fs"

	"github.com/spf13/cobra"

	"github.com/thomaslaurenson/moon/internal/bundler"
)

const rootLong = `moon composes agent-instruction bundles from markdown fragments.

A fragment is a single markdown file (src/fragments). A bundle is a named
composition of fragments (src/bundles). Use "moon fragment" to work with the
individual files, and "moon bundle" to compose them.

New here? Run this sequence:
  moon bundle list --long     see every bundle with a one-line description
  moon bundle show <name> -l  see the exact fragments a bundle expands to
  moon bundle show <name>     print the assembled bundle to stdout`

// ErrSilent signals that a command already printed everything the user needs to
// see; the entry point should exit non-zero without adding its own error line.
var ErrSilent = errors.New("")

// App holds the dependencies shared by every subcommand.
type App struct {
	e *bundler.Engine
}

// NewRootCmd builds moon's command tree backed by fsys (which must contain src/
// and bundles/), writing output to out and errw.
func NewRootCmd(fsys fs.FS, out, errw io.Writer) *cobra.Command {
	a := &App{e: bundler.New(fsys)}

	root := &cobra.Command{
		Use:           "moon",
		Short:         "Compose agent-instruction bundles from fragments",
		Long:          rootLong,
		SilenceErrors: true,
		SilenceUsage:  true,
		Version:       Version,
	}
	root.SetOut(out)
	root.SetErr(errw)

	root.AddCommand(
		a.newFragmentCmd(),
		a.newBundleCmd(),
		a.newCheckCmd(),
		a.newInitCmd(),
		newVersionCmd(),
	)
	return root
}
