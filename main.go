package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	_ "image/jpeg"
	"math"
	"os"

	"github.com/disintegration/imaging"
)

type Nail struct {
	X int
	Y int
}

type Coordinate struct {
	X float64
	Y float64
}

var (
	NailsQuantity = 300
	ImgSize       = 500
	MaxLines      = 20000
)

func main() {
	// Open a file
	file, err := os.Open("source_1.jpg")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		panic(err)
	}

	// Convert image to grayscale
	imgGray := imaging.Grayscale(img)

	// Crop it into a square if it's not already
	imgSquare := imgGray
	bounds := imgSquare.Bounds()
	if bounds.Dx() != bounds.Dy() {
		imgSquare = imaging.CropAnchor(imgSquare, bounds.Dx(), bounds.Dx(), imaging.Center)
	}

	// Crop it into a circle
	circleImg := image.NewRGBA(bounds)
	midPoint := bounds.Dx() / 2
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			xx, yy := float64(x-midPoint), float64(y-midPoint)
			if xx*xx+yy*yy <= float64(midPoint*midPoint) {
				circleImg.Set(x, y, imgSquare.At(x, y))
			} else {
				// Set white background
				circleImg.Set(x, y, color.RGBA{255, 255, 255, 255})
			}
		}
	}

	//rezise image to 500x500
	circleImgMin := imaging.Resize(circleImg, ImgSize, ImgSize, imaging.Lanczos)

	// Save the resulting base image to disk
	outBaseFile, err := os.Create("output_1_base.jpg")
	if err != nil {
		panic(err)
	}
	defer outBaseFile.Close()
	err = jpeg.Encode(outBaseFile, circleImgMin, nil)
	if err != nil {
		panic(err)
	}

	// Create a new image with the same dimensions as the source image
	imagePoints := circleImgMin
	draw.Draw(imagePoints, imagePoints.Bounds(), circleImgMin, circleImgMin.Bounds().Min, draw.Src)

	centerX := imagePoints.Bounds().Dx() / 2
	centerY := imagePoints.Bounds().Dy() / 2
	r := math.Min(float64(centerX), float64(centerY)) // take the smaller dimension as the radius

	nails := make([]Nail, NailsQuantity)
	for i := 0; i < NailsQuantity; i++ {
		// calculate the angle and the point coordinates
		alpha := float64(i) * 2 * math.Pi / float64(NailsQuantity)
		x := centerX + int(r*math.Cos(alpha))
		y := centerY + int(r*math.Sin(alpha))
		imagePoints.Set(x, y, color.RGBA{255, 0, 0, 255})

		nails[i] = Nail{X: x, Y: y}
	}

	// Save the image
	outPointFile, err := os.Create("output_1_points.jpg")
	if err != nil {
		panic(err)
	}

	err = jpeg.Encode(outPointFile, imagePoints, nil)
	if err != nil {
		panic(err)
	}

	// Step 2: New blank canvas
	imagePointsBounds := imagePoints.Bounds()
	canvas := image.NewRGBA(bounds)
	draw.Draw(canvas, bounds, imagePoints, imagePointsBounds.Min, draw.Src)

}

func getLineTraceWithBresenham(i, j int) []Coordinate {
	x0 := i
	y0 := j
	x1 := i + 1
	y1 := j + 1

	dx := math.Abs(float64(x1 - x0))
	dy := math.Abs(float64(y1 - y0))

	var sx, sy int
	if x0 < x1 {
		sx = 1
	} else {
		sx = -1
	}
	if y0 < y1 {
		sy = 1
	} else {
		sy = -1
	}

	err := dx - dy

	var res []Coordinate

	for {
		res = append(res, Coordinate{X: float64(x0), Y: float64(y0)})

		if x0 == x1 && y0 == y1 {
			break
		}

		e2 := 2 * err
		if e2 > -dy {
			err = err - dy
			x0 = x0 + sx
		}
		if e2 < dx {
			err = err + dx
			y0 = y0 + sy
		}
	}

	return res
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func findBestLine(srcImage, canvas image.Image, points []image.Point) (image.Point, image.Point) {
	bestLine := make([]image.Point, 2)
	maxDelta := 0.0

	// Calculate the total color difference between the srcImage and canvas (I use grayscale color here)
	srcGray := image.NewGray(srcImage.Bounds())
	draw.Draw(srcGray, srcGray.Bounds(), srcImage, srcImage.Bounds().Min, draw.Src)
	deltaImage := image.NewGray(canvas.Bounds())
	for i := 0; i < len(deltaImage.Pix); i++ {
		deltaImage.Pix[i] = uint8(math.Abs(float64(srcGray.Pix[i]) - float64(canvas.(*image.Gray).Pix[i])))
	}

	// Iterate over every pair of points looking for the line that reduces the delta the most
	for i := 0; i < len(points); i++ {
		for j := i + 1; j < len(points); j++ {
			currDelta := getLineScore(deltaImage, points[i], points[j]) // score lines based on how much they would reduce image delta
			if currDelta > maxDelta {
				maxDelta = currDelta
				bestLine[0] = points[i]
				bestLine[1] = points[j]
			}
		}
	}
	return bestLine[0], bestLine[1]
}

// Scores a line based on how much it would reduce the image delta
// You can consider alternative scoring metrics based on color, saturation, hue, etc.
func getLineScore(deltaImage *image.Gray, p1, p2 image.Point) float64 {
	score := 0.0
	for _, p := range getLineTrace(p1, p2) {
		score += float64(deltaImage.GrayAt(p.X, p.Y).Y)
	}
	return score
}

func getLineTrace(a, b image.Point) []image.Point {
	// This should use your Bresenham's algorithm implementation to generate
	// points along the line from `a` to `b`.
	return []image.Point{}
}
