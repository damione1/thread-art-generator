package threadGenerator

import (
	"errors"
	"image"
	"math"
)

func init() {
	Register(vrellis{})
}

type vrellis struct{}

func (vrellis) ID() Kind             { return KindVrellis }
func (vrellis) FormValue() string    { return "VRELLIS" }
func (vrellis) Label() string        { return "Vrellis" }
func (vrellis) Hint() string         { return "Consecutive darkest chord. Fast, one continuous walk." }
func (vrellis) UsesBrightness() bool { return true }

func (vrellis) Solve(tg *ThreadGenerator) error {
	sourceImage, err := tg.getSourceImage()
	if err != nil {
		return err
	}
	nailsList := tg.getNailsListFromImage(sourceImage)
	tg.computePathsListFromImage(sourceImage, nailsList)
	return nil
}

func (vrellis) RenderPreview(tg *ThreadGenerator) (image.Image, error) {
	if len(tg.pathsDictionary) == 0 {
		return nil, errors.New("Dictionary is empty")
	}

	w := tg.imgSize
	light := make([]float64, w*w)
	for i := range light {
		light[i] = 1
	}

	halfW := tg.threadHalfWidthPx()
	for _, path := range tg.pathsList {
		a := tg.nailsList[path.StartingNail]
		b := tg.nailsList[path.EndingNail]
		stampThread(light, w, w, float64(a.X), float64(a.Y), float64(b.X), float64(b.Y), halfW, threadAbsorb)
	}

	out := image.NewGray(image.Rect(0, 0, w, w))
	for i, t := range light {
		if t < 0 {
			t = 0
		}
		if t > 1 {
			t = 1
		}
		out.Pix[i] = uint8(math.Round(255 * t))
	}
	return out, nil
}
