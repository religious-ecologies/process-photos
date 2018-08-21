#!/usr/bin/env bash

# This script takes a path to an input file, modifies it, and creates an outfile
# that has been cropped and deskewed with imagemagick

infile=$1
tempfile=$2
outfile=$3

convert $infile -background "#000000" -rotate "90" -deskew 40% $tempfile
convert $tempfile -crop `convert $tempfile -virtual-pixel edge -blur 0x30 -fuzz 30% -trim -format '%wx%h%O' info:` +repage $outfile
rm -f $tempfile
