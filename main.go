package main

import (
	"fmt"
	"image/jpeg"
	_ "image/jpeg"
	"os"

	"github.com/Damione1/thread-art-generator/threadGenerator"
)

var OutputFolder = "output/"

func main() {
	tg := new(threadGenerator.ThreadGenerator)

	args := threadGenerator.Args{
		NailsQuantity:     300,
		ImgSize:           800,
		MaxPaths:          10000,
		StartingNail:      0,
		MinimumDifference: 10,
		BrightnessFactor:  50,
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
	fmt.Println(fmt.Sprintf("Estimated thread length: %d meters", stats.ThreadLength))
}
