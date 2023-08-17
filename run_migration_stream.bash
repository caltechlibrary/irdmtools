#!/bin/bash

if [ "$1" = "" ]; then
	echo "Missing REPO ID and/or EPRINT_ID list"
	exit 1
fi
if [ "$2" = "" ]; then
	echo "Missing EPRINT_ID list"
	exit 1
fi
#
# run migration stream takes a list of EPrint ids, makes a working
# copy and sets up logging for eprints_to_rdm.py.
# If the working list exists it is used for the next stream.
#
# E.g.
#
# run_migratraion_stream 1890s-eprints.txt
# 
# Results in working stream and log files
#    1890s-eprints.txt.working
#    1890s-eprints.txt.log
#
REPO_ID="$1"
EPRINT_ID_LIST="$2"
WORKING_LIST="${EPRINT_ID_LIST}.working"
WORKING_LOG="${EPRINT_ID_LIST}.log"

if [ -f "${REPO_ID}.env" ]; then
	echo "Sourcing ${REPO_ID}.env"
	. "${REPO_ID}.env"
else
	echo "Missing environment file ${REPO_ID}.env"
	exit 1
fi

if [ ! -f "${WORKING_LIST}" ]; then
	echo "Making working copy of ${EPRINT_ID_LIST}"
	cp -v "${EPRINT_ID_LIST}" "${WORKING_LIST}"
else
	echo "Using existing ${WORKING_LIST}"
fi
cat <<EOT

	environment used ${REPO_ID}.env
	Working list ${WORKING_LIST}
	Logging to ${WORKING_LOG}

EOT
touch "${WORKING_LOG}"
./eprints_to_rdm.py "${WORKING_LIST}" >"${WORKING_LOG}"

