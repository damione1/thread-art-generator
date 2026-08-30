package threadGenerator

import (
	"fmt"
	"image"
	"sort"
)

// Kind is the persisted / proto algorithm id.
type Kind int

const (
	// KindUnspecified is treated as Vrellis.
	KindUnspecified Kind = 0
	// KindVrellis is consecutive greedy on remaining darkness (Bresenham).
	KindVrellis Kind = 1
	// KindL2 is global L2 residual greedy (Birsak / bdring StringArt), then Euler.
	KindL2 Kind = 2
)

// Algorithm selects how paths are chosen. G-code emission is shared.
type Algorithm interface {
	ID() Kind
	FormValue() string
	Label() string
	Hint() string
	UsesBrightness() bool
	Solve(tg *ThreadGenerator) error
	RenderPreview(tg *ThreadGenerator) (image.Image, error)
}

var (
	registry []Algorithm
	byID     = map[Kind]Algorithm{}
	byForm   = map[string]Algorithm{}
)

// Register adds an algorithm. Call from init() in the algorithm's file.
func Register(a Algorithm) {
	if a == nil {
		panic("threadGenerator: nil algorithm")
	}
	if _, ok := byID[a.ID()]; ok {
		panic(fmt.Sprintf("threadGenerator: duplicate algorithm id %d", a.ID()))
	}
	if _, ok := byForm[a.FormValue()]; ok {
		panic(fmt.Sprintf("threadGenerator: duplicate algorithm form %q", a.FormValue()))
	}
	registry = append(registry, a)
	byID[a.ID()] = a
	byForm[a.FormValue()] = a
}

// Algorithms returns registered algorithms in ID order (Vrellis, then L2, …).
func Algorithms() []Algorithm {
	out := append([]Algorithm(nil), registry...)
	sort.Slice(out, func(i, j int) bool { return out[i].ID() < out[j].ID() })
	return out
}

// Lookup resolves a persisted / proto algorithm id. Unknown → Vrellis.
func Lookup(k Kind) Algorithm {
	if a, ok := byID[k]; ok {
		return a
	}
	if a, ok := byID[KindVrellis]; ok {
		return a
	}
	if len(registry) == 0 {
		panic("threadGenerator: no algorithms registered")
	}
	return registry[0]
}

// LookupForm resolves a composition-form radio value. Unknown → Vrellis.
func LookupForm(value string) Algorithm {
	if a, ok := byForm[value]; ok {
		return a
	}
	return Lookup(KindVrellis)
}

func normalizeKind(k Kind) Kind {
	return Lookup(k).ID()
}
