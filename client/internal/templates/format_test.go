package templates

import (
	"testing"

	"github.com/Damione1/thread-art-generator/core/pb"
)

func TestHumanInt(t *testing.T) {
	for _, c := range []struct {
		in   int32
		want string
	}{{0, "0"}, {7, "7"}, {450, "450"}, {3148, "3 148"}, {12000, "12 000"}, {2412000, "2 412 000"}, {-1500, "-1 500"}} {
		if got := humanInt(c.in); got != c.want {
			t.Errorf("humanInt(%d) = %q, want %q", c.in, got, c.want)
		}
	}
	if got := threadMetres(2_412_000); got != "2.4 km" {
		t.Errorf("threadMetres km = %q", got)
	}
	if got := threadMetres(1_225_000); got != "1.2 km" {
		t.Errorf("threadMetres 1225m = %q", got)
	}
	if got := threadMetres(2412); got != "2 m" {
		t.Errorf("threadMetres 2.4m = %q", got)
	}
	if got := spoolFraction(1_225_000); got != "24% of a 5 km spool" {
		t.Errorf("spoolFraction = %q", got)
	}
}

func TestCompositionParamsLineContrastPercent(t *testing.T) {
	got := compositionParamsLine(&pb.Composition{
		NailsQuantity:     280,
		MaxPaths:          4500,
		ImgSize:           800,
		MinimumDifference: 22,
		ImageContrast:     28,
	})
	want := "Vrellis · 280 nails · 4 500 paths · 800 px · gap 22 · contrast 28%"
	if got != want {
		t.Errorf("compositionParamsLine = %q, want %q", got, want)
	}

	got = compositionParamsLine(&pb.Composition{
		Algorithm:         pb.CompositionAlgorithm_COMPOSITION_ALGORITHM_L2,
		NailsQuantity:     280,
		MaxPaths:          4500,
		ImgSize:           800,
		MinimumDifference: 22,
		ImageContrast:     28,
	})
	want = "L2 residual · 280 nails · 4 500 paths · 800 px · gap 22 · contrast 28%"
	if got != want {
		t.Errorf("compositionParamsLine L2 = %q, want %q", got, want)
	}
}

func TestSolverModelHelpers(t *testing.T) {
	if algorithmValue(pb.CompositionAlgorithm_COMPOSITION_ALGORITHM_UNSPECIFIED) != "VRELLIS" {
		t.Fatal("unspecified form value")
	}
	if algorithmLabel(pb.CompositionAlgorithm_COMPOSITION_ALGORITHM_L2) != "L2 residual" {
		t.Fatal("L2 label")
	}
	if solverUsesBrightness(pb.CompositionAlgorithm_COMPOSITION_ALGORITHM_L2) {
		t.Fatal("L2 should hide thread weight")
	}
	if !solverUsesBrightness(pb.CompositionAlgorithm_COMPOSITION_ALGORITHM_VRELLIS) {
		t.Fatal("Vrellis should show thread weight")
	}
}
