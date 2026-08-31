package threadGenerator

import (
	"image/color"
	"testing"
)

func TestXiaolinWuHorizontal(t *testing.T) {
	t.Parallel()
	samples := xiaolinWu(1.5, 4.5, 10.5, 4.5, 16, 16, 1)
	if len(samples) < 8 {
		t.Fatalf("too few samples: %d", len(samples))
	}
	seen := map[int]bool{}
	for _, s := range samples {
		if s.c <= 0 || s.c > 1 {
			t.Fatalf("bad coverage %v", s)
		}
		seen[s.i] = true
	}
	if !seen[4*16+4] && !seen[4*16+5] {
		t.Fatal("expected the horizontal span to hit the mid pixels")
	}
}

func TestEulerTourCoversAndIsContinuous(t *testing.T) {
	t.Parallel()
	paths := eulerTour(4, 0, [][2]int{{0, 1}, {1, 2}, {2, 3}})
	if len(paths) < 3 {
		t.Fatalf("expected at least 3 hops, got %v", paths)
	}
	if paths[0].StartingNail != 0 {
		t.Fatalf("expected start 0, got %v", paths)
	}
	for i, p := range paths {
		if i == 0 {
			continue
		}
		if p.StartingNail != paths[i-1].EndingNail {
			t.Fatalf("not continuous: %v", paths)
		}
	}
}

func TestL2GenerateEmitsContinuousPaths(t *testing.T) {
	t.Parallel()
	path := writePNG(t, darkCenter(96, 96))
	cfg := smallTestConfig()
	cfg.Algorithm = KindL2
	cfg.MaxPaths = 50
	tg := NewThreadGenerator(cfg)
	stats, err := tg.Generate(Args{ImageName: path})
	if err != nil {
		t.Fatal(err)
	}
	if stats.TotalLines < 5 {
		t.Fatalf("expected several L2 lines, got %d", stats.TotalLines)
	}
	nodes := []int{tg.GetPathsList()[0].StartingNail}
	for _, p := range tg.GetPathsList() {
		if p.StartingNail != nodes[len(nodes)-1] {
			t.Fatalf("L2 path not continuous: %+v after %v", p, nodes)
		}
		nodes = append(nodes, p.EndingNail)
	}
	preview, err := tg.GeneratePathsImage()
	if err != nil {
		t.Fatal(err)
	}
	if preview.Bounds().Dx() != cfg.ImgSize {
		t.Fatalf("preview size %v", preview.Bounds())
	}
	if minGray(preview) > 200 {
		t.Fatal("L2 preview should show dark thread")
	}
}

func TestL2WhiteImageFewOrNoLines(t *testing.T) {
	t.Parallel()
	path := writePNG(t, solidGray(64, 64, 255))
	cfg := smallTestConfig()
	cfg.Algorithm = KindL2
	cfg.MaxPaths = 20
	tg := NewThreadGenerator(cfg)
	stats, err := tg.Generate(Args{ImageName: path})
	if err != nil {
		t.Fatal(err)
	}
	if stats.TotalLines > 8 {
		t.Fatalf("white target should barely need thread, got %d", stats.TotalLines)
	}
}

func TestL2IgnoresOutsideCircle(t *testing.T) {
	t.Parallel()
	// White disc, black corners. Pre-fix L2 inverted the corners to target=1
	// and pulled chords to the rim. After the mask fix this is a blank target.
	img := solidGray(64, 64, 0)
	cx, cy := 32.0, 32.0
	r2 := 32.0 * 32.0
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			dx, dy := float64(x)+0.5-cx, float64(y)+0.5-cy
			if dx*dx+dy*dy <= r2 {
				img.SetGray(x, y, color.Gray{Y: 255})
			}
		}
	}
	path := writePNG(t, img)
	cfg := smallTestConfig()
	cfg.Algorithm = KindL2
	cfg.ImgSize = 64
	cfg.MaxPaths = 20
	cfg.ImageContrast = 0
	tg := NewThreadGenerator(cfg)
	stats, err := tg.Generate(Args{ImageName: path})
	if err != nil {
		t.Fatal(err)
	}
	if stats.TotalLines > 8 {
		t.Fatalf("black corners outside the hoop should not attract thread, got %d", stats.TotalLines)
	}
}

func TestL2OutsideCircleTargetIsZero(t *testing.T) {
	t.Parallel()
	src := solidGray(32, 32, 40)
	tg := NewThreadGenerator(smallTestConfig())
	tg.imgSize = 32
	tg.imageContrast = 0
	pix, _ := tg.l2TargetAndNails(src, 32)
	if pix[0] != 0 || pix[31] != 0 {
		t.Fatalf("corners must be target 0, got %v %v", pix[0], pix[31])
	}
	if pix[16*32+16] <= 0.5 {
		t.Fatalf("dark interior should want thread, got %v", pix[16*32+16])
	}
}

func TestPercentileStretchIgnoresOutliers(t *testing.T) {
	t.Parallel()
	w := 16
	pix := make([]float64, w*w)
	for i := range pix {
		pix[i] = 0.5
	}
	pix[w/2*w+w/2] = 0
	pix[w/2*w+w/2+1] = 1
	percentileStretchInside(pix, w, 0.02, 0.98)
	mid := pix[w/2*w+w/2+2]
	if mid < 0.4 || mid > 0.6 {
		t.Fatalf("mid grey should stay mid after 2–98 stretch, got %v", mid)
	}
}

func TestPercentile(t *testing.T) {
	t.Parallel()
	if got := percentile([]float64{0, 1, 2, 3, 4}, 0.5); got != 2 {
		t.Fatalf("median got %v", got)
	}
	if got := percentile([]float64{10}, 0.02); got != 10 {
		t.Fatalf("single value got %v", got)
	}
}

func TestDefaultConfigAlgorithm(t *testing.T) {
	t.Parallel()
	if DefaultConfig().Algorithm != KindVrellis {
		t.Fatal("default algorithm should be Vrellis")
	}
}
