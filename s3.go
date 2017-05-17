package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/s3"
	//	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type s3Setter func(s *s3manager.UploadInput, str string) error

func setBucket(s *s3manager.UploadInput, str string) error {
	s.Bucket = &str
	return nil
}
func setCacheControl(s *s3manager.UploadInput, str string) error {
	s.CacheControl = &str
	return nil
}
func setContentDisposition(s *s3manager.UploadInput, str string) error {
	s.ContentDisposition = &str
	return nil
}
func setContentEncoding(s *s3manager.UploadInput, str string) error {
	s.ContentEncoding = &str
	return nil
}
func setContentLanguage(s *s3manager.UploadInput, str string) error {
	s.ContentLanguage = &str
	return nil
}
func setContentType(s *s3manager.UploadInput, str string) error {
	s.ContentType = &str
	return nil
}
func setStorageClass(s *s3manager.UploadInput, str string) error {
	str = strings.ToUpper(str)
	str = strings.Replace(str, "-", "_", -1)
	switch str {
	case s3.ObjectStorageClassStandard:
		s.StorageClass = nil
		return nil
	case s3.ObjectStorageClassReducedRedundancy, s3.ObjectStorageClassGlacier:
		s.StorageClass = &str
		return nil
	}
	return fmt.Errorf("Unknown storage class of %q", str)
}
func setWebsiteRedirectLocation(s *s3manager.UploadInput, str string) error {
	s.WebsiteRedirectLocation = &str
	return nil
}

var configMap = map[string]s3Setter{
	"bucket":                  setBucket,
	"cachecontrol":            setCacheControl,
	"contentdisposition":      setContentDisposition,
	"contentencoding":         setContentEncoding,
	"contentlanguage":         setContentLanguage,
	"contenttype":             setContentType,
	"storageclass":            setStorageClass,
	"websiteredirectlocation": setWebsiteRedirectLocation,
}

// NewUploadInput creates an s3manager.UploadInput or error
// Entries that are used from MapAny are deleted
// (so unhandled keys in input remain)
func NewUploadInput(a MapAny) (*s3manager.UploadInput, error) {
	upload := &s3manager.UploadInput{}
	for k, v := range a {
		newk := strings.ToLower(k)
		newk = strings.Replace(newk, "-", "", -1)
		newk = strings.Replace(newk, "_", "", -1)
		fn, ok := configMap[newk]
		if !ok {
			return nil, fmt.Errorf("Unknown s3 config %q", k)
		}
		delete(a, k)
		str, ok := v.(string)
		if ok && str != "" {
			err := fn(upload, str)
			if err != nil {
				return upload, fmt.Errorf("Error setting %q: %s", k, err)
			}
		}
	}
	return upload, nil
}

// DumpUploadInput is a crap debug function
func DumpUploadInput(obj *s3manager.UploadInput) string {
	out, _ := json.MarshalIndent(obj, "", "  ")
	str := string(out)
	lines := strings.Split(str, "\n")
	newlines := []string{}
	for _, line := range lines {
		if !strings.Contains(line, "null") {
			newlines = append(newlines, line)
		}
	}
	return strings.Join(newlines, "\n")
}
