#!/bin/bash
#

if [ "$1" = "" ]; then
	echo "Missing the eprint repository hostname."
	exit 1
fi
if [ "$2" = "" ]; then
	echo "Missing the dataset collection name"
	exit 1
fi
if [ "$3" = "" ]; then
	echo "Missing the CSV filename for the resource map"
	exit 1
fi
if [ "$4" = "" ]; then
    echo "Missing the CSV filename for the contributor map"
    exit 1
fi

REPO="$1"
C_NAME="$2"
RESOURCE_MAP="$3"
CONTRIBUTOR_MAP="$4"

if [ ! -d "${C_NAME}" ]; then
	dataset init "${C_NAME}"
fi
if [ ! -f eprintids.txt ]; then
	./bin/eprint2rdm -all-ids "${REPO}" >eprintids.txt
fi
while read -r EPRINTID; do
	if [ "${EPRINTID}" != "" ]; then
	    if ./bin/eprint2rdm \
					-resource-map "${RESOURCE_MAP}" \
					-contributor-map "${CONTRIBUTOR_MAP}" \
	        "${REPO}" "${EPRINTID}" \
	        >record.json; then
	        echo "fetched ${EPRINTID} as record.json"
	    else
	        echo "Something went wrong exporting ${EPRINTID}, stopping"
	        exit 1
	    fi
	    if [ -f record.json ]; then
			HAS_KEY="$(dataset haskey "${C_NAME}" "${EPRINTID}")"
			if [ "${HAS_KEY}" = "true" ]; then
	        	echo "Updating ${EPRINTID} to ${C_NAME}"
	        	dataset update -i record.json "${C_NAME}" "${EPRINTID}"
			else
	        	echo "Adding ${EPRINTID} to ${C_NAME}"
	        	dataset create -i record.json "${C_NAME}" "${EPRINTID}"
			fi
	        rm record.json
	    else
	        echo "Something went wrong, could not read record.json for ${EPRINTID}, stopping"
	        exit 1
	    fi
		# NOTE: Harvesting can overwhelm EPrints's REST API. Including
		# a sleep for a few seconds between calls to be polite.
		#sleep 2
	fi
done <eprintids.txt
