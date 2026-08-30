package threadGenerator

import (
	"image"
	"image/color"
	"math"

	"github.com/disintegration/imaging"
)

const (
	claheTiles       = 8
	unsharpSigma     = 1.0
	defaultEdgeMix   = 0.32
	claheClipFloor   = 1.0
	claheClipPerPct  = 1.0 / 50.0 // contrast 28 → clip 1.56
)

// prepareGraySquare is the shared solver input: centre-crop, resize, CLAHE,
// unsharp, edge-weighted darkening, then a white circular mask.
func (tg *ThreadGenerator) prepareGraySquare(src image.Image, w int) *image.NRGBA {
	if w < 8 {
		w = 8
	}
	gray := imaging.Grayscale(src)
	b := gray.Bounds()
	side := b.Dx()
	if dy := b.Dy(); dy < side {
		side = dy
	}
	square := imaging.CropAnchor(gray, side, side, imaging.Center)
	resized := imaging.Resize(square, w, w, imaging.Lanczos)
	clipped := applyCLAHE(resized, claheTiles, claheClip(tg.imageContrast))
	sharp := imaging.Sharpen(clipped, unsharpSigma)
	edged := mixEdges(sharp, defaultEdgeMix)
	return maskCircle(edged)
}

func claheClip(contrast float64) float64 {
	if contrast <= 0 {
		return 0
	}
	return claheClipFloor + contrast*claheClipPerPct
}

func applyCLAHE(src image.Image, tiles int, clipLimit float64) *image.NRGBA {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	out := imaging.Clone(src)
	if w < 2 || h < 2 || tiles < 2 || clipLimit <= 0 {
		return out
	}

	pix := make([]uint8, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, _, _, _ := src.At(b.Min.X+x, b.Min.Y+y).RGBA()
			pix[y*w+x] = uint8(r >> 8)
		}
	}

	tx, ty := tiles, tiles
	tileW := float64(w) / float64(tx)
	tileH := float64(h) / float64(ty)
	luts := make([][256]uint8, tx*ty)

	for tyi := 0; tyi < ty; tyi++ {
		y0 := int(math.Round(float64(tyi) * tileH))
		y1 := int(math.Round(float64(tyi+1) * tileH))
		if y1 > h {
			y1 = h
		}
		if y1 <= y0 {
			y1 = y0 + 1
		}
		for txi := 0; txi < tx; txi++ {
			x0 := int(math.Round(float64(txi) * tileW))
			x1 := int(math.Round(float64(txi+1) * tileW))
			if x1 > w {
				x1 = w
			}
			if x1 <= x0 {
				x1 = x0 + 1
			}
			luts[tyi*tx+txi] = tileLUT(pix, w, x0, y0, x1, y1, clipLimit)
		}
	}

	for y := 0; y < h; y++ {
		tyf := (float64(y)+0.5)/tileH - 0.5
		tyi := int(math.Floor(tyf))
		fy := tyf - float64(tyi)
		if tyi < 0 {
			tyi, fy = 0, 0
		}
		if tyi >= ty-1 {
			tyi, fy = ty-2, 1
		}
		for x := 0; x < w; x++ {
			txf := (float64(x)+0.5)/tileW - 0.5
			txi := int(math.Floor(txf))
			fx := txf - float64(txi)
			if txi < 0 {
				txi, fx = 0, 0
			}
			if txi >= tx-1 {
				txi, fx = tx-2, 1
			}
			v := pix[y*w+x]
			a := float64(luts[tyi*tx+txi][v])
			b1 := float64(luts[tyi*tx+txi+1][v])
			c := float64(luts[(tyi+1)*tx+txi][v])
			d := float64(luts[(tyi+1)*tx+txi+1][v])
			top := a*(1-fx) + b1*fx
			bot := c*(1-fx) + d*fx
			g := uint8(math.Round(top*(1-fy) + bot*fy))
			out.SetNRGBA(b.Min.X+x, b.Min.Y+y, color.NRGBA{R: g, G: g, B: g, A: 255})
		}
	}
	return out
}

func tileLUT(pix []uint8, w, x0, y0, x1, y1 int, clipLimit float64) [256]uint8 {
	var hist [256]int
	n := 0
	for y := y0; y < y1; y++ {
		row := y * w
		for x := x0; x < x1; x++ {
			hist[pix[row+x]]++
			n++
		}
	}
	if n == 0 {
		var ident [256]uint8
		for i := range ident {
			ident[i] = uint8(i)
		}
		return ident
	}
	clip := int(math.Round(clipLimit * float64(n) / 256))
	if clip < 1 {
		clip = 1
	}
	excess := 0
	for i := range hist {
		if hist[i] > clip {
			excess += hist[i] - clip
			hist[i] = clip
		}
	}
	redist := excess / 256
	rem := excess % 256
	for i := range hist {
		hist[i] += redist
	}
	for i := 0; i < rem; i++ {
		hist[i]++
	}
	var cdf [256]int
	sum := 0
	for i, h := range hist {
		sum += h
		cdf[i] = sum
	}
	var lut [256]uint8
	scale := 255.0 / float64(n)
	for i, c := range cdf {
		v := int(math.Round(float64(c) * scale))
		if v < 0 {
			v = 0
		}
		if v > 255 {
			v = 255
		}
		lut[i] = uint8(v)
	}
	return lut
}

// mixEdges darkens pixels that sit on Sobel edges so portraits keep outlines
// without lifting flat regions.
func mixEdges(src image.Image, mix float64) *image.NRGBA {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	out := imaging.Clone(src)
	if mix <= 0 || w < 3 || h < 3 {
		return out
	}
	if mix > 1 {
		mix = 1
	}
	lum := make([]float64, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, _, _, _ := src.At(b.Min.X+x, b.Min.Y+y).RGBA()
			lum[y*w+x] = float64(r>>8) / 255
		}
	}
	maxMag := 1e-9
	mag := make([]float64, w*h)
	for y := 1; y < h-1; y++ {
		for x := 1; x < w-1; x++ {
			i := y*w + x
			gx := -lum[i-w-1] + lum[i-w+1] - 2*lum[i-1] + 2*lum[i+1] - lum[i+w-1] + lum[i+w+1]
			gy := -lum[i-w-1] - 2*lum[i-w] - lum[i-w+1] + lum[i+w-1] + 2*lum[i+w] + lum[i+w+1]
			m := math.Hypot(gx, gy)
			mag[i] = m
			if m > maxMag {
				maxMag = m
			}
		}
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := y*w + x
			e := mag[i] / maxMag
			v := lum[i] * (1 - mix*e)
			if v < 0 {
				v = 0
			}
			g := uint8(math.Round(v * 255))
			out.SetNRGBA(b.Min.X+x, b.Min.Y+y, color.NRGBA{R: g, G: g, B: g, A: 255})
		}
	}
	return out
}
