package main

import (
	"errors"
	"strings"
)

func replaceFirstLabel(hostname string, replacement string) string {
	dot := strings.IndexByte(hostname, '.')
	if dot == -1 {
		return replacement
	} else {
		return replacement + hostname[dot:]
	}
}

func canonicalizeHostname(hostname string) (string, error) {
	if len(hostname) == 0 || hostname[0] == '.' || strings.IndexByte(hostname, '/') >= 0 {
		return "", errors.New("invalid hostname")
	}

	hostname = strings.ToLower(hostname)
	hostname = strings.TrimSuffix(hostname, ".")

	return hostname, nil
}

func wildcardHostname(hostname string) string {
	return replaceFirstLabel(hostname, "_")
}
