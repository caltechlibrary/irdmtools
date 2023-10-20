#!/bin/bash
#

# Clean the htdocs directories
echo "Clearing htdocs/*"
time rm -R htdocs/*

#
# Run the build sequence from harvesting to dataset collections to site generation.
#
if [ "$1" = "full" ]; then
	echo "Reload datasets"
	time ./make_datasets.bash full
else
	echo "Updating datasets"
	time ./make_datasets.bash
fi
echo "Building site"
time ./make_site.bash 
echo "Publishing to S3"
time bash publish.bash
