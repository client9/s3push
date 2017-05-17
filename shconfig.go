package main

import (
	"fmt"
	"strings"

	"github.com/google/shlex"
)

func RequireString0(args []string, fn func() error) error {
	if len(args) != 1 {
		return fmt.Errorf("%s: expect 1 arg got %d", args[0], len(args))
	}
	return fn()
}

func RequireString1(args []string, fn func(string) error) error {
	if len(args) != 2 {
		return fmt.Errorf("%s: expect 1 arg got %d", args[0], len(args))
	}
	return fn(args[1])
}

type Dispatcher interface {
	ConfCall(args []string) error
	ConfObject(args []string) (Dispatcher, error)
}

func Parse(root Dispatcher, conf string) error {
	stack := []Dispatcher{root}
	lines := strings.Split(conf, "\n")
	for num, s := range lines {
		parts, err := shlex.Split(s)
		if err != nil {
			return fmt.Errorf("Line %d: Unable to parse line: %s", num+1, err)
		}
		if len(parts) == 0 {
			continue
		}
		if parts[len(parts)-1] == "{" {
			parts = parts[:len(parts)-1]
			d, err := stack[len(stack)-1].ConfObject(parts)
			if err != nil {
				return fmt.Errorf("Line %d: config err %s", num+1, err)
			}
			stack = append(stack, d)
			continue
		}
		if len(parts) == 1 && parts[0] == "}" {
			if len(stack) == 0 {
				return fmt.Errorf("Line %d: fell off edge", num+1)
			}
			stack = stack[:len(stack)-1]
			continue
		}
		err = stack[len(stack)-1].ConfCall(parts)
		if err != nil {
			return fmt.Errorf("Line %d: config err %s", num+1, err)
		}
	}

	return nil
}
