#!/bin/bash

#
# Run some command line tests to confirm working cli
#

RC_FILE="test_rdmutil.rc"

if ! test -d "testout"; then
	mkdir -p testout
fi

if [[ ! -f $RC_FILE ]]; then
    echo "Missing $RC_FILE for testing rdmutil with authors.caltechlibrary.dev"
    exit 1
fi

#
# Basic get keys and harvest
#

# shellcheck source="test_rdmutil.rc"
. "${RC_FILE}"
if [[ "$C_NAME" = "" ]]; then
    echo "C_NAME not set, aborting"
    exit 1
fi
if [[ -d "$C_NAME" ]]; then
    echo "Using existing collection, $C_NAME"
    #rm -fR "$C_NAME"
fi
if [[ ! -d "$C_NAME" ]]; then
	echo "Creating $C_NAME for test"
	dataset init "$C_NAME" "sqlite://collection.sqlite"
fi
if [[ "$RDM_C_NAME" = "" ]]; then
	echo "RDM_C_NAME not set, aborting"
	exit 1
fi
if [[ "$RDM_INVENIO_API" = "" ]]; then
	echo "RDM_INVENIO_API not set, aborting"
	exit 1
fi
if [[ "$RDM_INVENIO_TOKEN" = "" ]]; then
	echo "RDM_INVENIO_TOKEN not set, aborting"
	exit 1
fi

echo "Testing Basics using $RC_FILE, Get All ids and Harvest"
if test -f testout/rdm_test_ids.json; then
	echo "Using testout/rdm_test_ids.json"
else
	if ! time ./bin/rdmutil get_all_ids >testout/rdm_test_ids.json; then
        echo "Failed to complete get_all_ids"
    	exit 1
	fi
fi 
if ! time ./bin/rdmutil harvest ./testout/rdm_test_ids.json; then
    echo "Failed to complete harvest"
    exit 1
fi
echo "OK, Tests completed."
