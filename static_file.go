package core

import (
	"path/filepath"
	"strings"
)

/*
* RegisterFolder registers a folder to a url
* @param url string
* @param prefix string
* @param path string
* @return void
* @example RegisterFolder("/static/", "/static/", "./static")
 */
func RegisterFolder(url string, path string) {
	LogInfo("Register folder: url = %s, path = %s", url, path)

	router.Static(url, sanitizeFilePath(path))
}

func sanitizeFilePath(filename string) string {
	// Remove any path traversal attempts
	return filepath.Clean(strings.Replace(filename, "..", "", -1))
}
