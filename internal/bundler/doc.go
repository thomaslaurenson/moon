// Package bundler resolves bundle definitions and assembles instruction bundles
// from a fragment tree.
//
// A bundle definition (src/bundles/<name>) is an ordered list of fragment paths
// relative to src/fragments. Blank lines and content after '#' are ignored. A line
// "@include <bundle>" expands another bundle in place, so bundles share a common base.
package bundler
