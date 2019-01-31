package xbcommon

import (
	"path/filepath"
	"strings"
)

func FilePath(dir, name string, args ...string) string {
	canonicalName := strings.Replace(name, "\\", "/", -1)
	parts := append([]string{dir}, strings.Split(canonicalName, "/")...)
	parts = append(parts, args...)
	return filepath.Join(parts...)
}
