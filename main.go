package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"
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
		return fmt.Errorf("error configuring %q: %s", up.Name, err)
	}
	upi.Bucket = &s3conf.Bucket
	key := path.Join("", up.Name)
	upi.Key = &key
	if upi.ContentType == nil && up.Mime != "" {
		upi.ContentType = &up.Mime
	}
	//log.Printf("%s: configuring: %s", up.Name, DumpUploadInput(upi))
	fd, err := os.Open(up.Path)
	defer func() { _ = fd.Close() }()

	if err != nil {
		return fmt.Errorf("unable to open %s", err)
	}
	upi.Body = fd
	// TODO: UploaderOutput
	_, err = s3conf.s3uploader.Upload(upi)
	if err != nil {
		return fmt.Errorf("unable to upload %s: %s", up.Name, err)
	}
	return nil
}

func worker(id int, conf *S3PushConfig, jobs <-chan FileStat, results chan<- error) {
	for up := range jobs {
		log.Printf("upload: %s", up.Name)
		err := upload(conf, up)
		if err != nil {
			log.Printf("%d error: %s", id, err)
		}
		results <- err
	}
}

var (
	confName = kingpin.Flag("conf", "Location of configuration file").Short('c').Default(".s3push.sh").OverrideDefaultFromEnvar("S3PUSH_CONF").String()
	bucket = kingpin.Flag("bucket", "S3 bucket name").Short('b').OverrideDefaultFromEnvar("S3PUSH_BUCKET").String()
	region = kingpin.Flag("region", "S3 bucket region").Short('r').OverrideDefaultFromEnvar("S3PUSH_REGION").String()
)

func main() {
	kingpin.Version(version)
	kingpin.Parse()

	s3conf, err := ReadConf(*confName)
	if err != nil {
		log.Fatalf("Unable to read conf: %s", err)
	}
	if *bucket != "" {
		s3conf.Bucket = *bucket
	}
	if *region != "" {
		s3conf.Region = *region
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

	t0 := time.Now()
	jobs := make(chan FileStat, len(uploads))
	results := make(chan error, len(uploads))

	// unlease some workers
	for w := 0; w < 8; w++ {
		go worker(w, s3conf, jobs, results)
	}

	// push into queue
	upbytes := int64(0)
	for _, up := range uploads {
		upbytes += up.Size
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
	seconds := time.Since(t0).Seconds()
	mbytes := float64(upbytes) / float64(1024*1024)

	log.Printf("Total Errors : %d", errcount)
	log.Printf("Total Bytes  : %d", upbytes)
	log.Printf("Total Time   : %s", time.Since(t0))
	log.Printf("MB/Sec       : %f", mbytes/seconds)

}
