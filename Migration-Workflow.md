
# Migration Workflow

This is a working document for how Caltech Library is migrating from EPrints to Invenio RDM using irdmtools for our CaltechAUTHORS repository.

EPrints is running version 3.3. RDM is running version 11.

## Requirements

1. A POSIX compatible shell (e.g. Bash)
2. Python 3 and packages in requirments.txt
3. The lastest irdmtools
4. Latest jq for working with JSON and extracting interesting bits

To install the depenedent packages found in the requirements.txt
file you can use the following command.

~~~
python3 -m pip install -r requirements.txt
~~~

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

You can also generate eprint id lists via using MySQL client directory.
See [get_eprintids_by_year.bash].

~~~
#!/bin/bash
#
# NOTE: REPO_ID is imported from the environment.
#
YEAR="$1"
mysql --batch --skip-column-names \
  --execute "SELECT eprintid FROM eprint WHERE date_year = '$YEAR' AND eprint_status = 'archive' ORDER BY date_year, date_month, date_day, eprintid" "${REPO_ID}"
~~~

## Migrating records

We have over 100,000 records so migration is going to take a couple days. We're migrating the metadata and any publicly visible files. I've found working by year to be helpful way of batching up record loads.

For a given set of eprintid in a file called "migrate-ids.txt" you can use
the `eprints_to_rdm.py` script along with an environment setup to
automatically migrate both metadata and files from CaltechAUTHORS to the
new RDM deployment.

~~~
. caltechauthors.env
./eprints_to_rdm.py migrate-ids.txt
~~~

The migration tool with stop on error. This is deliberate. When it stops you need to investigate the error and either manually migrate the record or take other mediation actions.

