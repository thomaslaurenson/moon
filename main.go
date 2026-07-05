// Command moon composes agent-instruction bundles from embedded markdown fragments.
package main

import (
	"embed"
	"errors"
	"fmt"
	"os"

	"github.com/thomaslaurenson/moon/cmd"
)

// content embeds the fragment tree (src/) and recipes (bundles/) into the binary.
// The all: prefix is required so files beginning with '_' (such as _core.md) are included.
//
//go:embed all:src all:bundles
var content embed.FS

func main() {
	root := cmd.NewRootCmd(content, os.Stdout, os.Stderr)
	if err := root.Execute(); err != nil {
		if !errors.Is(err, cmd.ErrSilent) {
			fmt.Fprintf(os.Stderr, "moon: %v\n", err)
		}
		os.Exit(1)
	}
}
