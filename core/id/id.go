package id

import "github.com/google/uuid"

// Generator creates opaque public IDs (UUIDs in resource names).
type Generator interface {
	New() string
}

// UUID is a v4 generator.
type UUID struct{}

func (UUID) New() string { return uuid.NewString() }
