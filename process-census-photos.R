#!/usr/bin/env Rscript

# Process the images which have been taken by creating copies which have been
# rotated, deskewed, and trimmed.

library(purrr)
library(readr)
library(fs)
library(processx)
library(progress)

original_dir <- "/media/data/relcensus-original"
cropped_dir  <- "/media/data/relcensus-cropped"

paths_in <- dir_ls(original_dir, recursive = TRUE, glob = "*.JPG")
paths_out <- paths_in %>% path_rel(original_dir) %>% path_abs(cropped_dir)

# Make sure that the paths for the out files exist
paths_out %>% path_dir() %>% unique() %>% dir_create()

# Create a progress bar
pb <- progress_bar$new(total = length(paths_in),
                       format = "processing [:bar] :percent | :current of :total | :eta left",
)

# Process each image by calling an external script. We shell out here because we
# can do more complex things with imagemagick with a command line script. We
# need to be able to not change the original images colors, but to figure out
# the right crop box, it is helpful to highly process the image. This is slow,
# but more effective than using the magick R package.
process_img <- function(in_file, out_file) {
  # Silently return success for images that already have been processed
  if (file_exists(out_file)) return(0L)
  status <- run("./trim-census-photo.sh", c(in_file, file_temp(ext = ".JPG"), out_file))
  if (status$status != 0)
    warning("Failed: ", in_file)
  pb$tick()
  return(status$status)
}

# Run the batch
status <- map2_int(paths_in, paths_out, process_img)
names(status) <- paths_in

# Check for failures, write them to disk, and notify that the job is done
failures <- status %>% keep(~ . != 0)
if (length(failures) > 0) write_lines(names(failures), "failures.txt")
RPushbullet::pbPost(title = "MARE image processing done")
