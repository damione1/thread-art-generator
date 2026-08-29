package auth

import (
	"sync"
	"time"
)

// AuthLimiter is a sliding-window limiter for login/signup/reset.
type AuthLimiter struct {
	mu      sync.Mutex
	hits    map[string][]time.Time
	window  time.Duration
	max     int
	now     func() time.Time
}

func NewAuthLimiter(window time.Duration, max int) *AuthLimiter {
	if window <= 0 {
		window = time.Minute
	}
	if max <= 0 {
		max = 10
	}
	return &AuthLimiter{
		hits:   make(map[string][]time.Time),
		window: window,
		max:    max,
		now:    time.Now,
	}
}

func DefaultAuthLimiter() *AuthLimiter {
	return NewAuthLimiter(time.Minute, 10)
}

func (l *AuthLimiter) Allow(key string) bool {
	if l == nil {
		return true
	}
	if key == "" {
		key = "unknown"
	}
	now := l.now()
	cutoff := now.Add(-l.window)

	l.mu.Lock()
	defer l.mu.Unlock()
	times := l.hits[key]
	kept := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= l.max {
		l.hits[key] = kept
		return false
	}
	l.hits[key] = append(kept, now)
	return true
}
