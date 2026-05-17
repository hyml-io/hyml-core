package utilities

import (
	"path/filepath"
	"strings"
)

type StringHandler struct{}

func (stringHandler StringHandler) ContainsString(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

func (stringHandler StringHandler) GetFilenameWithoutExtension(path string) string {
	filename := filepath.Base(path)
	extension := filepath.Ext(filename)

	if extension != "" {
		return strings.TrimSuffix(filename, extension)
	}
	return filename
}

func (stringHandler StringHandler) FilterBy(values []string, filter string) []string {
	var filtered []string
	for _, s := range values {
		if strings.HasPrefix(s, filter) {
			filtered = append(filtered, s)
		}
	}
	return filtered
}
