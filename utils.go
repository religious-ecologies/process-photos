package main

import (
	"gopkg.in/cheggaaa/pb.v1"
	"log"
	"path/filepath"
	"strconv"
	"strings"
)

// A worker pulls jobs off the queue and calls the processing functions for each
func worker(queue <-chan string, failures chan<- string, bar *pb.ProgressBar) {
	defer wg.Done()
	for img := range queue {
		err := processImg(img)
		if err != nil {
			log.Println(err)
			failures <- img
		}
		bar.Increment()
	}
}

// The input is in the format "2109 2743 +312 +274"
func parseTrim(input string) (width, height, x, y int) {
	params := strings.Fields(input)
	width, err := strconv.Atoi(params[0])
	if err != nil {
		log.Println("Failed to parse string: ", err)
	}
	height, err = strconv.Atoi(params[1])
	if err != nil {
		log.Println("Failed to parse string: ", err)
	}
	x, err = strconv.Atoi(params[2])
	if err != nil {
		log.Println("Failed to parse string: ", err)
	}
	y, err = strconv.Atoi(params[3])
	if err != nil {
		log.Println("Failed to parse string: ", err)
	}
	return
}

// Given an input file path, what should the output file path be?
func getOutPath(in string) string {
	return filepath.Join(outDir, filepath.Base(in))
}
