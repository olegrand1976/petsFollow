package petdossier

import (
	"path"
	"strings"
)

// DocumentZipPath builds documents/{title}.{ext} for the share pack.
// Ext comes from fileName (or objectKey fallback); title is sanitized.
func DocumentZipPath(title, fileName, objectKey string) string {
	title = strings.TrimSpace(title)
	fileName = strings.TrimSpace(fileName)
	ext := path.Ext(fileName)
	if ext == "" {
		ext = path.Ext(path.Base(strings.TrimSpace(objectKey)))
	}
	if ext == "" {
		ext = ".bin"
	}
	base := title
	if base == "" {
		base = strings.TrimSuffix(path.Base(fileName), path.Ext(fileName))
	}
	if base == "" {
		base = strings.TrimSuffix(path.Base(objectKey), path.Ext(path.Base(objectKey)))
	}
	if base == "" {
		base = "document"
	}
	// Drop accidental extension already present in the title.
	if strings.EqualFold(path.Ext(base), ext) {
		base = strings.TrimSuffix(base, path.Ext(base))
	}
	name := sanitizeZipName(base + ext)
	if name == "" {
		name = "document" + ext
	}
	return "documents/" + name
}
