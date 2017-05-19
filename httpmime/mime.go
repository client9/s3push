package httpmime

import (
	"mime"
	"os"
)

var typeFiles = []string{
	"/etc/mime.types",
	"/etc/apache2/mime.types",
	"/etc/apache/mime.types",
}

func Init() {
	for k, v := range mimeTypes {
		_ = mime.AddExtensionType(k, v)
	}
}

func MimeTypeFiles() []string {
	out := []string{}
	for _, fname := range typeFiles {
		if _, err := os.Stat(fname); err == nil {
			out = append(out, fname)
		}
	}
	return out
}
