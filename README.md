# process-photos

This program processes the photos of the schedules from the 1926 Census of Religious Bodies taken for the [American Religious Ecologies](http://religiousecologies.org) project.

## Useage

The program can take two different kinds of inputs. The most common case is that you have a directory with a batch of images, which will be processed into a different directory. Assuming that this program is stored at the root of directory with all the images, basic usage will look like this. Notice that you will first have to create the directory where you want to store the processed images.

```
mkdir -p 03-for-import/2019-06-12/box013
./process-photos 02-original/2019-06-12/box013 --out 03-for-import/2019-06-12/box013
```

If you want pass in a subset of the images in the directory, perhaps for testing, you can pass in a list or glob of images instead.

```
./process-photos 02-original/2019-06-12/box013/*.JPG --out 03-for-import/2019-06-12/box013
```

Note that the program will silently skip any images that already exist. You can use the second kind of input above to run a small batch first to make sure the settings are right, but it is up to you to delete the images that were output if they were not processed correctly.

You will need to set the options based on what the original images look like in order to process them correctly. The key options are what direction the images should be rotated, how much of them can be obviously cropped before auto-cropping takes over, and what color the background is. For instance, this is a more typical example:

```
./process-photos -r cw --background black --crop-width=0.2 --crop-height=0.2 \
    02-original/2019-06-12/box013 --out 03-for-import/2019-06-12/box013
```

See the built-in help for all the options:

```
./process-photos --help
```

## Compiling and testing

In general, this program should already be available to you on the server. But if you need to build it, you can clone this repository and run `make build`. Note that because this program uses the CGO bindings in order to build [Go Imagick](https://github.com/gographics/imagick) package, some environment variables need to be set. The Makefile takes care of this.

Note that this program assumes an ImageMagick in the 6.x version range is available, both for compiling and available as the `convert` command in your path. You can get this version of ImageMagick on Mac with `brew install imagemagick@6`. On Linux that is the version available in most package repositories.

You can test the program on the included images by running `make test`.

## License and acknowledgements

This software is available under the MIT License. See `LICENSE.md`.

Work on [American Religious Ecologies](http://religiousecologies.org) has been made possible in part by generous funding from the [National Endowment for the Humanities](https://www.neh.gov).
