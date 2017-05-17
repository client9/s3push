package main

import (
	"fmt"
)

// Matcher defines an interface for filematchers
//
type Matcher interface {
	Match(string) bool
	True() bool
	MarshalText() ([]byte, error)
}

// MatchAll is a example
type MatchAll struct{}

func (m MatchAll) Match(file string) bool {
	return true
}
func (m MatchAll) True() bool { return true }

func (m MatchAll) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", "match-all")), nil
}
