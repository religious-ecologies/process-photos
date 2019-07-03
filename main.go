package main

import (
	"errors"
	"fmt"
	flag "github.com/spf13/pflag"
	"gopkg.in/cheggaaa/pb.v1"
	"gopkg.in/gographics/imagick.v2/imagick"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// Settings that will be used for each image, but which will be the same for each batch
var rotate float64
var background string
var croph, cropw float64
var outDir string
var jobs int
var padding int
var wg sync.WaitGroup
var images = make([]string, 0, 1000)
var version string // Date/time of compilation is injected at compile time

// Define colors and other constants
const purple = "srgb(146, 147, 199)"
const gray = "srgb(95, 95, 125)"
const black = "srgb(30, 30, 37)"        // The black background isn't truly black
const extension = ".JPG"                // What kind of files are we processing?
const minForPb = 10                     // How many images do we have to process to show a progress bar?
const mirroredOutPath = "03-for-import" // The directory to mirror to
const originalsDir = "02-original"      // The directory where the originals are stored

func init() {

	rotateFlag := flag.StringP("rotate", "r", "ccw", "Rotate which direction? ccw or cw")
	backgroundFlag := flag.StringP("background", "b", "gray", "What color is the background? gray, black, purple")
	flag.Float64VarP(&croph, "crop-height", "h", 0.1, "Percentage of image to crop from the top and the bottom (0.0 to 1.0)")
	flag.Float64VarP(&cropw, "crop-width", "w", 0.1, "Percentage of image to crop from the left and the right (0.0 to 1.0)")
	flag.IntVarP(&padding, "padding", "p", 30, "How many pixels of extra padding should be added?")
	flag.IntVarP(&jobs, "jobs", "j", 0, "How many images should be processed in parallel? 0 sets an aggressive but sane default for the current system.")
	flag.StringVarP(&outDir, "out", "o", "_", "Where should the processed files be output?")
	mirrorPath := flag.Bool("mirror-path", false, "Mirror the output path instead of specifying the out directory.")
	showVersion := flag.Bool("version", false, "Show version")

	flag.ErrHelp = errors.New("")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, `USEAGE:
	process-photos [OPTIONS] --out=OUTPUTDIR INPUTDIR
	process-photos [OPTIONS] --out=OUTPUTDIR IMAGE-1.JPG IMAGE-2.JPG ...
	process-photos [OPTIONS] --mirror-path INPUTDIR`)
		fmt.Fprint(os.Stderr, "\nOPTIONS:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		os.Exit(0)
	}

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
	case "gray":
		background = gray
	default:
		log.Fatal("Background color must be set to black, purple, or gray.")
	}

	// ImageMagick is parallelized itself, so run fewer jobs than there are cores.
	maxJobs := runtime.NumCPU() / 2
	switch {
	case jobs == 0:
		jobs = runtime.NumCPU() / 2
	case jobs > maxJobs:
		jobs = maxJobs
		log.Printf("Be realistic. Setting the number of jobs to %v.\n", jobs)
	case jobs < 0:
		jobs = 1
	}

	// Are we dealing with a directory, for which we need to find all the files?
	// Or have we been passed in a list of files?
	if flag.NArg() < 1 {
		log.Fatal("Pass in a directory of images to process, or a list (or glob) of images to process.")
	}

	inStat, err := os.Stat(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}

	if inStat.IsDir() {
		// If the first argument is a directory, there should be only one argument.
		if flag.NArg() > 1 {
			log.Fatal("Pass in only a single directory, or pass in a list of image files.")
		}
		// Read the directory and keep only the files with the correct extension
		files, err := ioutil.ReadDir(flag.Arg(0))
		if err != nil {
			log.Fatal(err)
		}
		for _, f := range files {
			if filepath.Ext(f.Name()) == extension {
				filename := filepath.Join(flag.Arg(0), f.Name())
				images = append(images, filename)
			}
		}
	} else {
		// The first argument is not a directory, so assume we have been passed a list
		// of files
		for _, f := range flag.Args() {
			if filepath.Ext(f) == extension {
				images = append(images, f)
			}
		}
	}

	if len(images) < 1 {
		log.Fatalf("Could not find any images with the extension %s.\n", extension)
	}

	// Figure out whether we want to mirror the path or not.
	if outDir == "_" && !*mirrorPath {
		log.Fatal("Please specify an output directory with the --out/-o flag. Or did you mean to use --mirror-path?")
	} else if outDir != "_" && *mirrorPath {
		log.Fatal("Conflicting options. Please specify a directory with --out or use --mirror-path, but not both.")
	} else if outDir == "_" && *mirrorPath {
		// Check the assumption that the directory we will mirror to is in the same
		// directory as this program was run from.
		if _, err := os.Stat(mirroredOutPath); os.IsNotExist(err) {
			log.Fatalf("The directory %s was not found where expected.", mirroredOutPath)
		} else if err != nil {
			log.Fatal(err)
		}
		// Replace the base name of the directory and create it
		outDir = strings.Replace(filepath.Dir(images[0]), originalsDir, mirroredOutPath, 1)
		err = os.MkdirAll(outDir, 0770)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Output directory: %s", outDir)
	}
	outStat, err := os.Stat(outDir)
	if os.IsNotExist(err) {
		log.Fatalf("The output directory %s does not exist.\n", outDir)
	} else if err != nil {
		log.Fatal(err)
	} else if !outStat.IsDir() {
		log.Fatalf("This is not a directory: %s", outDir)
	}

}

func main() {

	// The queue channel will hold all of the images we need to process
	queue := make(chan string, len(images))
	// Failures will be kept in a channel as well and printed at the end.
	failures := make(chan string, len(images))

	bar := pb.New(len(images))
	bar.ShowTimeLeft = true

	// Start the workers which will begin to pull jobs off the channel
	for w := 1; w <= jobs; w++ {
		wg.Add(1)
		go worker(queue, failures, bar)
	}

	imagick.Initialize()
	defer imagick.Terminate()

	if len(images) >= minForPb {
		bar.Start()
	}

	// Add the images to the queue
	for _, i := range images {
		queue <- i
	}
	close(queue)

	wg.Wait()
	close(failures)

	if len(images) >= minForPb {
		bar.Finish()
	}

	if len(failures) > 0 {
		fmt.Println("\nThe following images were not correctly processed:")
		for f := range failures {
			fmt.Println(f)
		}
	}

}
