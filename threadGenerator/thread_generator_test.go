package threadGenerator

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	if cfg.NailsQuantity != 280 || cfg.ImgSize != 800 || cfg.MaxPaths != 4500 {
		t.Fatalf("unexpected size defaults: %+v", cfg)
	}
	if cfg.MinimumDifference != 22 || cfg.BrightnessFactor != 40 || cfg.ImageContrast != 28 {
		t.Fatalf("unexpected look defaults: %+v", cfg)
	}
	if cfg.PhysicalRadius != 304.8 || cfg.StopWeightThreshold != 10 || cfg.NailCooldown != 3 {
		t.Fatalf("unexpected physics/solver defaults: %+v", cfg)
	}
}

func TestValidateParams(t *testing.T) {
	t.Parallel()
	if errs := ValidateParams(280, 0, 22); len(errs) != 0 {
		t.Fatalf("valid params: %v", errs)
	}
	if errs := ValidateParams(2, 0, 1); len(errs) == 0 {
		t.Fatal("expected nails_quantity error")
	}
	errs := ValidateParams(10, 10, 5)
	if len(errs) != 2 {
		t.Fatalf("expected starting_nail + minDiff, got %v", errs)
	}
}

func TestPairKeySymmetric(t *testing.T) {
	t.Parallel()
	if pairKey(3, 9) != pairKey(9, 3) {
		t.Fatal("pairKey must be order-independent")
	}
	if pairKey(3, 9) == pairKey(3, 8) {
		t.Fatal("distinct pairs must differ")
	}
}

func TestCircularDiff(t *testing.T) {
	t.Parallel()
	if got := circularDiff(0, 1, 10); got != 1 {
		t.Fatalf("got %d", got)
	}
	if got := circularDiff(0, 9, 10); got != 1 {
		t.Fatalf("wrap got %d", got)
	}
	if got := circularDiff(0, 5, 10); got != 5 {
		t.Fatalf("opposite got %d", got)
	}
}

func TestPixelSizeUsesDiameter(t *testing.T) {
	t.Parallel()
	got := pixelSizeMM(304.8, 800)
	want := 2 * 304.8 / 800
	if got != want {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestLineLengthEuclidean(t *testing.T) {
	t.Parallel()
	tg := NewThreadGenerator(Config{PhysicalRadius: 10, ImgSize: 10})
	tg.nailsList = []Nail{{X: 0, Y: 0}, {X: 3, Y: 4}}
	got := tg.lineLength(0, 1)
	want := 5.0 * pixelSizeMM(10, 10)
	if got != want {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestNailsStayInBounds(t *testing.T) {
	t.Parallel()
	tg := NewThreadGenerator(Config{NailsQuantity: 280, ImgSize: 80})
	src := image.NewNRGBA(image.Rect(0, 0, 80, 80))
	nails := tg.getNailsListFromImage(src)
	for i, n := range nails {
		if n.X < 0 || n.Y < 0 || n.X >= 80 || n.Y >= 80 {
			t.Fatalf("nail %d out of bounds: %v", i, n)
		}
	}
}

func TestGenerateAcceptsNonSquareSource(t *testing.T) {
	t.Parallel()
	cfg := smallTestConfig()
	for _, img := range []image.Image{solidGray(120, 60, 40), solidGray(60, 120, 40)} {
		path := writePNG(t, img)
		tg := NewThreadGenerator(cfg)
		if _, err := tg.Generate(Args{ImageName: path}); err != nil {
			t.Fatal(err)
		}
	}
}

func TestGenerateRejectsBadStartingNail(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.StartingNail = cfg.NailsQuantity
	tg := NewThreadGenerator(cfg)
	_, err := tg.Generate(Args{ImageName: "missing.png"})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestGenerateWhiteImageStopsImmediately(t *testing.T) {
	t.Parallel()
	path := writePNG(t, solidGray(64, 64, 255))
	cfg := smallTestConfig()
	tg := NewThreadGenerator(cfg)
	stats, err := tg.Generate(Args{ImageName: path})
	if err != nil {
		t.Fatal(err)
	}
	if stats.TotalLines != 0 {
		t.Fatalf("white image should emit no lines, got %d", stats.TotalLines)
	}
}

func TestGenerateDarkImageEmitsPathsAndHonorsCooldown(t *testing.T) {
	t.Parallel()
	path := writePNG(t, darkCenter(96, 96))
	cfg := smallTestConfig()
	cfg.MaxPaths = 40
	cfg.NailCooldown = 3
	tg := NewThreadGenerator(cfg)
	stats, err := tg.Generate(Args{ImageName: path})
	if err != nil {
		t.Fatal(err)
	}
	if stats.TotalLines < 5 {
		t.Fatalf("expected several lines, got %d", stats.TotalLines)
	}

	nodes := []int{cfg.StartingNail}
	for _, p := range tg.GetPathsList() {
		if p.StartingNail != nodes[len(nodes)-1] {
			t.Fatalf("path not continuous: %+v after %v", p, nodes)
		}
		nodes = append(nodes, p.EndingNail)
	}
	cooldown := cfg.NailCooldown
	for k := 0; k < len(nodes)-1; k++ {
		next := nodes[k+1]
		start := k + 1 - cooldown
		if start < 0 {
			start = 0
		}
		for _, prev := range nodes[start : k+1] {
			if next == prev {
				t.Fatalf("nail %d reused inside cooldown window at step %d: %v", next, k, nodes)
			}
		}
	}

	preview, err := tg.GeneratePathsImage()
	if err != nil {
		t.Fatal(err)
	}
	if minGray(preview) > 40 {
		t.Fatalf("preview should reach near-black, min=%d", minGray(preview))
	}
	if stats.ThreadLength < 0 {
		t.Fatal("thread length should be non-negative")
	}
}

func TestThreadLengthIsMetres(t *testing.T) {
	t.Parallel()
	path := writePNG(t, darkCenter(96, 96))
	cfg := smallTestConfig()
	cfg.PhysicalRadius = 609.6
	cfg.ImgSize = 80
	cfg.MaxPaths = 40
	tg := NewThreadGenerator(cfg)
	stats, err := tg.Generate(Args{ImageName: path})
	if err != nil {
		t.Fatal(err)
	}
	if stats.TotalLines == 0 {
		t.Fatal("expected lines")
	}
	// 40 chords on a ~1.22 m hoop is tens of metres, never 0–2.
	if stats.ThreadLength < 8000 {
		t.Fatalf("thread length too small: %d mm for %d lines", stats.ThreadLength, stats.TotalLines)
	}
}

func TestGenerateDictionarySkipsShortChords(t *testing.T) {
	t.Parallel()
	tg := NewThreadGenerator(Config{NailsQuantity: 20, MinimumDifference: 4, ImgSize: 40})
	src := image.NewNRGBA(image.Rect(0, 0, 40, 40))
	nails := tg.getNailsListFromImage(src)
	dict := tg.generateDictionary(nails)
	if _, ok := dict[pairKey(0, 1)]; ok {
		t.Fatal("adjacent nails should not be in the dictionary")
	}
	if _, ok := dict[pairKey(0, 4)]; !ok {
		t.Fatal("legal chord missing from dictionary")
	}
}

func smallTestConfig() Config {
	return Config{
		NailsQuantity:       24,
		ImgSize:             64,
		MaxPaths:            80,
		StartingNail:        0,
		MinimumDifference:   4,
		BrightnessFactor:    40,
		ImageContrast:       1,
		PhysicalRadius:      100,
		StopWeightThreshold: 10,
		NailCooldown:        3,
		RotationAxis:        "A",
		NeedleAxis:          "X",
		SpindleAxis:         "Y",
	}
}

func solidGray(w, h int, y uint8) *image.Gray {
	img := image.NewGray(image.Rect(0, 0, w, h))
	for yy := 0; yy < h; yy++ {
		for xx := 0; xx < w; xx++ {
			img.SetGray(xx, yy, color.Gray{Y: y})
		}
	}
	return img
}

func darkCenter(w, h int) *image.Gray {
	img := solidGray(w, h, 255)
	cx, cy := w/2, h/2
	r2 := (w / 4) * (w / 4)
	for yy := 0; yy < h; yy++ {
		for xx := 0; xx < w; xx++ {
			dx, dy := xx-cx, yy-cy
			if dx*dx+dy*dy <= r2 {
				img.SetGray(xx, yy, color.Gray{Y: 10})
			}
		}
	}
	return img
}

func writePNG(t *testing.T, img image.Image) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "src.png")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
	return path
}

func minGray(img image.Image) uint8 {
	b := img.Bounds()
	minV := uint8(255)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, _, _, _ := img.At(x, y).RGBA()
			yv := uint8(r >> 8)
			if yv < minV {
				minV = yv
			}
		}
	}
	return minV
}
