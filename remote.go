package main

import (
	"fmt"
	"log"
	"mime"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	//	"github.com/client9/s3push/httpmime"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

type FileStat struct {
	Name         string
	Path         string
	ETag         string
	LastModified time.Time
	Size         int64
	Mime         string
}

func loadLocalFiles(basePath string) ([]FileStat, error) {
	out := []FileStat{}
	basePath = filepath.ToSlash(basePath)

	stat, err := os.Stat(basePath)
	if err != nil {
		return nil, err
	}

	absPath, err := filepath.Abs(basePath)
	if err != nil {
		return nil, err
	}

	if !stat.IsDir() {
		var sum string
		sum, err = MD5File(absPath)
		if err != nil {
			return nil, err
		}
		name := filepath.Base(basePath)
		out = append(out, FileStat{
			Name:         name,
			Path:         absPath,
			LastModified: stat.ModTime(),
			Size:         stat.Size(),
			ETag:         sum,
			Mime:         mime.TypeByExtension(path.Ext(name)),
		})
		return out, nil
	}

	err = filepath.Walk(basePath, func(filePath string, stat os.FileInfo, err error) error {
		relativePath, err := filepath.Rel(basePath, filepath.ToSlash(filePath))
		if err != nil {
			return err
		}

		// if skip
		//	return filepath.SkipDir
		if stat == nil || stat.IsDir() {
			return nil
		}
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			return err
		}
		sum, err := MD5File(absPath)
		if err != nil {
			return err
		}
		out = append(out, FileStat{
			Name:         relativePath,
			Path:         absPath,
			LastModified: stat.ModTime(),
			Size:         stat.Size(),
			ETag:         sum,
			Mime:         mime.TypeByExtension(path.Ext(relativePath)),
		})
		return nil
	})
	if err != nil {
		return out, nil
	}

	return out, nil
}

func listS3Files(svc *s3.S3, bucket, bucketPrefix string) ([]FileStat, error) {
	var token *string
	awsBucket := aws.String(bucket)
	awsBucketPrefix := aws.String(bucketPrefix)
	out := []FileStat{}

	log.Printf("Querying S3 for bucket/prefix of %s/%s", bucket, bucketPrefix)
	conf := &s3.ListObjectsV2Input{
		Bucket: awsBucket,
		Prefix: awsBucketPrefix,
	}
	count := 0
	for {
		count++
		conf.ContinuationToken = token
		list, err := svc.ListObjectsV2(conf)
		if err != nil {
			return nil, err
		}
		log.Printf("Iteration %d got %d items", count, len(list.Contents))
		for _, object := range list.Contents {
			out = append(out, FileStat{
				Name:         strings.TrimPrefix(*object.Key, bucketPrefix+"/"),
				Path:         *object.Key,
				ETag:         *object.ETag,
				Size:         *object.Size,
				LastModified: *object.LastModified,
			})
		}
		token = list.NextContinuationToken
		if token == nil {
			break
		}
	}

	return out, nil
}

func files(conf *S3PushConfig) ([]FileStat, []FileStat, error) {

	bucketPrefix := ""
	base := conf.Base

	var remoteFiles []FileStat
	var remoteErr error
	var localFiles []FileStat
	var localErr error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		t0 := time.Now()
		remoteFiles, remoteErr = listS3Files(conf.s3srv, conf.Bucket, bucketPrefix)
		log.Printf("Remote Check: %s", time.Since(t0))
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		t0 := time.Now()
		localFiles, localErr = loadLocalFiles(base)
		log.Printf("Local Check: %s", time.Since(t0))
	}()
	wg.Wait()

	if localErr != nil {
		return nil, nil, fmt.Errorf("local filesystem error: %s", localErr)
	}
	if remoteErr != nil {
		return nil, nil, fmt.Errorf("remote filesystem error: %s", remoteErr)
	}
	return localFiles, remoteFiles, nil
}

// give a list of local files and remote files, compute
// what gets uploaded or deleted.
// (technically the deleted files could be downloaded for a sync)
func compute(localFiles, remoteFiles []FileStat) ([]FileStat, []FileStat) {
	fmap := make(map[string]FileStat)
	for _, f := range remoteFiles {
		fmap[f.Name] = f
	}

	uploads := []FileStat{}
	for _, f := range localFiles {
		remoteStat, ok := fmap[f.Name]
		if !ok {
			fmt.Printf("NEW LOCAL FILE: %s\n", f.Name)
			uploads = append(uploads, f)
			continue

		}
		delete(fmap, f.Name)
		if f.ETag != remoteStat.ETag {
			// log.Printf("ETag diff for %s: %s vs %s", f.Name, f.ETag, remoteStat.ETag)
			uploads = append(uploads, f)
		}
	}

	// convert back to values
	deletes := make([]FileStat, 0, len(fmap))
	for _, v := range fmap {
		deletes = append(deletes, v)
	}

	fmt.Printf("Total S3 files      : %d\n", len(remoteFiles))
	fmt.Printf("Total Local files   : %d\n", len(localFiles))
	fmt.Printf("Total uploads       : %d\n", len(uploads))
	fmt.Printf("Files to be deleted : %d\n", len(deletes))

	return uploads, deletes
}
