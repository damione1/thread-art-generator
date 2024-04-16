# Thread Art Generator

![thread art generator](https://github.com/Damione1/thread-art-generator/assets/14912510/6b6ef9e1-9bad-4dd7-8579-17fe55ae9c13)

Thread Art Generator is a project that allows you to transform your images into unique pieces of circular thread art. This project is written in Golang, a powerful programming language renowned for its efficiency and simplicity.

## Introduction

The principle behind Thread Art Generator is simple: you provide an image, and the program generates a series of thread lines which, when arranged on a circular board and wound around nails, recreate your original image in a distinctively artistic style.

This program takes your source image and applies a configurable algorithm that calculates the best thread paths to recreate your image. The output is a JPEG visual representation of the paths and a text file containing explicit instructions (lists of starting and ending nails for each thread).

## Features

- Configurable number of nails. The nails define the possible paths for the threads in the circular board.
- Adjustable image size. Image size will directly affect the level of detail in your thread art.
- Maximum number of lines. Limit the complexity of your thread art by setting a maximum number of thread paths.
- Random starting nails. Each piece of art can be unique by using a different starting point.
- Control brightness factor and minimum difference to adjust the level of detail.

## Usage

To start using Thread Art Generator, run the main.go file. This will initialize a new thread generator and start the conversion process using the provided parameters. The source image will be taken from the specified "ImageName", and the output will be saved in the defined output folder.

## Example

```cli
go run main.go
```

The above command will execute the primary function in the program, and you should adjust the parameters according to your preferences.

## Results

After running the Thread Art Generator, you will receive two types of output:

1. An image file which represents the paths that threads will have to follow to create the artwork.

2. A text file containing a list of pairs of starting and ending nails for each thread. You can use this file as direct instructions to create the artwork in real life manually.

## Hardware

This project includes not only the software that designs the Thread Art but also the printable and millable parts of the machine that physically creates the thread art. The schematics and designs of these parts are sourced from the StringArt project by [Bdring](https://github.com/bdring/StringArt) and are part of this project to facilitate its recreation.

## FluidNC Configuration

Furthermore, the project conveniently packs the configuration file for FluidNC, a high-performance Grbl CNC firmware specially designed for ESP32 microcontrollers. You can find more about FluidNC on their [official GitHub page](https://github.com/bdring/FluidNC).

## Work In Progress

Thread Art Generator is a project under active development. Some exciting features are currently being worked on:

### GCode Generator

An essential upcoming feature is the Gcode generator for thread path. With this feature, users can easily convert the generated thread path into Gcode, in order to creating it with the machine.

### More Customisation Settings

Further customization options are being added to allow users to tailor the specifics of the project to their artistic vision. These will include tweaking nuances of the thread paths, adjusting the complexity level of the thread art.

### Web UI

I'm working on a Web UI that would not only make the project more accessible to non-technical users but also provide a more interactive way to visualize and customize the thread art generation process.

## Web Service

### Development

#### Database

Connect to the database using the following command:

```cli
docker-compose exec db psql local -U local -d local
```
