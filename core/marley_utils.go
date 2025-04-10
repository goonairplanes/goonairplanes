

package core

import (
	"path/filepath"
	"strings"
)


func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}


func getRoutePathFromFile(fullPath, basePath string) string {
	fullPath = filepath.ToSlash(fullPath)
	basePath = filepath.ToSlash(basePath)

	if !strings.HasSuffix(basePath, "/") {
		basePath += "/"
	}

	relativePath := fullPath
	if strings.HasPrefix(fullPath, basePath) {
		relativePath = strings.TrimPrefix(fullPath, basePath)
	}

	relativePath = strings.TrimSuffix(relativePath, ".html")

	if relativePath == "index" {
		return "/"
	} else if strings.HasSuffix(relativePath, "/index") {
		relativePath = strings.TrimSuffix(relativePath, "/index")
		if relativePath == "" {
			return "/"
		}
	}

	if relativePath != "/" && !strings.HasPrefix(relativePath, "/") {
		relativePath = "/" + relativePath
	}

	return relativePath
}
