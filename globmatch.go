package main

import (
	"fmt"

	"github.com/gobwas/glob"
)

// GlobMatch handle glob matching
type GlobMatch struct {
	orig    []string
	matcher []glob.Glob
	normal  bool
}

// NewGlobMatch compiles a new matcher.
// Arg true should be set to false if the output is inverted.
func NewGlobMatch(args []string, normal bool) (*GlobMatch, error) {
	matchers := make([]glob.Glob, len(args))
	for i, arg := range args {
		g, err := glob.Compile(arg)
		if err != nil {
			return nil, err
		}
		matchers[i] = g
	}
	return &GlobMatch{orig: args, matcher: matchers, normal: normal}, nil
}

// True returns true if this should be evaluated normally ("true is true")
//  and false if the result should be inverted ("false is true")
//
func (g *GlobMatch) True() bool { return g.normal }

// MarshalJSON is really a debug function
func (g *GlobMatch) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s: %v %v\"", "GlobMatch", g.normal, g.orig)), nil
}

// Match satifies the Matcher interface
func (g *GlobMatch) Match(file string) bool {
	for _, gg := range g.matcher {
		if gg.Match(file) {
			return true
		}
	}
	return false
}
