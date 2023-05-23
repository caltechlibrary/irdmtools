#!/bin/bash

if [ "$REPO_ID" = "" ]; then
	echo 'Missing REPO_ID for resetting migration data'
	exit 1
fi

if [ ! -f "${REPO_ID}.env" ]; then
	echo "Can't read or find ${REPO_ID}.env"
	exit 1
else
	echo "Loading environent from ${REPO_ID}.env"
	# shellcheck disable=SC1090
	.  "${REPO_ID}.env"
fi

if [ "${C_NAME}" = "" ] || [ ! -d "${C_NAME}" ]; then
	echo "Can't find dataset collection ${C_NAME}"
	exit 1
else
	echo "There are $(dataset count "${C_NAME}") records in ${C_NAME}"
	echo "Deleting records from $C_NAME, this takes a while"
	dataset keys "$C_NAME" >deleting-ids.txt
	echo "" >>deleting-ids.txt
	while read -r KEY; do
		dataset delete "${C_NAME}" "${KEY}"
	done <deleting-ids.txt
	rm deleting-ids.txt
fi

