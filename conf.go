package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type Filter struct {
	Matchers []Matcher
	Conf     MapAny
}

func NewFilter() *Filter {
	return &Filter{
		Conf: make(MapAny),
	}
}
func (f *Filter) Match(file string) bool {

	// Normal: OR
	// false, false -> false
	// false, true  -> true
	// true, false -> true
	// true, true -> true

	// Invert:
	// false, false -> false
	// false, true -> false
	// true, false -> true
	// true, true -> false
	use := false
	for _, m := range f.Matchers {
		if m.Match(file) {
			use = m.True()
		}
	}
	return use
}

type S3PushConfig struct {
	s3srv      *s3.S3
	s3uploader *s3manager.Uploader

	Base   string
	Bucket string
	Prefix string
	Region string

	Upload []*Filter
}

func (s *S3PushConfig) InitS3() error {
	// this check probably belongs elsewhere
	if s.Bucket == "" {
		return fmt.Errorf("S3 Bucket not specified")
	}
	log.Printf("Making session in region %s", s.Region)
	// The session the S3 Uploader will use
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(s.Region)})
	if err != nil {
		return err
	}
	// S3 service client the Upload manager will use.
	service := s3.New(sess)

	// Create an uploader with S3 client and default options
	uploader := s3manager.NewUploaderWithClient(service)

	s.s3srv = service
	s.s3uploader = uploader

	return nil
}
func (s *S3PushConfig) Match(file string) MapAny {
	var a MapAny
	for _, m := range s.Upload {
		if m.Match(file) {
			a = Merge(a, m.Conf)
		}
	}
	return a
}

func (c *S3PushConfig) ConfCall(args []string) error {
	switch args[0] {
	case "base":
		return RequireString1(args, c.setBase)
	case "bucket":
		return RequireString1(args, c.setBucket)
	case "prefix":
		return RequireString1(args, c.setPrefix)
	case "region":
		return RequireString1(args, c.setRegion)
	}
	return fmt.Errorf("unknown command %s", args[0])
}

func (c *S3PushConfig) ConfObject(args []string) (Dispatcher, error) {
	switch args[0] {
	case "upload":
		f := NewFilter()
		c.Upload = append(c.Upload, f)
		return f, nil
	}
	return nil, fmt.Errorf("Unknown object %s", args[0])
}

func (c *S3PushConfig) setBase(arg string) error {
	c.Base = arg
	return nil
}

func (c *S3PushConfig) setBucket(arg string) error {
	c.Bucket = arg
	return nil
}
func (c *S3PushConfig) setPrefix(arg string) error {
	c.Prefix = arg
	return nil
}
func (c *S3PushConfig) setRegion(arg string) error {
	c.Region = arg
	return nil
}

func (c *Filter) ConfCall(args []string) error {
	cmd := strings.ToLower(args[0])
	cmd = strings.Replace(cmd, "-", "", -1)
	cmd = strings.Replace(cmd, "_", "", -1)
	switch cmd {
	case "include":
		return c.setInclude(true, args)
	case "exclude":
		return c.setInclude(false, args)
	case "bucket", "cachecontrol", "contentdisposition", "contentencoding", "contentlanguage", "contenttype", "storageclass", "websiteredirectlocation":
		args[0] = cmd
		return c.setMap(args)
	}
	return fmt.Errorf("Filter unknown command %s %s", args[0], cmd)
}

func (c *Filter) ConfObject(args []string) (Dispatcher, error) {
	return nil, fmt.Errorf("No sub-objects")
}

func (c *Filter) setInclude(truth bool, args []string) error {
	matcher, err := NewGlobMatch(args, truth)
	if err != nil {
		return err
	}
	c.Matchers = append(c.Matchers, matcher)
	return nil
}

func (c *Filter) setMap(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("%s expected 1 arg, got %d", args[0], len(args)-1)
	}
	c.Conf[args[0]] = args[1]
	return nil
}
func ReadConf(filename string) (*S3PushConfig, error) {
	rawconf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	root := new(S3PushConfig)
	err = Parse(root, string(rawconf))
	return root, err
}
