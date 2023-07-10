#!/bin/bash

#
# Run some command line tests to confirm working cli
#

RC_FILE="test_eprint2rdm.env"

if ! test -d "testout"; then
	mkdir -p testout
fi

if [[ ! -f "$RC_FILE" ]]; then
    echo "Missing $RC_FILE for testing eprint2rdm with authors.caltechlibrary.dev"
    exit 1
fi

#
# Basic get keys and harvest
#
. test_eprint2rdm.env
if [[ "$C_NAME" = "" ]]; then
    echo "C_NAME not set, aborting"
    exit 1
fi
if [[ "$EPRINT_HOST" = "" ]]; then
	echo "EPRINT_HOST not set, aborting"
	exit 1
fi
if [[ "$EPRINT_USER" = "" ]]; then
	echo "EPRINT_USER not set, aborting"
fi
if [[ "$EPRINT_PASSWORD" = "" ]]; then
	echo "EPRINT_PASSWORD not set, aborting"
fi

echo "Testing basics using $RC_FILE, get all ids, harvesting to a dataset"
if [[ -d "$C_NAME" ]]; then
    echo "Using existing collection, $C_NAME"
else 
    echo "Creating $C_NAME for test"
    dataset init "$C_NAME" "sqlite://collection.sqlite"
fi
if test -f testout/eprintids.txt; then
	echo "Using testout/eprintids.txt"
else
	if ! time ./bin/eprint2rdm -all-ids "$EPRINT_HOST" >testout/eprintids.txt; then
    	echo "Failed to complete eprint2rdm -all-ids $EPRINT_HOST"
    	exit 1
	fi
fi
if ! time ./bin/eprint2rdm -id-list testout/eprintids.txt -harvest "$C_NAME" "$EPRINT_HOST"; then
    echo "Failed to complete eprint2rdm -id-list testout/eprintids.txt -harvest $C_NAME $EPRINT_HOST"
    exit 1
fi
echo "OK, Tests completed."
