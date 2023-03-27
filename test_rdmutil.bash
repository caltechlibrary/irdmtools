#!/bin/bash

#
# Run some command line tests to confirm working cli
#

if [[ -d "testout" ]]; then
    rm -fR testout
fi
mkdir -p testout

if [[ ! -f authors-test.rc ]]; then
    echo "Missing authors-test.rc for testing rdmutil with authors.caltechlibrary.dev"
    exit 1
fi

#
# Basic get keys and harvest
#
echo 'Testing Basics, Get All ids, Harvesting'
. authors-test.rc
# FIXME: Make sure dataset collection exists
C_NAME="$TEST_C_NAME"
if [[ "$C_NAME" = "" ]]; then
    echo "TEST_C_NAME not set, aborting"
    exit 1
fi
if [[ -d "$C_NAME" ]]; then
    echo "Removing stale collection, $C_NAME"
    rm -fR "$C_NAME"
fi
echo "Creating $C_NAME for test"
dataset init "$C_NAME" "sqlite://collection.sqlite"
if ! time ./bin/rdmutil -debug get_all_ids >testout/author_test_ids.json; then
    echo "Failed to complete get_all_ids"
    exit 1
fi
if ! time ./bin/rdmutil -debug harvest; then
    ceho "Failed to complete harvest"
    exit 1
fi
echo "OK, Tests completed."
