
# Migration Workflow

This is a working document for how Caltech Library is migrating from EPrints to Invenio RDM using irdmtools for our CaltechAUTHORS repository.

EPrints is running version 3.3. RDM is running version 11.

## Requirements

1. A POSIX compatible shell (e.g. Bash)
2. The lastest irdmtools
3. jq for working with JSON and extracting interesting bits
4. Unix split command to split files into smaller lists

## Setup

You need to configure your environment correctly for this to work. Here is an example environment file based on ours for CaltechAUTHORS.

~~~sh
#!/bin/sh
#
# Setup for caltechauthors
# This will be sourced from the environment by 
#
REPO_ID="<REPO_ID>"
EPRINT_HOST="<EPRINT_HOSTNAME>"
EPRINT_USER="<EPRINT_REST_USERNAME>"
EPRINT_PASSWORD="EPRITN_REST_PASSWORD>"

# Dataset collection setup
DB_USER="<POSTGRES_USERNAME>"
DB_PASSWORD="<POSTGRES_PASSWORD>"

#
# Invenio-RDM access setup
#
RDM_URL="<URL_TO_RDM_REPOSITORY>"
RDMTOK="<RDM_TOKEN_FOR_USER_ACCOUNT_USED_TO_MIGRATE>"
# RDM_COMMUNITY_ID should be the default community you are migrating
# content to.
RDM_COMMUNITY_ID="<RDM_COMMUNITY_ID>"

export REPO_ID
export EPRINT_HOST
export EPRINT_USER
export EPRINT_PASSWORD
export DB_USER
export DB_PASSWORD
export RDM_URL
export RDMTOK

#
# Setup psql environment
#
export PSQL_EDITOR="vi" # "/Users/rsdoiel/bin/micro"
~~~

I usually source this at the beginning of my working session.

~~~sh
. caltechauthors.env
~~~

## Getting a list of ids to migrate

At this stage of our migration project we can support all the record types in RDM we want to migrate from EPrints. As a result we can migrate all the EPrint ids remaining in CaltechAUTHORS. You can generate a list of record ids using eprint2rdm and the option `-all-ids`

~~~sh
eprint2rdm -all-ids $EPRINT_HOST >eprint-ids.txt
~~~

## Migrating records

We have over 100,000 records so migration is going to take a couple days. We're migrating the metadata and any publicly visible files.

I've found it convient to split up the eprint-ids.txt file into a series of smaller files and migrate them in smaller batches. This POSIX split command does this nicely.

~~~sh
split -l 1000 eprint-ids.txt eprints-ids.
~~~

This will result in files with an extensions of ".aa" through ".zz" for the file you split.

You can use a simple bash script to iterate over a file and process the EPrints record and file information to migrate from one system to another.  Here are the steps to migrate a single record. Below are the steps taken to migrate each record.

1. Get an EPrint record id from our list (E.g. it is numeric) and fetch the JSON version of the record using eprint2rdm
2. Create a new RDM draft using migrate_record.py and rdmutil
3. Use rdmutil to create a new RDM draft record; capture the assigned RDM record id by piping the result through `jq -r .id`.
4. Evaludate the EPrint JSON to see if we have files to attach
5. If we have files, using rdmutil to upload them to the draft RDM record
6. Send the draft RDM record to the community using the environment RDM_COMMUNITY_ID
7. Accept the draft record in the community
8. Cleanup temportary files and proceed to next record

Example of processing, assumes eprint id was assigned to the environment variable EP_PRINT

~~~sh
eprint2rdm $EPRINT_HOST $EPRINT_ID >"${EPRINT_ID}.json"
RDM_RECORD_ID=$(cat "${EPRINT_ID}.json" | ./migrate_record.py | rdmutil new_record | jq -r .id)
rdmutil get_draft "${RDM_RECORD_ID}" >"${RDM_RECORD_ID}.json"
HAS_FILES=$(jq .files.enable "${RDM_RECORD_ID})
if [ "$HAS_FILES" = "true" ]; then
    # Attach the files listing in ${EPRINT_ID}.json
    # FIXME: need to implement scripted implemented
fi

~~~
