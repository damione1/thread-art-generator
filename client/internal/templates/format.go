package templates

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

const polyesterSpoolMM = 5_000_000 // ordinary 5 km polyester spool

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

// threadMetres renders thread length stored in millimetres.
func threadMetres(mm int32) string {
	if mm <= 0 {
		return "0 m"
	}
	metres := float64(mm) / 1000
	if metres >= 1000 {
		s := strconv.FormatFloat(metres/1000, 'f', 1, 64)
		s = strings.TrimSuffix(s, ".0")
		return s + " km"
	}
	return humanInt(int32(math.Round(metres))) + " m"
}

func spoolFraction(mm int32) string {
	if mm <= 0 {
		return ""
	}
	pct := 100 * float64(mm) / polyesterSpoolMM
	return fmt.Sprintf("%.0f%% of a 5 km spool", pct)
}
