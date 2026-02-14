package main

import "fmt"

// Version follows brimdata/super release versions (major.minor.patch)
// See: https://github.com/brimdata/super/releases
const Version = "0.1.0"

// LSPPatch is incremented for LSP-specific bug fixes between syncs
// Reset to 0 when syncing to a new upstream version
const LSPPatch = 0

// SuperCommit is the brimdata/super commit SHA this version is synced to
// Updated by /sync command
const SuperCommit = "e8764da"

// FullVersion returns the complete version string
// Format: <super-version>.<lsp-patch>+<commit-sha>
// Example: 0.1.0.0+e8764da
func FullVersion() string {
	v := fmt.Sprintf("%s.%d", Version, LSPPatch)
	if SuperCommit != "" {
		return v + "+" + SuperCommit
	}
	return v
}
