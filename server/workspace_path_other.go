//go:build !windows

package server

import (
	"path/filepath"
	"strings"
)

// resolveRealPath returns the final path of p with every symlink resolved.
func resolveRealPath(p string) (string, bool) {
	resolved, err := filepath.EvalSymlinks(filepath.Clean(p))
	if err != nil {
		return "", false
	}
	return resolved, true
}

// pathInside reports whether child is base or below it.
func pathInside(child, base string) bool {
	return child == base || strings.HasPrefix(child, base+string(filepath.Separator))
}
