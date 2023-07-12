#!/bin/bash

#
# Run some command line tests to confirm working cli
#

RC_FILE="test_doi2rdm.env"

if ! test -d "testout"; then
	mkdir -p testout
fi

#
# Basic get keys and harvest
#
. test_doi2rdm.env
echo "Testing doi2rdm basics using $RC_FILE"
for DOI in "10.1063/5.0122760" "10.1029/2022gl101441" "10.1038/s41583-022-00670-w"; do
	FNAME=$(basename "$DOI")
	if ! ./bin/doi2rdm "$DOI" >"testout/${FNAME}.json"; then
		echo "doi2rdm failed for ${DOI}, aborting"
		exit 1
	fi
done
echo "OK, Tests completed."
