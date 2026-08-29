package util

import "strings"

// IsDevelopment is local/Tilt. Empty ENVIRONMENT counts as development.
func IsDevelopment(env string) bool {
	e := strings.ToLower(strings.TrimSpace(env))
	return e == "" || e == "development"
}
