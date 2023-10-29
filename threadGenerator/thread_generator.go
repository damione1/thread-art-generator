package threadGenerator

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"sync"
	"time"

	"github.com/disintegration/imaging"
)

type (
	Nail = image.Point

	ThreadGenerator struct {
		nailsQuantity     int
		imgSize           int
		maxPaths          int
		startingNail      int
		minimumDifference int
		brightnessFactor  int
		imageName         string
		imageContrast     float64
		physicalRadius    float64 // Radius of the circle in mm
		pathsDictionary   map[string][]Nail
		pathsList         []Path
		nailsList         []Nail
		pixelSize         float64
		threadLength      float64 // Length of the thread in mm
	}

	Path struct {
		StartingNail int
		EndingNail   int
	}

	Args struct {
		NailsQuantity     int
		ImgSize           int
		MaxPaths          int
		StartingNail      int
		MinimumDifference int
		BrightnessFactor  int
		ImageName         string
		PhysicalRadius    float64
	}

	OutputStats struct {
		TotalLines   int
		ThreadLength int
		TotalTime    time.Duration
	}

	weightResult struct {
		Weight  int
		Line    []image.Point
		NailIdx int
	}
)

func (tg *ThreadGenerator) getDefaults() {
	tg.nailsQuantity = 300
	tg.imgSize = 800
	tg.maxPaths = 10000
	tg.startingNail = 0
	tg.minimumDifference = 10
	tg.brightnessFactor = 50
	tg.imageContrast = 40
	tg.physicalRadius = 609.6 // 24 inches
}

func (tg *ThreadGenerator) mergeArgs(args Args) error {
	tg.getDefaults()

	if args.NailsQuantity > 0 {
		tg.nailsQuantity = args.NailsQuantity
	}
	if args.ImgSize > 0 {
		tg.imgSize = args.ImgSize
	}
	if args.MaxPaths > 0 {
		tg.maxPaths = args.MaxPaths
	}
	if args.StartingNail > 0 {
		tg.startingNail = args.StartingNail
	}
	if args.MinimumDifference > 0 {
		tg.minimumDifference = args.MinimumDifference
	}
	if args.BrightnessFactor > 0 {
		tg.brightnessFactor = args.BrightnessFactor
	}

	if args.PhysicalRadius > 0 {
		tg.physicalRadius = args.PhysicalRadius
	}

	tg.pixelSize = tg.physicalRadius / float64(tg.imgSize)

	if args.ImageName != "" {
		tg.imageName = args.ImageName
	} else {
		return errors.New("Image name is required")
	}

	return nil
}

func (tg *ThreadGenerator) Generate(args Args) (*OutputStats, error) {
	start := time.Now()
	err := tg.mergeArgs(args)
	if err != nil {
		return nil, err
	}

	sourceImage, err := tg.getSourceImage()
	if err != nil {
		return nil, err
	}

	nailsList := tg.getNailsListFromImage(sourceImage)

	tg.computePathsListFromImage(sourceImage, nailsList)

	return &OutputStats{
		TotalLines:   len(tg.pathsList),
		ThreadLength: int(tg.threadLength / 1000), //thread length from mm in meters
		TotalTime:    time.Since(start),
	}, nil
}

func (tg *ThreadGenerator) getSourceImage() (*image.NRGBA, error) {
	file, err := os.Open(tg.imageName)
	if err != nil {
		return nil, err
	}

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	imgGray := imaging.Grayscale(img)

	imgGray = imaging.AdjustContrast(imgGray, float64(tg.imageContrast))

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
				circleImg.Set(x, y, color.RGBA{255, 255, 255, 255})
			}
		}
	}

	circleImgMin := imaging.Resize(circleImg, tg.imgSize, tg.imgSize, imaging.Lanczos)

	return circleImgMin, nil
}

// getNailsListFromImage generates a list of nails from the source image in a circle
func (tg *ThreadGenerator) getNailsListFromImage(sourceImage image.Image) []Nail {
	centerX := sourceImage.Bounds().Dx() / 2
	centerY := sourceImage.Bounds().Dy() / 2
	radius := math.Min(float64(centerX), float64(centerY))
	tg.nailsList = make([]image.Point, tg.nailsQuantity)
	for i := 0; i < tg.nailsQuantity; i++ {
		alpha := float64(i) * 2 * math.Pi / float64(tg.nailsQuantity)
		x := centerX + int(radius*math.Cos(alpha))
		y := centerY + int(radius*math.Sin(alpha))
		tg.nailsList[i] = Nail{X: x, Y: y}
	}
	return tg.nailsList
}

// computePathsListFromImage generates a list of paths from the source image
func (tg *ThreadGenerator) computePathsListFromImage(sourceImage image.Image, nailsList []Nail) []Path {
	sourceImageBounds := sourceImage.Bounds()
	canvas := image.NewGray(sourceImageBounds)
	for y := sourceImageBounds.Min.Y; y < sourceImageBounds.Max.Y; y++ {
		for x := sourceImageBounds.Min.X; x < sourceImageBounds.Max.X; x++ {
			canvas.Set(x, y, sourceImage.At(x, y))
		}
	}

	tg.generateDictionary(nailsList)

	var nailIndex = tg.startingNail
	var pathsList = []Path{}
	usedPaths := make(map[string]bool)

	for i := 0; i < tg.maxPaths; i++ {
		// create a channel to gather results
		channel := make(chan weightResult, len(nailsList)-1)

		var wg sync.WaitGroup
		// loop through all possible next nails
		for nextnailIndex := 0; nextnailIndex < len(nailsList); nextnailIndex++ {
			// skip if the nail is the same
			if nailIndex == nextnailIndex {
				continue
			}

			wg.Add(1) // add a waitgroup before goroutine

			// calculate weight in a goroutine
			go func(nailIdx, nextnailIdx int) {
				defer wg.Done()
				weight := 0
				line := []image.Point{}
				difference := int(math.Abs(float64(nextnailIdx) - float64(nailIdx)))

				if difference < tg.minimumDifference || difference > (len(nailsList)-tg.minimumDifference) {
					return
				}

				if _, exists := usedPaths[tg.getPairKey(nextnailIdx, nailIdx)]; exists {
					return
				}

				line = tg.pathsDictionary[tg.getPairKey(nailIdx, nextnailIdx)]
				weight = len(line) * 255

				for _, pixelPosition := range line {
					pixelColor := canvas.GrayAt(pixelPosition.X, pixelPosition.Y).Y
					weight -= int(pixelColor)
				}

				weight = weight / len(line)

				if weight == 0 {
					return
				}

				// send the result through the channel
				channel <- weightResult{
					Weight:  weight,
					Line:    line,
					NailIdx: nextnailIdx,
				}

				return
			}(nailIndex, nextnailIndex) // pass nextnailIndex as an argument to avoid data race
		}

		//initialize maxWeight outside the loop
		maxWeight := 0
		var maxLine = []image.Point{}
		var maxnailIndex = 0
		wg.Wait() // wait for all goroutines to finish
		close(channel)

		// read from channel after closing it
		for res := range channel {
			if res.Weight > maxWeight {
				maxWeight = res.Weight
				maxLine = res.Line
				maxnailIndex = res.NailIdx
			}
		}

		if nailIndex == maxnailIndex {
			break
		}

		usedPaths[tg.getPairKey(nailIndex, maxnailIndex)] = true
		pathsList = append(pathsList, Path{nailIndex, maxnailIndex})
		tg.threadLength += tg.lineLength(nailIndex, maxnailIndex)
		nailIndex = maxnailIndex

		// Brighthen brightness of chosen line
		for _, pixelPosition := range maxLine {
			var pixel = int(canvas.GrayAt(pixelPosition.X, pixelPosition.Y).Y)
			value := uint8(min(255, pixel+tg.brightnessFactor))
			canvas.SetGray(pixelPosition.X, pixelPosition.Y, color.Gray{value})
		}

	}
	tg.pathsList = pathsList
	return pathsList
}

// GenerateDictionary generates a dictionary of all possible lines between nails
// It's way faster to generate all possible lines at the beginning than to calculate them on the fly
func (tg *ThreadGenerator) generateDictionary(nailsList []image.Point) map[string][]Nail {
	nailsQuantity := len(nailsList)
	tg.pathsDictionary = make(map[string][]Nail, nailsQuantity*(nailsQuantity-1)/2)

	for i := 0; i < nailsQuantity; i++ {
		for j := i + 1; j < nailsQuantity; j++ {
			tg.pathsDictionary[tg.getPairKey(i, j)] = tg.bresenham(nailsList[i], nailsList[j])
		}
	}
	return tg.pathsDictionary
}

// Bresenham's line algorithm - https://en.wikipedia.org/wiki/Bresenham%27s_line_algorithm
// Returns a list of points between two points
func (tg *ThreadGenerator) bresenham(startPoint, endPoint image.Point) []image.Point {
	xDifference := tg.abs(endPoint.X - startPoint.X)
	yDifference := -tg.abs(endPoint.Y - startPoint.Y)

	signX, signY := -1, -1

	// Determine direction for X
	if startPoint.X < endPoint.X {
		signX = 1
	}

	// Determine direction for Y
	if startPoint.Y < endPoint.Y {
		signY = 1
	}

	error := xDifference + yDifference

	var linePoints []image.Point
	// Continue until end point is reached
	for {
		linePoints = append(linePoints, startPoint)
		if startPoint == endPoint {
			break
		}
		errorDouble := 2 * error

		// Handle X direction
		if errorDouble >= yDifference {
			error += yDifference
			startPoint.X += signX
		}

		// Handle Y direction
		if errorDouble <= xDifference {
			error += xDifference
			startPoint.Y += signY
		}
	}
	return linePoints
}

func (tg *ThreadGenerator) abs(x int) int {
	return int(math.Abs(float64(x)))
}

// getPairKey returns a key for a map of lines between two points
func (tg *ThreadGenerator) getPairKey(a, b int) string {
	switch {
	case a < b:
		return fmt.Sprintf("%d:%d", a, b)
	case a > b:
		return fmt.Sprintf("%d:%d", b, a)
	default:
		return fmt.Sprintf("%d:%d", b, a)
	}
}

func (tg *ThreadGenerator) GeneratePathsImage() (image.Image, error) {
	if len(tg.pathsDictionary) == 0 {
		return nil, errors.New("Dictionary is empty")
	}

	pathsImage := image.NewGray(image.Rect(0, 0, tg.imgSize, tg.imgSize))

	for x := 0; x < tg.imgSize; x++ {
		for y := 0; y < tg.imgSize; y++ {
			pathsImage.SetGray(x, y, color.Gray{255})
		}
	}

	for i := 0; i < len(tg.pathsList); i++ {
		line := tg.pathsDictionary[tg.getPairKey(tg.pathsList[i].StartingNail, tg.pathsList[i].EndingNail)]
		for _, point := range line {
			currentValue := pathsImage.GrayAt(point.X, point.Y).Y
			newValue := max(int(currentValue)-20, 0)
			pathsImage.SetGray(point.X, point.Y, color.Gray{uint8(newValue)})
		}
	}

	return pathsImage, nil
}

func (tg *ThreadGenerator) GetPathsList() []Path {
	return tg.pathsList
}

func (tg *ThreadGenerator) GetGcode() []string {
	gCodeLines := []string{"G28 X0 Y0 Z0"} // GCode for homing X, Y, Z
	feedRate := 300                        // Define the feed rate - adjust value as necessary
	for i, path := range tg.pathsList {
		if i == 0 {
			gCodeLines = append(gCodeLines, fmt.Sprintf("G01 X%d F%d; Move to nail %d", path.StartingNail, feedRate, path.StartingNail))
			gCodeLines = append(gCodeLines, "M0 ; Pause")
		}
		// Calculate the delta between the starting and ending nails
		fromPin := path.StartingNail % tg.nailsQuantity
		toPin := path.EndingNail
		delta := toPin - fromPin

		// Convert delta to the nearest delta within half turn range
		if delta > tg.nailsQuantity/2 {
			delta -= tg.nailsQuantity
		} else if delta < -tg.nailsQuantity/2 {
			delta += tg.nailsQuantity
		}

		// Generate GCode lines for the thread movement
		gCodeLines = append(gCodeLines, tg.pinWrapGcode(fromPin, delta, feedRate)...)
	}
	return gCodeLines
}

func (tg *ThreadGenerator) pinWrapGcode(fromPin, delta, feedRate int) []string {
	gCodeLines := []string{}
	AxisXMax := 5
	AxisXMin := 0
	nailOffset := 0.5

	// Move to the nail position minus the offset
	startPos := fmt.Sprintf("G01 Z%.2f F%d; Move to nail %d", float64(fromPin)+nailOffset, feedRate, fromPin)
	gCodeLines = append(gCodeLines, startPos)

	// Retract the needle
	moveXMax := fmt.Sprintf("G01 X%d F%d", AxisXMax, feedRate*5)
	gCodeLines = append(gCodeLines, moveXMax)

	// Move to the nail position plus the offset to pass the thread around the nail
	endPos := fmt.Sprintf("G01 Z%.2f F%d", float64(fromPin+delta)+nailOffset, feedRate)
	gCodeLines = append(gCodeLines, endPos)

	// Move back the needle to the starting position
	moveXMin := fmt.Sprintf("G01 X%d F%d", AxisXMin, feedRate*5)
	gCodeLines = append(gCodeLines, moveXMin)

	return gCodeLines
}

func (tg *ThreadGenerator) lineLength(startNail, endNail int) float64 {
	pixels := tg.pathsDictionary[tg.getPairKey(startNail, endNail)]
	distance := float64(len(pixels)) * tg.pixelSize // multiply with the size of a pixel
	return distance
}
