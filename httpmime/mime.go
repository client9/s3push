package httpmime

import "strings"

// https://golang.org/src/mime/type.go
func TypeByExtension(ext string) string {
	// Case-sensitive lookup.
	if v := mimeTypes[ext]; v != "" {
		return v
	}

	// Case-insensitive lookup.
	// Optimistically assume a short ASCII extension and be
	// allocation-free in that case.
	var buf [10]byte
	lower := buf[:0]
	const utf8RuneSelf = 0x80 // from utf8 package, but not importing it.
	for i := 0; i < len(ext); i++ {
		c := ext[i]
		if c >= utf8RuneSelf {
			// Slow path.
			return mimeTypes[strings.ToLower(ext)]
		}
		if 'A' <= c && c <= 'Z' {
			lower = append(lower, c+('a'-'A'))
		} else {
			lower = append(lower, c)
		}
	}
	// The conversion from []byte to string doesn't allocate in
	// a map lookup.
	return mimeTypes[string(lower)]
}
