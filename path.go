package yar

import (
	"bytes"
	"fmt"
	"strings"
)

type Param struct {
	Key   string
	Value string
}

type Params []Param

func (ps Params) Value(key string) string {
	for _, p := range ps {
		if p.Key == key {
			return p.Value
		}
	}
	return ""
}

type Path struct {
	UrlPattern string
	ParamKeys  []string
}

func NewPath(urlPattern string) *Path {
	return &Path{
		UrlPattern: urlPattern,
		ParamKeys:  getParamKeys(urlPattern),
	}
}

func (p *Path) Url(params ...string) string {
	if len(params) != len(p.ParamKeys) {
		panic(fmt.Sprintf("parameter number mismatch for url=%s,  params=", p.UrlPattern, len(params)))
	}
	var buffer bytes.Buffer
	pattern := p.UrlPattern
	i, j := 0, 0
	for i < len(pattern) {
		if !IsParam(pattern[i]) {
			buffer.WriteByte(pattern[i])
		} else if j < len(params) {
			buffer.WriteString(params[j])
			nextSlash := strings.Index(pattern[i:], "/")
			if nextSlash < 0 {
				nextSlash = len(pattern) - i
			}
			i = i + nextSlash - 1
			j++
		}
		i++
	}
	if i != len(pattern) || j != len(params) { // This should never happen
		panic(fmt.Sprintf("parameter number mismatch for url=%s, %d path params, %d provided params", p.UrlPattern, len(p.ParamKeys), len(params)))
	}
	return buffer.String()
}

func IsParam(char byte) bool {
	return char == '*' || char == ':'
}

func getParamKeys(urlPattern string) []string {

	parts := strings.Split(urlPattern, "/")
	keys := []string{}
	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			key := part[1:]
			mustBeValidKey(key)
			keys = append(keys, key)
		}
		if strings.HasPrefix(part, "*") {
			key := part[1:]
			mustBeValidKey(key)
			keys = append(keys, key)
			if i != len(parts)-1 {
				panic("wilcard parameter must last in the path")
			}
		}
	}
	return keys
}

func mustBeValidKey(key string) {
	if len(key) == 0 {
		panic("Parameters must have names")
	}
	if strings.Count(key, ":") > 0 || strings.Count(key, "*") > 0 {
		panic(fmt.Sprintf("parameter key cannot contain ':' or '*', param=%s", key))
	}
}
