#!/bin/bash
#

#
# Migrate a single record using eprint2rdm, ./migrate_record.py and rdmutil.
#
CFG_FILE="$1"
EPRINT_ID="$2"
if [ "$CFG_FILE" = "" ]; then
	echo "Missing configuration filename"
	exit 1
fi
if [ "$EPRINT_ID" = "" ]; then
	echo "Missing eprint id"
	exit 1
fi
if [ ! -f "${CFG_FILE}" ]; then
	echo "Cannot find confiugration file ${CFG_FILE}"
	exit 1
fi

#
# Source our configuration
#
# shellcheck disable=SC1090
source "${CFG_FILE}"

#
# Make sure we have enough environment after sourcing configuration file.
#
if [ "$EPRINT_HOST" = "" ]; then
	echo "EPRINT_HOST must be set and exported in $CFG_FILE"
	exit 1
fi
if [ "$RDM_URL" = "" ]; then
	echo "RDM_URL must be set and exported in $CFG_FILE"
	exit 1
fi
if [ "$RDMTOK" = "" ]; then
	echo "RDMTOK must be set and exported in $CFG_FILE"
	exit 1
fi
if [ "$RDM_COMMUNITY_ID" = "" ]; then
	echo "RDM_COMMUNITY_ID must be set and exported in $CFG_FILE"
	exit 1
fi

function migrate_record() {
	EPRINT_HOST="$1"
	EPRINT_ID="$2"
	eprint2rdm "${EPRINT_HOST}" "${EPRINT_ID}" >"${EPRINT_ID}.json"
	RDM_RECORD_ID=$(cat "${EPRINT_ID}.json" | ./migrate_record.py | rdmutil new_record | jq -r .id)
	if [ "${RDM_RECORD_ID}" = "" ]; then
		echo "Failed to return RDM_RECORD_ID from pipline"
		exit 1
	fi
	rdmutil get_draft "${RDM_RECORD_ID}" >"${RDM_RECORD_ID}.json"
	HAS_FILES=$(jq .files.enable "${RDM_RECORD_ID}.json")
	if [ "$HAS_FILES" = "true" ]; then
	    # Attach the public files. FIXME: Need to iterate over the map to files, convert each URL into a file path or curl them down.
	fi
	echo "${RDM_RECORD_ID}"
}

function send_to_community() {
	RECORD_ID="$1"
	COMMUNITY_ID="$2"
	rdmutil -debug send_to_community "${RECORD_ID}" "${COMMUNITY_ID}"
}

echo "Using configuration $CFG_FILE"
##RDM_RECORD_ID=$(migrate_record "$EPRINT_HOST" "${EPRINT_ID}")
RDM_RECORD_ID="wdjkt-cv540"
echo "RDM recorid is ${RDM_RECORD_ID}"
# FIXME: Set the community
send_to_community "${RDM_RECORD_ID}" "$RDM_COMMUNITY_ID"


# FIXME: Need to see if we need to migrate public files.
#        If so set files to enable, then add the public files.
# FIXME: Set access appropriately
# FIXME: Submit draft for review
# FIXME: Accept reviewed draft
