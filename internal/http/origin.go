package http

import (
	"regexp"
	"strings"
)

func IsOriginAllowed(origin string, patterns []string) bool {
	for _, pattern := range patterns {
		switch {
		case len(pattern) > 2 && pattern[0] == '/' && pattern[len(pattern)-1] == '/':
			// Treat as raw regex (/^https:\/\/foo\.com$/)
			if matched, _ := regexp.MatchString(pattern[1:len(pattern)-1], origin); matched {
				return true
			}

		case strings.Contains(pattern, "*"):
			// Convert glob-style wildcard (*) to regex
			re := "^" + regexp.QuoteMeta(pattern)
			re = strings.ReplaceAll(re, `\*`, ".*") + "$"
			if matched, _ := regexp.MatchString(re, origin); matched {
				return true
			}

		default:
			// Exact match
			if origin == pattern {
				return true
			}
		}
	}
	return false
}
