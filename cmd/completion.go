package cmd

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/thomaslaurenson/moon/internal/target"
)

// completeBundles offers bundle names when completing a bundle <name> argument.
func (a *App) completeBundles(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	names, err := a.e.List()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	return filterByPrefix(names, toComplete), cobra.ShellCompDirectiveNoFileComp
}

// completeFragments offers fragment paths when completing a fragment <path> argument.
func (a *App) completeFragments(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	paths, err := a.e.ListFragments()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	return filterByPrefix(paths, toComplete), cobra.ShellCompDirectiveNoFileComp
}

// completeInit completes an init invocation: the target name first, then bundle
// names for the optional trailing bundle arguments.
func (a *App) completeInit(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) == 0 {
		return filterByPrefix(target.Names(), toComplete), cobra.ShellCompDirectiveNoFileComp
	}
	names, err := a.e.List()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	return filterByPrefix(names, toComplete), cobra.ShellCompDirectiveNoFileComp
}

// filterByPrefix returns the candidates that start with prefix, preserving order.
// An empty prefix returns every candidate.
func filterByPrefix(candidates []string, prefix string) []string {
	if prefix == "" {
		return candidates
	}
	var out []string
	for _, c := range candidates {
		if strings.HasPrefix(c, prefix) {
			out = append(out, c)
		}
	}
	return out
}
