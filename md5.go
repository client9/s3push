package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
)

func MD5Reader(f io.Reader) (string, error) {
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("\"%x\"", h.Sum(nil)), nil
}

func MD5File(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()
	return MD5Reader(f)
}
