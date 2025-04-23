package util

import (
	"regexp"
	"strings"
	"unicode"
)

func Slugify(s string) string {
	var re = regexp.MustCompile(`[^a-zA-Z0-9-_]+`)
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) {
			return r
		}
		return -1
	}, s)
	s = re.ReplaceAllString(s, "")
	s = strings.ToLower(s)

	return s
}
