package httpmime

import (
	"mime"
	"os"
)

// TypeFiles is the list of files used by golang to load mime types.
//
// See https://golang.org/pkg/mime/#TypeByExtension
//
var TypeFiles = []string{
	"/etc/mime.types",
	"/etc/apache2/mime.types",
	"/etc/apache/mime.types",
}

// MimeTypeFiles returns a list of default mime files found
// by golang's native loader.  This is done by inspection
// and may or may not represent what was actually loaded or
// if what order.
//
func MimeTypeFiles() []string {
	out := []string{}
	for _, fname := range TypeFiles {
		if _, err := os.Stat(fname); err == nil {
			out = append(out, fname)
		}
	}
	return out
}

// IsLoaded returns true if the internal golang mapping
// appears to be sane.
func IsLoaded() bool {
	// golang's internal table does not include .txt
	// if this is missing something is wrong.
	return mime.TypeByExtension(".txt") != ""
}

// AddDefaults adds a standard set of extensions and mime-types to
//  golang's mime package.  This may over-ride existing configuration.
func AddDefaults() {
	for k, v := range mimeTypes {
		_ = mime.AddExtensionType(k, v)
	}
}

// Init attempts to make sure golang's mime package is initialized
// correctly on platforms missing basic mime type files.
//
// Returns true if new defaults were added.
// Returns false if no modification
func Init() bool {
	if !IsLoaded() {
		AddDefaults()
		return true
	}
	return false
}
