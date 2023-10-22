package main

import (
	"fmt"
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

type Line struct {
	StartingNail int
	EndingNail   int
}

var (
	NailsQuantity = 300
	ImgSize       = 500
	MaxLines      = 10000
	StartingNail  = 0
	OutputFolder  = "output"
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

	// Crop it into a square if it's not already but don't distort it
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
	outBaseFile, err := os.Create(OutputFolder + "/output_1_base.jpg")
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

	nailsList := make([]image.Point, NailsQuantity)
	for i := 0; i < NailsQuantity; i++ {
		// calculate the angle and the point coordinates
		alpha := float64(i) * 2 * math.Pi / float64(NailsQuantity)
		x := centerX + int(r*math.Cos(alpha))
		y := centerY + int(r*math.Sin(alpha))
		imagePoints.Set(x, y, color.RGBA{255, 0, 0, 255})

		nailsList[i] = image.Point{X: x, Y: y}
	}

	// Save the image
	outPointFile, err := os.Create(OutputFolder + "output_1_points.jpg")
	if err != nil {
		panic(err)
	}
	err = jpeg.Encode(outPointFile, imagePoints, nil)
	if err != nil {
		panic(err)
	}

	// Create a new image with the same dimensions as the source image
	canvas := image.NewGray(imagePoints.Bounds())
	// Fill the canvas with white
	draw.Draw(canvas, canvas.Bounds(), image.NewUniform(color.White), image.ZP, draw.Src)
	// Set the line color to black
	lineColor := color.Black
	sourceImage := imagePoints
	threadList := make([]Line, MaxLines)
	currentNail := StartingNail
	for threadIndex := 0; threadIndex < MaxLines; threadIndex++ {
		// Find the best line between all the nails
		nextNail, threadPath := findBestLine(currentNail, nailsList, sourceImage, canvas)

		// Draw the best line on the canvas
		drawLine(canvas, threadPath, lineColor)

		// Save the best line for later usage
		threadList[threadIndex] = Line{StartingNail: currentNail, EndingNail: nextNail}
		currentNail = nextNail
	}

	// Save the image
	outThreadFile, err := os.Create(OutputFolder + "output_1_threads.jpg")
	if err != nil {
		panic(err)
	}
	err = jpeg.Encode(outThreadFile, canvas, nil)
	if err != nil {
		panic(err)
	}

	//write to txt file
	f, err := os.Create(OutputFolder + "output_1_threads.txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	for _, thread := range threadList {
		_, err := f.WriteString(fmt.Sprintf("%v %v\n", thread.StartingNail, thread.EndingNail))
		if err != nil {
			panic(err)
		}
	}

}

func findBestLine(currentNail int, nailsList []image.Point, sourceImage image.Image, canvas draw.Image) (int, []image.Point) {
	var bestScore float64 = math.MaxFloat64
	var bestPath []image.Point
	var bestEndingNail int

	// Iterate over all the nails
	for nailIndex, nail := range nailsList {
		if nailIndex == currentNail {
			continue
		}

		// Calculate the path between the two nails
		path := calculatePathWithBresenham(nail, nailsList[currentNail], sourceImage)

		// Get path score
		score := calculatePathScore(path, sourceImage, canvas)

		// If the score is better than the best score, save it
		if score < bestScore {
			bestScore = score
			bestPath = path
			bestEndingNail = nailIndex
		}
	}

	return bestEndingNail, bestPath
}

func calculatePathWithBresenham(start, end image.Point, sourceImage image.Image) []image.Point {
	// Calculate the path between the two nails
	path := bresenham(start, end)
	return path
}

func calculatePathScore(path []image.Point, sourceImage image.Image, canvas image.Image) float64 {
	var score float64
	for _, point := range path {

		sourcePixel := color.GrayModel.Convert(sourceImage.At(point.X, point.Y)).(color.Gray).Y
		canvasPixel := color.GrayModel.Convert(canvas.At(point.X, point.Y)).(color.Gray).Y

		// check if a thread has already covered this area of the canvas
		if canvasPixel > 0 {
			// Penalize paths that overlap with existing thread,
			// but still allow for potential re-threading in the case the colors are very close.
			score -= math.Abs(float64(sourcePixel - canvasPixel))
		} else {
			// Weigh the score by the intensity of the pixel in the source,
			// this will favor using darker source pixels
			score += float64(sourcePixel)
		}
	}
	return score
}

// Bresenham's line algorithm
func bresenham(start, end image.Point) []image.Point {
	dx := abs(end.X - start.X)
	dy := -abs(end.Y - start.Y)
	sx, sy := -1, -1
	if start.X < end.X {
		sx = 1
	}
	if start.Y < end.Y {
		sy = 1
	}
	err := dx + dy // error value e_xy

	var points []image.Point
	for { // loop
		points = append(points, start)
		if start == end {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			start.X += sx
		}
		if e2 <= dx {
			err += dx
			start.Y += sy
		}
	}
	return points
}

func abs(x int) int {
	return int(math.Abs(float64(x)))
}

func drawLine(canvas draw.Image, threadPath []image.Point, lineColor color.Color) {
	// Create a new mask image
	mask := image.NewRGBA(canvas.Bounds())

	// Define a color with a 20% alpha
	translucentBlack := color.RGBA{R: 0, G: 0, B: 0, A: 51}

	// Draw the line onto the mask
	for _, point := range threadPath {
		mask.Set(point.X, point.Y, translucentBlack)
	}

	// Apply the mask
	draw.DrawMask(canvas, canvas.Bounds(), &image.Uniform{translucentBlack}, image.ZP, mask, image.ZP, draw.Over)
}
