#!/bin/bash

#
# Run some command line tests to confirm working cli
#

if [[ -d "testout" ]]; then
    rm -fR testout
fi
mkdir -p testout

#
# Basic get keys and harvest
#
echo 'Testing doi2rdm basics'
. authors-test.rc
for DOI in "10.1063/5.0122760" "10.1029/2022gl101441" "10.1038/s41583-022-00670-w"; do
	FNAME=$(basename "$DOI")
	if ! ./bin/doi2rdm "$DOI" >"testout/${FNAME}.json"; then
		echo "doi2rdm failed for ${DOI}, aborting"
		exit 1
	fi
done
echo "OK, Tests completed."
