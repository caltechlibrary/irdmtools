#!/bin/bash

if [ "$RDMTOK" = "" ]; then
	echo "Missing RDMTOK from the envinronment"
	exit 1
fi

#
# Using RDMTOK from environment to access /api/users
#
curl -H "Authorization: Bearer $RDMTOK" https://authors.caltechlibrary.dev/api/users







