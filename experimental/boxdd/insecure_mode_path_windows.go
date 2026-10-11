//go:build with_boxdd && (windows)

package main

import (
	"path/filepath"
	"strings"
)

func normalizeRestrictedPath(path string) string {
	return strings.ToLower(filepath.Clean(path))
}
