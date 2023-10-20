#!/bin/bash
#

# Clean the htdocs directories
echo "Clearing htdocs/*"
time rm -R htdocs/*

#
# Run the build sequence from harvesting to dataset collections to site generation.
#
echo "Updating datasets"
time ./make_datasets.bash
echo "Building site"
time ./make_site.bash

