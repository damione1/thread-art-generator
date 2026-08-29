package threadGenerator

import (
	"testing"
)

func TestAlgorithmsRegistry(t *testing.T) {
	t.Parallel()
	algos := Algorithms()
	if len(algos) < 2 {
		t.Fatalf("expected Vrellis and L2, got %d", len(algos))
	}
	if algos[0].ID() != KindVrellis || algos[0].FormValue() != "VRELLIS" {
		t.Fatalf("first algorithm should be Vrellis, got %+v", algos[0])
	}
	if algos[1].ID() != KindL2 || algos[1].FormValue() != "L2" {
		t.Fatalf("second algorithm should be L2, got %+v", algos[1])
	}
	if !algos[0].UsesBrightness() {
		t.Fatal("Vrellis uses brightness")
	}
	if algos[1].UsesBrightness() {
		t.Fatal("L2 does not use brightness")
	}
}

func TestLookupUnknownFallsBackToVrellis(t *testing.T) {
	t.Parallel()
	if Lookup(KindUnspecified).ID() != KindVrellis {
		t.Fatal("unspecified should resolve to Vrellis")
	}
	if Lookup(Kind(99)).ID() != KindVrellis {
		t.Fatal("unknown id should resolve to Vrellis")
	}
	if LookupForm("").ID() != KindVrellis {
		t.Fatal("empty form should resolve to Vrellis")
	}
	if LookupForm("L2").ID() != KindL2 {
		t.Fatal("L2 form should stay L2")
	}
}

func TestNormalizeKind(t *testing.T) {
	t.Parallel()
	if normalizeKind(KindUnspecified) != KindVrellis {
		t.Fatal("unspecified should be Vrellis")
	}
	if normalizeKind(KindL2) != KindL2 {
		t.Fatal("L2 must stay L2")
	}
}
