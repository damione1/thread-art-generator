package clock

import "time"

// Clock is injectable time. Tests use Fake.
type Clock interface {
	Now() time.Time
}

// Real is wall-clock time.
type Real struct{}

func (Real) Now() time.Time { return time.Now() }

// Fake is a controllable clock for tests.
type Fake struct {
	T time.Time
}

func (f *Fake) Now() time.Time { return f.T }

func (f *Fake) Advance(d time.Duration) { f.T = f.T.Add(d) }
