package main

import (
	"bytes"
	"fmt"
)

type MapAny map[string]interface{}

func (m MapAny) Copy() MapAny {
	a := make(MapAny)
	for k, v := range m {
		a[k] = v
	}
	return a
}

func (m MapAny) Dump() string {
	buf := bytes.Buffer{}
	for k, v := range m {
		buf.WriteString(fmt.Sprintf("%s: %v\n", k, v))
	}
	return buf.String()
}
func Merge(a MapAny, b MapAny) MapAny {
	if a == nil {
		a = make(MapAny)
	}
	for k, v := range b {
		a[k] = v
	}

	return a
}

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
			tmp[i] = fmt.Sprintf("%s", val)
		}
		return tmp, nil
	}

	return nil, fmt.Errorf("Expected string or []string, got %T", value)
}
