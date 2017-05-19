package main

import (
	"fmt"
	"log"
	"os"
	"path"
)

// over-ridden via goreleaser
var version = "master"

// single upload, no channels, no nothing
func upload(s3conf *S3PushConfig, up FileStat) error {
	ma := s3conf.Match(up.Name)
	if ma == nil {
		return fmt.Errorf("Want to upload %q but got no HTTP conf", up.Name)
	}
	upi, err := NewUploadInput(ma)
	if err != nil {
		return fmt.Errorf("Error configuring %q: %s", up.Name, err)
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
	defer func() { _ = fd.Close() }()

	if err != nil {
		return fmt.Errorf("Unable to open %s", err)
	}
	upi.Body = fd
	// TODO: UploaderOutput
	_, err = s3conf.s3uploader.Upload(upi)
	if err != nil {
		return fmt.Errorf("Unable to upload %s: %s", up.Name, err)
	}
	return nil
}

func worker(id int, conf *S3PushConfig, jobs <-chan FileStat, results chan<- error) {
	for up := range jobs {
		err := upload(conf, up)
		if err != nil {
			log.Printf("worker: %d: finished", id)
		} else {
			log.Printf("worker %d errors: %s", id, err)
		}
		results <- err
	}
}

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

	jobs := make(chan FileStat, len(uploads))
	results := make(chan error, len(uploads))

	// unlease some workers
	for w := 0; w < 1; w++ {
		go worker(w, s3conf, jobs, results)
	}

	// push into queue
	for _, up := range uploads {
		jobs <- up
	}

	// all done
	close(jobs)

	// get results
	errcount := 0
	for a := 0; a < len(uploads); a++ {
		err := <-results
		if err != nil {
			errcount++
		}
	}
	log.Printf("Total Errors: %d", errcount)
}
