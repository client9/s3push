package main

import (
	"log"
	"os"
	"path"
)

func main() {
	confName := ".s3push.sh"
	s3conf, err := ReadConf(confName)
	if err != nil {
		log.Fatalf("Unable to read conf: %s", err)
	}
	if err = s3conf.InitS3(); err != nil {
		log.Fatalf("Unable to init s3: %s", err)
	}
	localFiles, remoteFiles, err := files(s3conf)
	if err != nil {
		log.Fatalf("Unable to process files: %s", err)
	}
	uploads, downloads := compute(localFiles, remoteFiles)
	log.Printf("Uploads: %d", len(uploads))
	log.Printf("Downloads: %d", len(downloads))

	for _, up := range uploads {
		ma := s3conf.Match(up.Name)
		if ma == nil {
			log.Printf("Want to upload %q but got no HTTP conf", up.Name)
			continue
		}
		upi, err := NewUploadInput(ma)
		if err != nil {
			log.Fatalf("Error configuring %q: %s", up.Name, err)
		}
		upi.Bucket = &s3conf.Bucket
		key := path.Join("", up.Name)
		upi.Key = &key
		if upi.ContentType == nil && up.Mime != "" {
			upi.ContentType = &up.Mime
		}
		log.Printf("%s: configuring: %s", up.Name, DumpUploadInput(upi))
		log.Printf("PATH= %s", up.Path)
		fd, err := os.Open(up.Path)
		defer fd.Close()

		if err != nil {
			log.Printf("Unable to open %s", err)
			continue
		}
		upi.Body = fd
		result, err := s3conf.s3uploader.Upload(upi)
		if err != nil {
			log.Printf("Unable to upload %s: %s", up.Name, err)
		}
		log.Printf("Uploaded %s", result.Location)
	}
}
