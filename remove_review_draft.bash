#!/bin/bash
#
if [ "$1" = "" ]; then
	echo "Missing RDM_ID"
	exit 1
fi
RDM_ID="$1"
. caltechauthors.env
if rdmutil review_request "${RDM_ID}" decline; then
	echo "${RDM_ID} removed from review"
else
	echo "Could not remove ${RDM_ID} from review"
	exit 1
fi
if rdmutil discard_draft "${RDM_ID}"; then
	echo "${RDM_ID} draft discarded"
else
	echo "Could not discard draft ${RDM_ID}"
	exit 1
fi
