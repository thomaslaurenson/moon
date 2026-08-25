// Command moon composes agent-instruction bundles from embedded markdown fragments.
package main

import (
	"embed"
	"errors"
	"fmt"
	"os"

	"github.com/thomaslaurenson/moon/cmd"
)

// content embeds the fragment tree (src/fragments) and bundle definitions
// (src/bundles) into the binary. The all: prefix is required so files beginning
// with '_' (such as _core.md) are included.
//
//go:embed all:src
var content embed.FS

func main() {
	root := cmd.NewRootCmd(content, os.Stdout, os.Stderr)
	if err := root.Execute(); err != nil {
		var ec *cmd.ExitCodeError
		if errors.As(err, &ec) {
			os.Exit(ec.Code)
		}
		fmt.Fprintf(os.Stderr, "moon: %v\n", err)
		os.Exit(1)
	}
}
