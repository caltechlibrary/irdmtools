#!/bin/bash
#

#
# Run the build sequence from harvesting to dataset collections to site generation.
#
time ./make_datasets.bash
time ./make_site.bash

