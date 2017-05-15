package main

import (
	"encoding/json"
	"fmt"
	"log"

	"gopkg.in/yaml.v2"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type Matcher interface {
	Match(string) bool
	True() bool
	MarshalJSON() ([]byte, error)
}

type MatchAll struct{}

func (m MatchAll) Match(file string) bool {
	return true
}
func (m MatchAll) True() bool { return true }

func (m MatchAll) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", "match-all")), nil
}

type Filter struct {
	Matchers []Matcher
	Conf     MapAny
}

func (f *Filter) DefaultAll() {
	if len(f.Matchers) == 0 {
		f.Matchers = []Matcher{MatchAll{}}
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

type XFilter struct {
	FilterRaw MapAny `yaml:"filter"`
	Conf      MapAny `yaml:"conf"`
}

func (x *XFilter) Convert() (*Filter, error) {
	var matchers []Matcher

	for k, v := range x.FilterRaw {
		switch k {
		case "include":
			args, err := getStrings(v)
			if err != nil {
				return nil, fmt.Errorf("Filter %s: %s", k, err)
			}
			m, err := NewGlobMatch(args, true)
			if err != nil {
				return nil, fmt.Errorf("Filter %s: %s", k, err)
			}
			matchers = append(matchers, m)
			delete(x.FilterRaw, k)
		case "exclude":
			args, err := getStrings(v)
			if err != nil {
				return nil, fmt.Errorf("Filter %s: %s", k, err)
			}
			m, err := NewGlobMatch(args, false)
			if err != nil {
				return nil, fmt.Errorf("Filter %s: %s", k, err)
			}
			matchers = append(matchers, m)
			delete(x.FilterRaw, k)
		default:
			return nil, fmt.Errorf("Unknown filter name: %q", k)
		}
	}
	// if len(x.FilterRaw) != 0
	// then these lefts are unknown

	return &Filter{
		Matchers: matchers,
		Conf:     x.Conf,
	}, nil
}

type S3PushConfig struct {
	s3srv      *s3.S3
	s3uploader *s3manager.Uploader

	Base   string `yaml:"base"`
	Bucket string `yaml:"bucket"`
	Region string `yaml:"region"`

	Conf []Filter
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
	for _, m := range s.Conf {
		if m.Match(file) {
			a = Merge(a, m.Conf)
		}
	}
	return a
}

type XPush struct {
	Base   string    `yaml:"base"`
	Bucket string    `yaml:"bucket"`
	Region string    `yaml:"region"`
	Select []XFilter `yaml:"select"`
}

func (x *XPush) Convert() (*S3PushConfig, error) {
	s := make([]Filter, len(x.Select))
	for i, v := range x.Select {
		f, err := v.Convert()
		if err != nil {
			return nil, err
		}
		s[i] = *f
	}

	return &S3PushConfig{
		Base:   x.Base,
		Bucket: x.Bucket,
		Region: x.Region,
		Conf:   s,
	}, nil
}

func ReadConf(config string) (*S3PushConfig, error) {
	var obj XPush
	raw := []byte(config)
	if err := yaml.Unmarshal(raw, &obj); err != nil {
		return nil, fmt.Errorf("Unable to read: %s", err)
	}
	newobj, err := obj.Convert()
	if err != nil {
		return nil, fmt.Errorf("Unable to convert: %s", err)
	}
	if false {
		out, err := json.MarshalIndent(&newobj, "", "  ")
		if err != nil {
			log.Fatalf("Marshal fail: %s", err)
		}
		fmt.Println(string(out))
	}

	return newobj, nil
}

/*
	conf := newobj.Match("foobar.xgif")
	if conf == nil {
		log.Fatalf("Didn't find anything")
	}
	out, err = json.MarshalIndent(&conf, "", "  ")
	if err != nil {
		log.Fatalf("Marshal fail: %s", err)
	}
	fmt.Println(string(out))
*/
