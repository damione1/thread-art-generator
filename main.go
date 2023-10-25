package main

import (
	"fmt"
	"image"
	"image/jpeg"
	_ "image/jpeg"
	"math/rand"
	"os"

	"github.com/Damione1/strings-art-generator/threadGenerator"
)

type Nail = image.Point

type Coordinate struct {
	X float64
	Y float64
}

type Line struct {
	StartingNail int
	EndingNail   int
}

var (
	NailsQuantity     = 300
	ImgSize           = 800
	MaxLines          = 10000
	StartingNail      = rand.Intn(NailsQuantity)
	OutputFolder      = "output/"
	minimumDifference = 10
	brightnessFactor  = 50
)

func main() {
	tg := new(threadGenerator.ThreadGenerator)

	args := threadGenerator.Args{
		NailsQuantity:     NailsQuantity,
		ImgSize:           ImgSize,
		MaxPaths:          MaxLines,
		StartingNail:      StartingNail,
		MinimumDifference: minimumDifference,
		BrightnessFactor:  brightnessFactor,
		ImageName:         "source_1.jpg",
	}

	stats, err := tg.Generate(args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	pathsImage, err := tg.GeneratePathsImage()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Save the image
	outThreadFile, err := os.Create(OutputFolder + "output_1_threads.jpg")
	if err != nil {
		panic(err)
	}
	defer outThreadFile.Close()
	err = jpeg.Encode(outThreadFile, pathsImage, nil)
	if err != nil {
		panic(err)
	}

	pathsList := tg.GetPathsList()
	// Save the thread list
	outThreadListFile, err := os.Create(OutputFolder + "output_1_threads.txt")
	if err != nil {
		panic(err)
	}
	defer outThreadListFile.Close()

	for _, line := range pathsList {
		_, err := outThreadListFile.WriteString(fmt.Sprintf("%v	%v\n", line.StartingNail, line.EndingNail))
		if err != nil {
			panic(err)
		}
	}

	gCodeLines := tg.GetGcode()
	// Save the thread list
	outGcodeFile, err := os.Create(OutputFolder + "output_1_gcode.gcode")
	if err != nil {
		panic(err)
	}
	defer outGcodeFile.Close()

	for _, line := range gCodeLines {
		_, err := outGcodeFile.WriteString(line + "\n")
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("Done in", stats.TotalTime)
	fmt.Println("Number of lines", stats.TotalLines)
}
