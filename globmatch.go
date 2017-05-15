package main

import (
	"fmt"

	"github.com/gobwas/glob"
)

type GlobMatch struct {
	orig    []string
	matcher []glob.Glob
	normal  bool
}

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
func (g *GlobMatch) True() bool { return g.normal }

func (g *GlobMatch) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s: %v %v\"", "GlobMatch", g.normal, g.orig)), nil
}

func (g *GlobMatch) Match(file string) bool {
	for _, gg := range g.matcher {
		if gg.Match(file) {
			return true
		}
	}
	return false
}
