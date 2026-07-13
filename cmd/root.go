// Package cmd wires up moon's subcommands with Cobra. It handles argument
// parsing and dispatch only; the actual bundling logic lives in internal/bundler.
package cmd

import (
	"errors"
	"fmt"
	"io"
	"io/fs"

	"github.com/spf13/cobra"

	"github.com/thomaslaurenson/moon/internal/bundler"
)

const rootLong = `moon composes agent-instruction bundles from markdown fragments.

New here? Run this sequence:
  moon list --long          see every bundle with a one-line description
  moon recipe <name>        see the exact fragments a bundle expands to
  moon show <name>          print it (or a single fragment path) to stdout`

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
		a.newListCmd(),
		a.newShowCmd(),
		a.newBuildCmd(),
		a.newRecipeCmd(),
		a.newCheckCmd(),
		a.newInitCmd(),
		newVersionCmd(),
	)
	return root
}

// resolveContent returns the assembled content for a bundle name, or the raw
// content of a fragment path if no bundle matches. Bundles are checked first so
// existing bundle names never change meaning.
func (a *App) resolveContent(name string) ([]byte, error) {
	switch {
	case a.e.HasBundle(name):
		return a.e.Assemble(name)
	case a.e.HasFragment(name):
		return a.e.Fragment(name)
	default:
		return nil, fmt.Errorf("%s: not a known bundle or fragment (run moon list to see bundles)", name)
	}
}
