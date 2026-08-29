package templates

import (
	"strconv"
	"strings"
)

// humanInt groups thousands with a thin space, the way the interface prints
// every number the solver measures: 3148 becomes "3 148".
func humanInt[T int | int32 | int64](n T) string {
	s := strconv.FormatInt(int64(n), 10)
	neg := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(s, "-")

	var b strings.Builder
	for i, r := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteRune(' ')
		}
		b.WriteRune(r)
	}
	if neg {
		return "-" + b.String()
	}
	return b.String()
}

// threadMetres renders a thread length, stored in millimetres, as metres.
func threadMetres(mm int32) string {
	return humanInt(int32(float64(mm)/1000+0.5)) + " m"
}
