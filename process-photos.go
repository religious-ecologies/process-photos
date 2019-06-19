package main

import (
	"flag"
	"fmt"
	"gopkg.in/gographics/imagick.v3/imagick"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

var rotate float64
var background string
var cropt, cropb, cropl, cropr float64
var imagePath string

func init() {

	rotateFlag := flag.String("rotate", "ccw", "Rotate which direction? ccw or cw")
	backgroundFlag := flag.String("background", "black", "What color is the background? black or purple")
	flag.Float64Var(&cropt, "crop-top", 0.1, "Percentage of image to crop from the top (0.0 to 1.0)")
	flag.Float64Var(&cropb, "crop-bottom", 0.1, "Percentage of image to crop from the bottom (0.0 to 1.0)")
	flag.Float64Var(&cropl, "crop-left", 0.1, "Percentage of image to crop from the left (0.0 to 1.0)")
	flag.Float64Var(&cropr, "crop-right", 0.1, "Percentage of image to crop from the right (0.0 to 1.0)")
	flag.Parse()

	switch *rotateFlag {
	case "ccw":
		rotate = 270
	case "cw":
		rotate = 90
	default:
		log.Fatal("Rotate must be set to either ccw or cw.")
	}

	// What color is the background?
	const purple = "srgb(146,147,199)"
	const black = "srgb(30, 30, 37)"
	switch *backgroundFlag {
	case "black":
		background = black
	case "purple":
		background = purple
	default:
		log.Fatal("Background color must be set to black or purple.")
	}

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

	fmt.Println("Get going.")
	imagick.Initialize()
	defer imagick.Terminate()

	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	// Create a background color
	bg := imagick.NewPixelWand()
	defer bg.Destroy()
	bg.SetColor(background)

	// True black eventually
	black := imagick.NewPixelWand()
	defer black.Destroy()
	black.SetColor("#000000")

	err := mw.ReadImage(imagePath)
	if err != nil {
		log.Fatal(err)
	}

	err = mw.SetBackgroundColor(bg)
	if err != nil {
		log.Fatal(err)
	}
	err = mw.SetImageBackgroundColor(bg)
	if err != nil {
		log.Fatal(err)
	}

	err = mw.RotateImage(bg, rotate)
	if err != nil {
		log.Fatal(err)
	}

	w := float64(mw.GetImageWidth())
	h := float64(mw.GetImageHeight())

	xStart := w * cropl
	yStart := h * cropt
	newWidth := w - (w * cropl) - (w * cropr)
	newHeight := h - (h * cropt) - (h * cropb)

	err = mw.CropImage(uint(newWidth), uint(newHeight), int(xStart), int(yStart))
	if err != nil {
		log.Fatal(err)
	}

	err = mw.BorderImage(bg, 200, 200, imagick.COMPOSITE_OP_COPY)
	if err != nil {
		log.Fatal(err)
	}

	// Manually repage the image after the crop and border
	err = mw.SetImagePage(0, 0, 0, 0)
	if err != nil {
		log.Fatal(err)
	}

	// We need to write to a temporary file
	tempf, err := ioutil.TempFile("", "relecprocessing")
	if err != nil {
		log.Fatal(err)
	}
	tempFilename := tempf.Name()
	defer tempf.Close()

	err = mw.WriteImageFile(tempf)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(tempFilename)

	cmd := "convert " + tempFilename + " -virtual-pixel edge -blur 0x30 -fuzz 15% -trim -format '%w %h %X %Y' info:"
	trim := exec.Command("bash", "-c", cmd)

	trimOut, err := trim.Output()
	if err != nil {
		log.Fatal(err)
	}

	err = os.Remove(tempFilename)
	if err != nil {
		log.Fatal(err)
	}

	width, height, x, y := parseTrim(string(trimOut))

	padding := 20

	err = mw.CropImage(uint(width+2*padding), uint(height+2*padding), x-padding, y-padding)
	if err != nil {
		log.Fatal(err)
	}

	err = mw.FloodfillPaintImage(black, 15000, bg, 0, 0, false)
	if err != nil {
		log.Fatal(err)
	}

	err = mw.SetBackgroundColor(black)
	if err != nil {
		log.Fatal(err)
	}
	err = mw.SetImageBackgroundColor(black)
	if err != nil {
		log.Fatal(err)
	}

	err = mw.DeskewImage(100)
	if err != nil {
		log.Fatal(err)
	}

	err = mw.WriteImage("test/test.jpg")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Git er done.")

}
