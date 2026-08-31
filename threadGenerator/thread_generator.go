package threadGenerator

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"os"
	"time"

	"github.com/disintegration/imaging"
)

type (
	Nail = image.Point

	ThreadGenerator struct {
		nailsQuantity       int
		imgSize             int
		maxPaths            int
		startingNail        int
		minimumDifference   int
		brightnessFactor    int
		imageName           string
		imageContrast       float64
		physicalRadius      float64 // Radius of the circle in mm
		pathsDictionary     map[int64][]Nail
		pathsList           []Path
		nailsList           []Nail
		pixelSize           float64
		threadLength        float64 // Length of the thread in mm
		rotationAxis        string
		needleAxis          string
		spindleAxis         string
		stopWeightThreshold int
		nailCooldown        int
		threadDiameterMM    float64
		algorithm           Kind
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

	// Config holds all possible configuration options for ThreadGenerator
	Config struct {
		NailsQuantity       int     // Number of nails around the circle
		ImgSize             int     // Size of the image in pixels
		MaxPaths            int     // Maximum number of paths to generate
		StartingNail        int     // Starting nail index
		MinimumDifference   int     // Minimum difference between nails
		BrightnessFactor    int     // Brightness factor for line drawing
		ImageContrast       float64 // Contrast adjustment in percent (-100..100, typical 1–100)
		PhysicalRadius      float64 // Physical radius in mm
		RotationAxis        string  // Rotation axis name
		NeedleAxis          string  // Needle axis name
		SpindleAxis         string  // Spindle axis name
		StopWeightThreshold int     // Stop when best line darkness is at or below this (0–255)
		NailCooldown        int     // Do not revisit a nail used in the last N steps
		ThreadDiameterMM    float64 // Physical thread diameter in mm (ordinary polyester ≈ 0.3)
		Algorithm           Kind    // Path-selection algorithm
	}

	OutputStats struct {
		TotalLines   int
		ThreadLength int // millimetres
		TotalTime    time.Duration
	}

	// ParamError is a field-level config problem (API + generator).
	ParamError struct {
		Field   string
		Message string
	}
)

const (
	defaultThreadDiameterMM = 0.3
	threadAbsorb            = 0.78
)

func DefaultConfig() Config {
	return Config{
		NailsQuantity:       280,
		ImgSize:             800,
		MaxPaths:            4500,
		StartingNail:        0,
		MinimumDifference:   22,
		BrightnessFactor:    40,
		ImageContrast:       28,    // percent → CLAHE clip 1 + contrast/50
		PhysicalRadius:      304.8, // 24-inch hoop diameter
		RotationAxis:        "A",
		NeedleAxis:          "X",
		SpindleAxis:         "Y",
		StopWeightThreshold: 10,
		NailCooldown:        3,
		ThreadDiameterMM:    defaultThreadDiameterMM,
		Algorithm:           KindVrellis,
	}
}

// ValidateParams checks cross-field constraints that proto cannot express.
func ValidateParams(nails, startingNail, minDiff int) []ParamError {
	var errs []ParamError
	if nails < 3 {
		errs = append(errs, ParamError{
			Field:   "nails_quantity",
			Message: "need at least 3 nails",
		})
	}
	if nails >= 3 && (startingNail < 0 || startingNail >= nails) {
		errs = append(errs, ParamError{
			Field:   "starting_nail",
			Message: "must be less than the number of nails",
		})
	}
	if nails >= 3 && (minDiff < 1 || minDiff >= nails/2) {
		errs = append(errs, ParamError{
			Field:   "minimum_difference",
			Message: "must be less than half the number of nails",
		})
	}
	return errs
}

// NewThreadGenerator creates a new ThreadGenerator with the given configuration
func NewThreadGenerator(config Config) *ThreadGenerator {
	diameter := config.ThreadDiameterMM
	if diameter <= 0 {
		diameter = defaultThreadDiameterMM
	}
	return &ThreadGenerator{
		nailsQuantity:       config.NailsQuantity,
		imgSize:             config.ImgSize,
		maxPaths:            config.MaxPaths,
		startingNail:        config.StartingNail,
		minimumDifference:   config.MinimumDifference,
		brightnessFactor:    config.BrightnessFactor,
		imageContrast:       config.ImageContrast,
		physicalRadius:      config.PhysicalRadius,
		rotationAxis:        config.RotationAxis,
		needleAxis:          config.NeedleAxis,
		spindleAxis:         config.SpindleAxis,
		stopWeightThreshold: config.StopWeightThreshold,
		nailCooldown:        config.NailCooldown,
		threadDiameterMM:    diameter,
		algorithm:           normalizeKind(config.Algorithm),
		pixelSize:           pixelSizeMM(config.PhysicalRadius, config.ImgSize),
	}
}

func pixelSizeMM(radius float64, imgSize int) float64 {
	if imgSize <= 0 {
		return 0
	}
	return 2 * radius / float64(imgSize)
}

// SetImage sets the image to process
func (tg *ThreadGenerator) SetImage(imagePath string) {
	tg.imageName = imagePath
}

func (tg *ThreadGenerator) mergeArgs(args Args) error {
	if args.NailsQuantity > 0 {
		tg.nailsQuantity = args.NailsQuantity
	}
	if args.ImgSize > 0 {
		tg.imgSize = args.ImgSize
	}
	if args.MaxPaths > 0 {
		tg.maxPaths = args.MaxPaths
	}
	if args.StartingNail >= 0 {
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

	if args.ImgSize > 0 || args.PhysicalRadius > 0 {
		tg.pixelSize = pixelSizeMM(tg.physicalRadius, tg.imgSize)
	}

	if args.ImageName != "" {
		tg.imageName = args.ImageName
	} else {
		return errors.New("Image name is required")
	}

	return nil
}

// Generate processes the image and creates thread art based on configuration
func (tg *ThreadGenerator) Generate(args Args) (*OutputStats, error) {
	start := time.Now()

	if args.ImageName != "" &&
		args.NailsQuantity == 0 &&
		args.ImgSize == 0 &&
		args.MaxPaths == 0 &&
		args.StartingNail == 0 &&
		args.MinimumDifference == 0 &&
		args.BrightnessFactor == 0 &&
		args.PhysicalRadius == 0 {
		tg.imageName = args.ImageName
	} else {
		err := tg.mergeArgs(args)
		if err != nil {
			return nil, err
		}
	}

	if errs := ValidateParams(tg.nailsQuantity, tg.startingNail, tg.minimumDifference); len(errs) > 0 {
		return nil, fmt.Errorf("%s: %s", errs[0].Field, errs[0].Message)
	}

	if err := Lookup(tg.algorithm).Solve(tg); err != nil {
		return nil, err
	}

	return &OutputStats{
		TotalLines:   len(tg.pathsList),
		ThreadLength: int(math.Round(tg.threadLength)), // millimetres
		TotalTime:    time.Since(start),
	}, nil
}

func (tg *ThreadGenerator) getSourceImage() (*image.NRGBA, error) {
	file, err := os.Open(tg.imageName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	return tg.prepareGraySquare(img, tg.imgSize), nil
}

func maskCircle(img *image.NRGBA) *image.NRGBA {
	b := img.Bounds()
	out := imaging.Clone(img)
	w, h := b.Dx(), b.Dy()
	midX, midY := w/2, h/2
	r2 := midX * midX
	if midY*midY < r2 {
		r2 = midY * midY
	}
	white := color.NRGBA{255, 255, 255, 255}
	for y := 0; y < h; y++ {
		dy := y - midY
		for x := 0; x < w; x++ {
			dx := x - midX
			if dx*dx+dy*dy > r2 {
				out.SetNRGBA(b.Min.X+x, b.Min.Y+y, white)
			}
		}
	}
	return out
}

// getNailsListFromImage generates a list of nails from the source image in a circle
func (tg *ThreadGenerator) getNailsListFromImage(sourceImage image.Image) []Nail {
	centerX := sourceImage.Bounds().Dx() / 2
	centerY := sourceImage.Bounds().Dy() / 2
	radius := math.Min(float64(centerX), float64(centerY)) - 1
	if radius < 1 {
		radius = 1
	}
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
	draw.Draw(canvas, canvas.Bounds(), sourceImage, sourceImageBounds.Min, draw.Src)

	tg.generateDictionary(nailsList)

	nailCount := len(nailsList)
	nailIndex := tg.startingNail
	pathsList := []Path{}
	usedPaths := make(map[int64]struct{})
	recent := make([]int, 0, tg.nailCooldown)
	if tg.nailCooldown > 0 {
		recent = append(recent, nailIndex)
	}

	for i := 0; i < tg.maxPaths; i++ {
		maxWeight := 0
		var maxLine []image.Point
		maxNailIndex := -1

		for next := 0; next < nailCount; next++ {
			if next == nailIndex {
				continue
			}
			if circularDiff(nailIndex, next, nailCount) < tg.minimumDifference {
				continue
			}
			if inList(next, recent) {
				continue
			}
			key := pairKey(nailIndex, next)
			if _, used := usedPaths[key]; used {
				continue
			}
			line := tg.pathsDictionary[key]
			if len(line) == 0 {
				continue
			}

			sum := 0
			for _, pixelPosition := range line {
				sum += 255 - int(canvas.GrayAt(pixelPosition.X, pixelPosition.Y).Y)
			}
			weight := sum / len(line)
			if weight > maxWeight {
				maxWeight = weight
				maxLine = line
				maxNailIndex = next
			}
		}

		if maxNailIndex < 0 || maxWeight <= tg.stopWeightThreshold {
			break
		}

		usedPaths[pairKey(nailIndex, maxNailIndex)] = struct{}{}
		pathsList = append(pathsList, Path{nailIndex, maxNailIndex})
		tg.threadLength += tg.lineLength(nailIndex, maxNailIndex)
		nailIndex = maxNailIndex

		if tg.nailCooldown > 0 {
			recent = append(recent, maxNailIndex)
			if len(recent) > tg.nailCooldown {
				recent = recent[len(recent)-tg.nailCooldown:]
			}
		}

		for _, pixelPosition := range maxLine {
			tg.paintThread(canvas, pixelPosition)
		}
	}
	tg.pathsList = pathsList
	return pathsList
}

// generateDictionary generates a dictionary of lines between nails that are
// far enough apart to be legal moves.
func (tg *ThreadGenerator) generateDictionary(nailsList []image.Point) map[int64][]Nail {
	nailsQuantity := len(nailsList)
	tg.pathsDictionary = make(map[int64][]Nail, nailsQuantity*(nailsQuantity-1)/2)

	for i := 0; i < nailsQuantity; i++ {
		for j := i + 1; j < nailsQuantity; j++ {
			if circularDiff(i, j, nailsQuantity) < tg.minimumDifference {
				continue
			}
			tg.pathsDictionary[pairKey(i, j)] = tg.bresenham(nailsList[i], nailsList[j])
		}
	}
	return tg.pathsDictionary
}

// Bresenham's line algorithm - https://en.wikipedia.org/wiki/Bresenham%27s_line_algorithm
func (tg *ThreadGenerator) bresenham(startPoint, endPoint image.Point) []image.Point {
	xDifference := absInt(endPoint.X - startPoint.X)
	yDifference := -absInt(endPoint.Y - startPoint.Y)

	signX, signY := -1, -1
	if startPoint.X < endPoint.X {
		signX = 1
	}
	if startPoint.Y < endPoint.Y {
		signY = 1
	}

	err := xDifference + yDifference

	var linePoints []image.Point
	for {
		linePoints = append(linePoints, startPoint)
		if startPoint == endPoint {
			break
		}
		errorDouble := 2 * err

		if errorDouble >= yDifference {
			err += yDifference
			startPoint.X += signX
		}

		if errorDouble <= xDifference {
			err += xDifference
			startPoint.Y += signY
		}
	}
	return linePoints
}

func pairKey(a, b int) int64 {
	if a > b {
		a, b = b, a
	}
	return int64(a)<<32 | int64(uint32(b))
}

func circularDiff(a, b, n int) int {
	if n <= 0 {
		return 0
	}
	d := a - b
	if d < 0 {
		d = -d
	}
	if wrap := n - d; wrap < d {
		return wrap
	}
	return d
}

func inList(idx int, list []int) bool {
	for _, v := range list {
		if v == idx {
			return true
		}
	}
	return false
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (tg *ThreadGenerator) GeneratePathsImage() (image.Image, error) {
	return Lookup(tg.algorithm).RenderPreview(tg)
}

func (tg *ThreadGenerator) threadHalfWidthPx() float64 {
	d := tg.threadDiameterMM
	if d <= 0 {
		d = defaultThreadDiameterMM
	}
	if tg.pixelSize <= 0 {
		return 0.5
	}
	return 0.5 * d / tg.pixelSize
}

func (tg *ThreadGenerator) paintThread(canvas *image.Gray, p image.Point) {
	r := int(math.Ceil(tg.threadHalfWidthPx() + 0.5))
	if r < 1 {
		r = 1
	}
	b := canvas.Bounds()
	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			x, y := p.X+dx, p.Y+dy
			if !image.Pt(x, y).In(b) {
				continue
			}
			pixel := int(canvas.GrayAt(x, y).Y)
			canvas.SetGray(x, y, color.Gray{Y: uint8(min(255, pixel+tg.brightnessFactor))})
		}
	}
}

func stampThread(buf []float64, w, h int, x0, y0, x1, y1, halfW, absorb float64) {
	pad := halfW + 1
	minX := max(0, int(math.Floor(min(x0, x1)-pad)))
	maxX := min(w-1, int(math.Ceil(max(x0, x1)+pad)))
	minY := max(0, int(math.Floor(min(y0, y1)-pad)))
	maxY := min(h-1, int(math.Ceil(max(y0, y1)+pad)))
	for y := minY; y <= maxY; y++ {
		py := float64(y) + 0.5
		for x := minX; x <= maxX; x++ {
			d := distToSeg(float64(x)+0.5, py, x0, y0, x1, y1)
			c := halfW + 0.5 - d
			if c <= 0 {
				continue
			}
			if c > 1 {
				c = 1
			}
			buf[y*w+x] *= 1 - absorb*c
		}
	}
}

func distToSeg(px, py, x0, y0, x1, y1 float64) float64 {
	dx, dy := x1-x0, y1-y0
	l2 := dx*dx + dy*dy
	if l2 == 0 {
		return math.Hypot(px-x0, py-y0)
	}
	t := ((px-x0)*dx + (py-y0)*dy) / l2
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	return math.Hypot(px-(x0+t*dx), py-(y0+t*dy))
}

func (tg *ThreadGenerator) GetPathsList() []Path {
	return tg.pathsList
}

func (tg *ThreadGenerator) GetGcode() []string {
	gCodeLines := []string{fmt.Sprintf("G28 %s5 %s0 %s0", tg.needleAxis, tg.spindleAxis, tg.rotationAxis)} // GCode for homing
	feedRate := 3000
	var nailOffset float32 = 0.5
	for i, path := range tg.pathsList {
		if i == 0 {
			gCodeLines = append(gCodeLines, fmt.Sprintf("G01 %s%d F%d; Move to nail %d", tg.needleAxis, path.StartingNail, feedRate, path.StartingNail))
			gCodeLines = append(gCodeLines, "M0 ; Pausing to allow for thread to be attached")
		}
		fromPin := path.StartingNail % tg.nailsQuantity
		toPin := path.EndingNail

		delta := toPin - (fromPin % tg.nailsQuantity)

		if abs(delta) < (tg.nailsQuantity / 2) {
			move := tg.moveToPin(toPin, feedRate, nailOffset)
			gCodeLines = append(gCodeLines, move)
		} else {
			gCodeLines = append(gCodeLines, "G91 ; Switch to relative positioning mode")
			toPinRelative := tg.nailsQuantity - (fromPin % tg.nailsQuantity) + toPin
			if delta > 0 {
				toPinRelative = -(tg.nailsQuantity - abs(delta))
			}
			gCodeLines = append(gCodeLines, tg.moveByDelta(toPinRelative, toPin, feedRate, nailOffset))
			gCodeLines = append(gCodeLines, "G90 ; Switch back to absolute positioning mode")
			gCodeLines = append(gCodeLines, fmt.Sprintf("G92 %s%.2f; Set current position to %.2f", tg.rotationAxis, float32(toPin)-nailOffset, float32(toPin)-nailOffset))
		}
		gCodeLines = append(gCodeLines, tg.pinWrapGcode(fromPin, toPin, nailOffset)...)
	}
	return gCodeLines
}

func (tg *ThreadGenerator) pinWrapGcode(fromPin, toPin int, nailOffset float32) []string {
	gCodeLines := []string{}
	AxisXMax := -10
	AxisXMin := 0
	feedrateBetweenNails := 200
	nailFeedRate := 2000

	moveXMax := fmt.Sprintf("G01 %s%d F%d", tg.needleAxis, AxisXMax, nailFeedRate)
	gCodeLines = append(gCodeLines, moveXMax)

	endPos := fmt.Sprintf("G01 %s%.2f F%d", tg.rotationAxis, float32(toPin)+nailOffset, feedrateBetweenNails)
	gCodeLines = append(gCodeLines, endPos)

	moveXMin := fmt.Sprintf("G01 %s%d F%d", tg.needleAxis, AxisXMin, nailFeedRate)
	gCodeLines = append(gCodeLines, moveXMin)

	return gCodeLines
}

func (tg *ThreadGenerator) moveToPin(pin, feedrate int, nailOffset float32) string {
	return fmt.Sprintf("G01 %s%.2f F%d; Move to nail %d", tg.rotationAxis, float32(pin)-nailOffset, feedrate, pin)
}

func (tg *ThreadGenerator) moveByDelta(delta, nail, feedrate int, nailOffset float32) string {
	return fmt.Sprintf("G01 %s%.2f F%d; Move by delta %d (nail %d)", tg.rotationAxis, float32(delta)-nailOffset, feedrate, delta, nail)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (tg *ThreadGenerator) GenerateHolesGcode() []string {
	rotationSpeed := 200
	feedRateIn := 170
	feedRateOut := 1000
	AxisYMin := -0.5
	AxisYMax := -3.20

	gCodeLines := []string{fmt.Sprintf("G28 %s0 %s0", tg.spindleAxis, tg.rotationAxis)} // GCode for homing

	for i := 0; i < tg.nailsQuantity; i++ {
		gCodeLines = append(gCodeLines, fmt.Sprintf("G01 %s%d F%d; Move to nail %d", tg.rotationAxis, i, rotationSpeed, i))
		gCodeLines = append(gCodeLines, fmt.Sprintf("G01 %s%.2f F%d; Drill hole at nail %d", tg.spindleAxis, AxisYMax, feedRateIn, i))
		gCodeLines = append(gCodeLines, fmt.Sprintf("G01 %s%.2f F%d; Retract needle", tg.spindleAxis, AxisYMin, feedRateOut))
	}

	return gCodeLines
}

func (tg *ThreadGenerator) lineLength(startNail, endNail int) float64 {
	a := tg.nailsList[startNail]
	b := tg.nailsList[endNail]
	dx := float64(a.X - b.X)
	dy := float64(a.Y - b.Y)
	return math.Hypot(dx, dy) * tg.pixelSize
}
