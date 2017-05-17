package main

import (
	"bytes"
	"fmt"
)

// MapAny provides map of strings to anything
type MapAny map[string]interface{}

// Copy provides a shallow copy of the map
func (m MapAny) Copy() MapAny {
	a := make(MapAny)
	for k, v := range m {
		a[k] = v
	}
	return a
}

// Dump is crappy debug output
func (m MapAny) Dump() string {
	buf := bytes.Buffer{}
	for k, v := range m {
		buf.WriteString(fmt.Sprintf("%s: %v\n", k, v))
	}
	return buf.String()
}

// Merges two maps into one.
// either a or b can be nil or empty
// return the input for chaining
func Merge(a MapAny, b MapAny) MapAny {
	if a == nil {
		a = make(MapAny)
	}
	for k, v := range b {
		a[k] = v
	}

	return a
}

// helper to turn an interface into an array of strings
//
func getStrings(value interface{}) ([]string, error) {
	switch value.(type) {
	case string:
		return []string{value.(string)}, nil
	case []string:
		return value.([]string), nil
	case []interface{}:
		itmp := value.([]interface{})
		tmp := make([]string, len(itmp))
		for i, val := range itmp {
			// gross. I know reflect has a function
			// we can use instead or using Stringer
			// interface
			tmp[i] = fmt.Sprintf("%s", val)
		}
		return tmp, nil
	}

	return nil, fmt.Errorf("Expected string or []string, got %T", value)
}
