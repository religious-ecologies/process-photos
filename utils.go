package main

import (
	"log"
	"strconv"
	"strings"
)

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
