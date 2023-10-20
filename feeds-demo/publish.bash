#!/bin/bash

#
# Publish takes the content in the htdocs directory and copies it to the S3 bucket indicated by the environment
# variable BUCKET.
#
BUCKET="s3://feeds-test.library.caltech.edu"
if [ -f "publish.env" ]; then
	echo "Loading environment from publish.env"
	. publish.env
fi
if [ "$BUCKET" = "" ]; then
	echo "No bucket to publish, set BUCKET"
	exit 1
fi
echo "Copying htdocs to $BUCKET"
time s5cmd cp --acl "public-read" htdocs/ "$BUCKET"
