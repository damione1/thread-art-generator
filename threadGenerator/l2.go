package threadGenerator

import (
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
	"sort"

	_ "golang.org/x/image/webp"
)

func init() {
	Register(l2{})
}

type l2 struct{}

func (l2) ID() Kind             { return KindL2 }
func (l2) FormValue() string    { return "L2" }
func (l2) Label() string        { return "L2 residual" }
func (l2) Hint() string         { return "Birsak / StringArt: global L2, Xiaolin Wu, then an Euler tour." }
func (l2) UsesBrightness() bool { return false }

func (l2) Solve(tg *ThreadGenerator) error {
	return tg.generateL2()
}

func (l2) RenderPreview(tg *ThreadGenerator) (image.Image, error) {
	return tg.l2PreviewImage(), nil
}

const l2BadRunsLimit = 1000

type l2Edge struct {
	a, b    int
	samples []wuSample
	delta   float64
	used    bool
}

type pixHit struct {
	edge int
	c    float64
}

func (tg *ThreadGenerator) generateL2() error {
	file, err := os.Open(tg.imageName)
	if err != nil {
		return err
	}
	defer file.Close()
	src, _, err := image.Decode(file)
	if err != nil {
		return err
	}

	w := tg.imgSize
	if w < 8 {
		w = 8
	}
	target, nails := tg.l2TargetAndNails(src, w)
	tg.nailsList = nails

	scale := tg.threadHalfWidthPx() * 2
	if scale <= 0 {
		scale = 1
	}
	if scale > 1 {
		scale = 1
	}

	n := len(nails)
	edges := make([]l2Edge, 0, n*(n-1)/2)
	hits := make([][]pixHit, w*w)

	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			if circularDiff(i, j, n) < tg.minimumDifference {
				continue
			}
			samples := xiaolinWu(
				float64(nails[i].X)+0.5, float64(nails[i].Y)+0.5,
				float64(nails[j].X)+0.5, float64(nails[j].Y)+0.5,
				w, w, scale,
			)
			if len(samples) == 0 {
				continue
			}
			ei := len(edges)
			edges = append(edges, l2Edge{a: i, b: j, samples: samples})
			for _, s := range samples {
				hits[s.i] = append(hits[s.i], pixHit{edge: ei, c: s.c})
			}
		}
	}
	if len(edges) == 0 {
		return nil
	}

	recon := make([]float64, w*w)
	for ei := range edges {
		edges[ei].delta = l2Delta(target, recon, edges[ei].samples)
	}

	l2 := 0.0
	for _, t := range target {
		l2 += t * t
	}

	picked := make([]int, 0, tg.maxPaths)
	bestPicked := 0
	bestL2 := l2
	badRuns := 0

	for iter := 0; iter < tg.maxPaths; iter++ {
		best := -1
		bestDelta := math.Inf(1)
		for i := range edges {
			if edges[i].used {
				continue
			}
			if edges[i].delta < bestDelta {
				bestDelta = edges[i].delta
				best = i
			}
		}
		if best < 0 {
			break
		}

		applyL2Edge(target, recon, hits, edges, best)
		edges[best].used = true
		l2 += bestDelta
		picked = append(picked, best)

		if l2+1e-12 < bestL2 {
			bestL2 = l2
			bestPicked = len(picked)
			badRuns = 0
		} else {
			badRuns++
			if badRuns >= l2BadRunsLimit {
				break
			}
		}
	}
	if bestPicked < len(picked) {
		picked = picked[:bestPicked]
	}

	pairs := make([][2]int, len(picked))
	for i, ei := range picked {
		pairs[i] = [2]int{edges[ei].a, edges[ei].b}
	}
	tg.pathsList = eulerTour(n, tg.startingNail, pairs)
	tg.threadLength = 0
	for _, p := range tg.pathsList {
		tg.threadLength += tg.lineLength(p.StartingNail, p.EndingNail)
	}
	return nil
}

func (tg *ThreadGenerator) l2TargetAndNails(src image.Image, w int) ([]float64, []Nail) {
	nrgba := tg.prepareGraySquare(src, w)

	pix := make([]float64, w*w)
	for y := 0; y < w; y++ {
		for x := 0; x < w; x++ {
			pix[y*w+x] = float64(nrgba.NRGBAAt(x, y).R) / 255
		}
	}
	// Stretch only the disc. Min-max is wrecked by a single black corner or
	// grey backdrop; 2–98 percentiles ignore those outliers.
	percentileStretchInside(pix, w, 0.02, 0.98)
	// invertInput: dark photograph → high target (thread wanted).
	// Outside the circle must stay 0 so chords are not pulled to the rim.
	for i, v := range pix {
		if !insideCircle(i, w) {
			pix[i] = 0
			continue
		}
		pix[i] = 1 - v
	}

	nails := make([]Nail, tg.nailsQuantity)
	center := w / 2
	radius := float64(center) - 1
	if radius < 1 {
		radius = 1
	}
	for i := 0; i < tg.nailsQuantity; i++ {
		alpha := float64(i) * 2 * math.Pi / float64(tg.nailsQuantity)
		nails[i] = Nail{
			X: center + int(radius*math.Cos(alpha)),
			Y: center + int(radius*math.Sin(alpha)),
		}
	}
	return pix, nails
}

func insideCircle(i, w int) bool {
	cx, cy := float64(w)/2-0.5, float64(w)/2-0.5
	r2 := float64(w) * float64(w) / 4
	x := float64(i%w) - cx
	y := float64(i/w) - cy
	return x*x+y*y <= r2
}

// percentileStretchInside maps the disc to [0,1] using loP/hiP percentiles so
// a dark corner or pale backdrop cannot set the black/white points.
func percentileStretchInside(pix []float64, w int, loP, hiP float64) {
	inside := make([]float64, 0, len(pix))
	for i, v := range pix {
		if insideCircle(i, w) {
			inside = append(inside, v)
		}
	}
	if len(inside) == 0 {
		return
	}
	lo, hi := percentile(inside, loP), percentile(inside, hiP)
	rangeV := hi - lo
	if rangeV < 1e-9 {
		return
	}
	for i, v := range pix {
		if !insideCircle(i, w) {
			continue
		}
		t := (v - lo) / rangeV
		if t < 0 {
			t = 0
		}
		if t > 1 {
			t = 1
		}
		pix[i] = t
	}
}

func percentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}
	cp := append([]float64(nil), values...)
	sort.Float64s(cp)
	if p <= 0 {
		return cp[0]
	}
	if p >= 1 {
		return cp[len(cp)-1]
	}
	idx := p * float64(len(cp)-1)
	lo := int(idx)
	hi := lo + 1
	if hi >= len(cp) {
		return cp[lo]
	}
	frac := idx - float64(lo)
	return cp[lo]*(1-frac) + cp[hi]*frac
}

func l2Delta(target, recon []float64, samples []wuSample) float64 {
	d := 0.0
	for _, s := range samples {
		t := target[s.i]
		oldR := recon[s.i]
		newR := oldR + s.c
		if newR > 1 {
			newR = 1
		}
		d += (t-newR)*(t-newR) - (t-oldR)*(t-oldR)
	}
	return d
}

func applyL2Edge(target, recon []float64, hits [][]pixHit, edges []l2Edge, ei int) {
	for _, s := range edges[ei].samples {
		oldR := recon[s.i]
		newR := oldR + s.c
		if newR > 1 {
			newR = 1
		}
		if newR == oldR {
			continue
		}
		t := target[s.i]
		recon[s.i] = newR
		for _, hit := range hits[s.i] {
			if hit.edge == ei || edges[hit.edge].used {
				continue
			}
			oldC := sq(t-clamp01(oldR+hit.c)) - sq(t-oldR)
			newC := sq(t-clamp01(newR+hit.c)) - sq(t-newR)
			edges[hit.edge].delta += newC - oldC
		}
	}
}

func sq(v float64) float64 { return v * v }

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func (tg *ThreadGenerator) l2PreviewImage() image.Image {
	w := tg.imgSize
	if w < 1 {
		w = 1
	}
	recon := make([]float64, w*w)
	scale := tg.threadHalfWidthPx() * 2
	if scale <= 0 {
		scale = 1
	}
	if scale > 1 {
		scale = 1
	}
	for _, p := range tg.pathsList {
		a := tg.nailsList[p.StartingNail]
		b := tg.nailsList[p.EndingNail]
		for _, s := range xiaolinWu(float64(a.X)+0.5, float64(a.Y)+0.5, float64(b.X)+0.5, float64(b.Y)+0.5, w, w, scale) {
			v := recon[s.i] + s.c
			if v > 1 {
				v = 1
			}
			recon[s.i] = v
		}
	}
	out := image.NewGray(image.Rect(0, 0, w, w))
	if tg.lightOnDark() {
		for i, v := range recon {
			out.Pix[i] = uint8(math.Round(255 * v))
		}
		return out
	}
	for i, v := range recon {
		out.Pix[i] = uint8(math.Round(255 * (1 - v)))
	}
	return out
}
