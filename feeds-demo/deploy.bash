#!/bin/bash
#

#
# This script deploys an empty feeds demo to the directory indicated. Directory
# must already exist and have the right permissions.
#
function deploy() {
	TARGET="$1"
}

#
# Main 
#
if [ "$1" = "" ]; then
	APP_NAME=$(basename "$0")
	echo "usage: ${APP_NAME} TARGET_DIRECTORY"
	exit 1
fi
TARGET_DIR="${1%/}"

if [ -d "$TARGET_DIR" ]; then
	echo "Deploying to $TARGET_DIR"
else
	echo "$TARGET_DIR does not exist, aborting"
	exit 1
fi

# Find the files to copy to target root
ls -1 *.py *.sql *.bash *.md README.md *.md |\
while read -r FNAME; do
	cp $FNAME $TARGET_DIR/
done
cp -vfR static $TARGET_DIR/
cp -vfR templates $TARGET_DIR/
