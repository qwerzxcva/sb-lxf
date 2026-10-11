//go:build with_boxdd && (linux)

package main

import "path/filepath"

func normalizeRestrictedPath(path string) string {
	return filepath.Clean(path)
}
