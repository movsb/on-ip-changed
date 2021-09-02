package website

import (
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"strings"
)

type Extractor interface {
	Extract() (string, error)
}

type JsonExtractor struct {
	s    string
	path string
}

func NewJsonExtractor(s string, path string) Extractor {
	return &JsonExtractor{s: s, path: path}
}

func (e *JsonExtractor) Extract() (string, error) {
	var i interface{}
	if err := json.Unmarshal([]byte(e.s), &i); err != nil {
		return ``, fmt.Errorf(`json_extractor: error decoding: %w`, err)
	}
	for _, path := range strings.Split(e.path, `.`) {
		m, ok := i.(map[string]interface{})
		if !ok {
			return "", fmt.Errorf(`json_extractor: not an object`)
		}
		ii, ok := m[path]
		if !ok {
			return "", fmt.Errorf(`json_extractor: path not found`)
		}
		i = ii
	}
	if si, ok := i.(string); ok {
		return si, nil
	}
	return ``, fmt.Errorf(`json_extractor: value not string`)
}

type RawExtractor struct {
	s string
}

func NewRawExtractor(s string) Extractor {
	return &RawExtractor{s: s}
}

func (e *RawExtractor) Extract() (string, error) {
	s := strings.TrimSpace(e.s)
	ip := net.ParseIP(s).To4()
	if ip == nil {
		return ``, fmt.Errorf(`text_extractor: not an ip: %s`, s)
	}
	return ip.String(), nil
}

type SearchExtractor struct {
	s string
}

func NewSearchExtractor(s string) Extractor {
	return &SearchExtractor{s: s}
}

func (e *SearchExtractor) Extract() (string, error) {
	re := regexp.MustCompile(`\d+\.\d+\.\d+\.\d+`)
	// TODO search upon success
	ipstr := re.FindString(e.s)
	if ipstr == `` {
		return ``, fmt.Errorf(`search_extractor: ip not found`)
	}
	ip := net.ParseIP(ipstr).To4()
	if ip == nil {
		return ``, fmt.Errorf(`search_extrator: not an ip: %s`, ipstr)
	}
	return ip.String(), nil
}
