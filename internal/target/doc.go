// Package target computes the files a moon init target (claude, copilot, ...)
// writes for a set of resolved bundles. Plan is pure: it never touches disk, which
// keeps it fully unit-testable. The caller (cmd) is responsible for actually
// writing the returned files.
package target
