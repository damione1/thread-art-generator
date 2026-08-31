package threadGenerator

import (
	"image"
	"image/color"
	"testing"
)

func TestCLAHEIncreasesLocalContrast(t *testing.T) {
	t.Parallel()
	img := image.NewGray(image.Rect(0, 0, 64, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			img.SetGray(x, y, color.Gray{Y: uint8(x * 255 / 63)})
		}
	}
	out := applyCLAHE(img, 8, 2)
	_, _, left := grayAt(out, 4, 32)
	_, _, right := grayAt(out, 60, 32)
	if right <= left {
		t.Fatalf("gradient should stay left-to-right, left=%d right=%d", left, right)
	}
	if right-left < 80 {
		t.Fatalf("CLAHE should keep a wide range, left=%d right=%d", left, right)
	}
}

func TestMixEdgesDarkensBoundary(t *testing.T) {
	t.Parallel()
	img := solidGray(48, 48, 200)
	for y := 0; y < 48; y++ {
		for x := 0; x < 24; x++ {
			img.SetGray(x, y, color.Gray{Y: 40})
		}
	}
	out := mixEdges(img, 0.8)
	_, _, edge := grayAt(out, 24, 24)
	_, _, flat := grayAt(out, 8, 24)
	if edge >= 80 {
		t.Fatalf("edge should darken, got %d", edge)
	}
	if flat > 50 {
		t.Fatalf("flat dark region should stay dark, got %d", flat)
	}
}

func TestPrepareGraySquareMasksOutsideWhite(t *testing.T) {
	t.Parallel()
	tg := NewThreadGenerator(smallTestConfig())
	tg.imageContrast = 28
	out := tg.prepareGraySquare(solidGray(40, 40, 10), 32)
	if out.Bounds().Dx() != 32 {
		t.Fatalf("size %v", out.Bounds())
	}
	if out.NRGBAAt(0, 0).R < 250 {
		t.Fatalf("outside circle should be white, got %v", out.NRGBAAt(0, 0))
	}
}

func TestClaheClipMapping(t *testing.T) {
	t.Parallel()
	if got := claheClip(0); got != 0 {
		t.Fatalf("zero contrast should skip, got %v", got)
	}
	if got := claheClip(28); got < 1.5 || got > 1.6 {
		t.Fatalf("28%% should map near 1.56, got %v", got)
	}
}

func grayAt(img image.Image, x, y int) (int, int, int) {
	r, g, b, _ := img.At(x, y).RGBA()
	return int(r >> 8), int(g >> 8), int(b >> 8)
}
