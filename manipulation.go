package main

import (
	"gopkg.in/gographics/imagick.v2/imagick"
	"io/ioutil"
	"os"
	"os/exec"
)

func processImg(inPath string) error {

	outPath := getOutPath(inPath)

	// Check if the out file does not exist. If it does already exist, then just
	// skip it.
	if _, err := os.Stat(outPath); !os.IsNotExist(err) {
		return nil
	}

	// Allocate memory for the image and read it in
	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	err := mw.ReadImage(inPath)
	if err != nil {
		return err
	}

	// The current background color in this image
	bg := imagick.NewPixelWand()
	defer bg.Destroy()
	bg.SetColor(background)

	err = mw.SetBackgroundColor(bg)
	err = mw.SetImageBackgroundColor(bg)
	if err != nil {
		return err
	}

	// Rotate the image so it is portrait orientation
	err = mw.RotateImage(bg, rotate)
	if err != nil {
		return err
	}

	// Lop off as much of the image as can safely be specified as a percentage.
	// This helps ensure that we are not getting asking the autocrop to do too much,
	// especially since the corners distant from the schedule tend to be darker and
	// have a gradient-like appearance.
	w := float64(mw.GetImageWidth())
	h := float64(mw.GetImageHeight())
	xStart := w * cropw
	yStart := h * croph
	newWidth := w - (w * cropw) - (w * cropw)
	newHeight := h - (h * croph) - (h * croph)
	err = mw.CropImage(uint(newWidth), uint(newHeight), int(xStart), int(yStart))
	if err != nil {
		return err
	}

	// Put a border around the image using the average color of the background.
	// This is necessary because the color of the corner pixels will determine
	// which color is the background and this avoids the possibility of the color
	// being close but not exact.
	err = mw.BorderImage(bg, 100, 100)
	if err != nil {
		return err
	}

	// Try to make the image appear level
	err = mw.DeskewImage(10000)
	if err != nil {
		return err
	}

	// Manually repage the image after the crop and border. This is necessary so
	// that the dimensions for our second crop are the same as what we are going
	// to read from command-line imagemagick.
	err = mw.SetImagePage(0, 0, 0, 0)
	if err != nil {
		return err
	}

	// Next we need to trim the image, meaning autocrop it so that as much of the
	// background color is removed as possible. We want to know the dimensions to
	// autocrop, not to actually autocrop right away, for two reasons. First we
	// want to leave a little border around so it doesn't look like we have
	// cropped the actual schedule. Second, we need to blur the image before
	// autocroping, but we can't do that with the imagick API. Long way of saying,
	// we need to create a tempfile, and shell out to command line ImageMagick to
	// get to the dimensions of the best crop.
	tempf, err := ioutil.TempFile("", "relecprocessing")
	if err != nil {
		return err
	}
	tempFilename := tempf.Name()
	err = mw.WriteImageFile(tempf)
	defer tempf.Close()
	if err != nil {
		return err
	}

	// This command blurs the image to deal with minor imperfections and
	// variations in color, then returns the dimensions of the best trim.
	cmd := "convert " + tempFilename + " -virtual-pixel edge -blur 0x30 -fuzz 12% -trim -format '%w %h %X %Y' info:"
	trim := exec.Command("bash", "-c", cmd)

	// Capture the output from ImageMagick
	trimOut, err := trim.Output()
	if err != nil {
		return err
	}

	err = os.Remove(tempFilename)
	if err != nil {
		// Keep going if the only problem is that we couldn't delete the tempfile
	}

	// Turn those dimensions into something with can work with
	width, height, x, y := parseTrim(string(trimOut))

	// Crop the image, putting the padding all around
	err = mw.CropImage(uint(width+2*padding), uint(height+2*padding), x-padding, y-padding)
	if err != nil {
		return err
	}

	// Write the processed image out to the correct location
	err = mw.WriteImage(outPath)
	if err != nil {
		return err
	}

	return nil

}
