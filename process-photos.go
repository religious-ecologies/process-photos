package main

import (
	flag "github.com/spf13/pflag"
	"gopkg.in/gographics/imagick.v3/imagick"
	"log"
	"os"
)

// Settings that will be used for each image, but which will be the same for each batch
var rotate float64
var background string
var croph, cropw float64
var imagePath string

// Define colors and other constants
const purple = "srgb(146, 147, 199)"
const black = "srgb(30, 30, 37)" // The black background isn't truly black
const trueblack = "srgb(0, 0, 0)"
const padding = 20 // How much of a border to leave around the schedule

func init() {

	rotateFlag := flag.StringP("rotate", "r", "ccw", "Rotate which direction? ccw or cw")
	backgroundFlag := flag.StringP("background", "b", "black", "What color is the background? black or purple")
	flag.Float64VarP(&croph, "crop-height", "h", 0.1, "Percentage of image to crop from the top and the bottom (0.0 to 1.0)")
	flag.Float64VarP(&cropw, "crop-width", "w", 0.1, "Percentage of image to crop from the left and the right (0.0 to 1.0)")
	flag.Parse()

	// Which direction should the images be rotated?
	switch *rotateFlag {
	case "ccw":
		rotate = 270
	case "cw":
		rotate = 90
	default:
		log.Fatal("Rotate must be set to either ccw or cw.")
	}

	// What color is the background?
	switch *backgroundFlag {
	case "black":
		background = black
	case "purple":
		background = purple
	default:
		log.Fatal("Background color must be set to black or purple.")
	}

	// What is the path to the image?
	if flag.NArg() != 1 {
		log.Fatal("Pass in one argument with the path to the image to process.")
	}
	imagePath = flag.Args()[0]
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		log.Println("The file you passed in does not exist.")
		log.Fatal(err)
	}

}

func main() {
	imagick.Initialize()
	defer imagick.Terminate()
	processImg(imagePath)
}
