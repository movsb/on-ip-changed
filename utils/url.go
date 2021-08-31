package utils

import "strings"

func AddHTTPPrefix(u string) string {
	if !strings.Contains(u, `://`) {
		return `http://` + u
	}
	return u
}
