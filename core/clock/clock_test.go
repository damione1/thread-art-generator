package clock

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFakeAdvance(t *testing.T) {
	t.Parallel()
	start := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
	f := &Fake{T: start}
	require.Equal(t, start, f.Now())
	f.Advance(10 * time.Minute)
	require.Equal(t, start.Add(10*time.Minute), f.Now())
}

func TestRealNowMoves(t *testing.T) {
	t.Parallel()
	var c Real
	a := c.Now()
	time.Sleep(time.Millisecond)
	b := c.Now()
	require.True(t, b.After(a) || b.Equal(a))
}
