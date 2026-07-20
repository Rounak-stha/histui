// Package pathfilter matches repository-relative paths against ignore patterns.
package pathfilter

import (
	"path"
	"strings"
)

// Match reports whether a repository-relative path matches one of the patterns.
// Patterns use forward slashes. A plain directory name or a pattern ending in /
// matches the complete subtree. ** matches zero or more path segments.
// A pattern without a slash is also matched against each path's basename.
func Match(filePath string, patterns []string) bool {
	filePath = normalize(filePath)
	if filePath == "" {
		return false
	}
	for _, raw := range patterns {
		pattern := normalizePattern(raw)
		if pattern == "" {
			continue
		}

		hasMeta := strings.ContainsAny(pattern, "*?[")
		if !hasMeta {
			if filePath == pattern || strings.HasPrefix(filePath, pattern+"/") {
				return true
			}
			if !strings.Contains(pattern, "/") && path.Base(filePath) == pattern {
				return true
			}
			continue
		}

		if globMatch(strings.Split(pattern, "/"), strings.Split(filePath, "/")) {
			return true
		}
		if !strings.Contains(pattern, "/") {
			matched, err := path.Match(pattern, path.Base(filePath))
			if err == nil && matched {
				return true
			}
		}
	}
	return false
}

func normalize(value string) string {
	value = strings.ReplaceAll(strings.TrimSpace(value), "\\", "/")
	value = strings.TrimPrefix(value, "./")
	value = strings.Trim(value, "/")
	if value == "." {
		return ""
	}
	return value
}

func normalizePattern(value string) string {
	value = strings.ReplaceAll(strings.TrimSpace(value), "\\", "/")
	value = strings.TrimPrefix(value, "./")
	// A trailing slash explicitly denotes a directory subtree.
	if strings.HasSuffix(value, "/") {
		value = strings.TrimRight(value, "/") + "/**"
	}
	return strings.Trim(value, "/")
}

func globMatch(pattern, value []string) bool {
	if len(pattern) == 0 {
		return len(value) == 0
	}
	if pattern[0] == "**" {
		// ** can consume no segments or one segment and remain active.
		return globMatch(pattern[1:], value) || (len(value) > 0 && globMatch(pattern, value[1:]))
	}
	if len(value) == 0 {
		return false
	}
	matched, err := path.Match(pattern[0], value[0])
	return err == nil && matched && globMatch(pattern[1:], value[1:])
}
