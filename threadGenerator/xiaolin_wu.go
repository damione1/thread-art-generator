package threadGenerator

import "math"

type wuSample struct {
	i int
	c float64
}

func ipart(x float64) float64 {
	if x >= 0 {
		return math.Floor(x)
	}
	return math.Ceil(x)
}

func fpart(x float64) float64 {
	return x - ipart(x)
}

func rfpart(x float64) float64 {
	return 1 - fpart(x)
}

func appendWu(out []wuSample, x, y, c float64, w, h int, scale float64) []wuSample {
	if c <= 0 {
		return out
	}
	xi, yi := int(x), int(y)
	if xi < 0 || yi < 0 || xi >= w || yi >= h {
		return out
	}
	cov := c * scale
	if cov > 1 {
		cov = 1
	}
	return append(out, wuSample{i: yi*w + xi, c: cov})
}

// xiaolinWu is the MATLAB XiaolinWu.m rasterizer, clipped to [0,w)×[0,h).
// scale multiplies coverage (thread thinner than one pixel).
func xiaolinWu(x1, y1, x2, y2 float64, w, h int, scale float64) []wuSample {
	if scale <= 0 {
		scale = 1
	}
	dx := x2 - x1
	dy := y2 - y1
	cap := int(2*math.Sqrt(dx*dx+dy*dy)) + 8
	out := make([]wuSample, 0, cap)

	swapped := false
	if math.Abs(dx) < math.Abs(dy) {
		x1, y1 = y1, x1
		x2, y2 = y2, x2
		dx, dy = dy, dx
		swapped = true
	}
	if x2 < x1 {
		x1, x2 = x2, x1
		y1, y2 = y2, y1
	}
	if dx == 0 {
		return appendWu(out, x1, y1, 1, w, h, scale)
	}
	gradient := dy / dx

	plot := func(px, py, c float64) {
		if swapped {
			px, py = py, px
		}
		out = appendWu(out, px, py, c, w, h, scale)
	}

	xend := math.Round(x1)
	yend := y1 + gradient*(xend-x1)
	xgap := rfpart(x1 + 0.5)
	xpxl1 := xend
	ypxl1 := ipart(yend)
	plot(xpxl1, ypxl1, rfpart(yend)*xgap)
	plot(xpxl1, ypxl1+1, fpart(yend)*xgap)
	intery := yend + gradient

	xend = math.Round(x2)
	yend = y2 + gradient*(xend-x2)
	xgap = fpart(x2 + 0.5)
	xpxl2 := xend
	ypxl2 := ipart(yend)
	plot(xpxl2, ypxl2, rfpart(yend)*xgap)
	plot(xpxl2, ypxl2+1, fpart(yend)*xgap)

	for x := xpxl1 + 1; x <= xpxl2-1; x++ {
		y := ipart(intery)
		plot(x, y, rfpart(intery))
		plot(x, y+1, fpart(intery))
		intery += gradient
	}
	return mergeWu(out)
}

func mergeWu(in []wuSample) []wuSample {
	if len(in) < 2 {
		return in
	}
	acc := make(map[int]float64, len(in))
	order := make([]int, 0, len(in))
	for _, s := range in {
		if _, ok := acc[s.i]; !ok {
			order = append(order, s.i)
		}
		c := acc[s.i] + s.c
		if c > 1 {
			c = 1
		}
		acc[s.i] = c
	}
	out := make([]wuSample, len(order))
	for i, pix := range order {
		out[i] = wuSample{i: pix, c: acc[pix]}
	}
	return out
}
