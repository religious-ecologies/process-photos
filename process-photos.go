package main

import (
	"errors"
	"fmt"
	flag "github.com/spf13/pflag"
	"gopkg.in/gographics/imagick.v3/imagick"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

// Settings that will be used for each image, but which will be the same for each batch
var rotate float64
var background string
var croph, cropw float64
var inDir, outDir string
var jobs int
var wg sync.WaitGroup

// Define colors and other constants
const purple = "srgb(146, 147, 199)"
const trueblack = "srgb(0, 0, 0)"
const black = "srgb(30, 30, 37)" // The black background isn't truly black
const padding = 20               // How much of a border to leave around the schedule
const extension = ".JPG"         // What kind of files are we processing?
const buffersize = 10000         // How many images will we put in the queue? Should be basically all.

func init() {

	rotateFlag := flag.StringP("rotate", "r", "ccw", "Rotate which direction? ccw or cw")
	backgroundFlag := flag.StringP("background", "b", "black", "What color is the background? black or purple")
	flag.Float64VarP(&croph, "crop-height", "h", 0.1, "Percentage of image to crop from the top and the bottom (0.0 to 1.0)")
	flag.Float64VarP(&cropw, "crop-width", "w", 0.1, "Percentage of image to crop from the left and the right (0.0 to 1.0)")
	flag.IntVarP(&jobs, "jobs", "j", 4, "How many images should be processed in parallel?")

	flag.ErrHelp = errors.New("")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, "USEAGE: process-photos [OPTIONS] input-dir output-dir\n")
		fmt.Fprint(os.Stderr, "OPTIONS:\n")
		flag.PrintDefaults()
	}

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

	if jobs > runtime.NumCPU() {
		jobs = runtime.NumCPU()
		log.Printf("Be realistic. Setting the number of jobs to %v.\n", jobs)
	}

	// What is the path to the input and output directories?
	if flag.NArg() != 2 {
		log.Fatal("Pass in two arguments with the path to the image to process.")
	}
	inDir = flag.Args()[0]
	outDir = flag.Args()[1]
	if _, err := os.Stat(inDir); os.IsNotExist(err) {
		log.Println("The directory you passed in does not exist.")
		log.Fatal(err)
	}

}

func main() {

	//TODO Make the output directory if necessary

	// The queue channel will hold all of the images we need to process
	queue := make(chan string, buffersize)
	// Failures will be kept in a channel as well and printed at the end.
	failures := make(chan string)

	// Start the workers which will begin to pull jobs off the channel
	for w := 1; w <= jobs; w++ {
		wg.Add(1)
		go worker(queue, failures)
	}

	// Walk the input directory and add each file found to the queue
	filepath.Walk(inDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && path != inDir {
			log.Printf("Skipping unexpected subdirectory %s.\n", path)
		}
		if filepath.Ext(path) == extension {
			fmt.Println("Adding to queue", path)
			queue <- path
			return nil
		}
		return nil

	})

	close(queue)
	wg.Wait()

	imagick.Initialize()
	defer imagick.Terminate()
	// processImg(imagePath)
}
