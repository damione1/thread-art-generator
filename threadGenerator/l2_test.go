package threadGenerator

import (
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

func TestDefaultConfigAlgorithm(t *testing.T) {
	t.Parallel()
	if DefaultConfig().Algorithm != KindVrellis {
		t.Fatal("default algorithm should be Vrellis")
	}
}
