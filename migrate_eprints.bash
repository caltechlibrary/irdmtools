#!/bin/bash

#
# Migration reads a replica of the EPrints repository database
# from the localhost (this could be changed to point at production)
# and updates a local dataset collection with records harvested using
# eprint2rdm.
#
function usage() {
	APP_NAME=$(basename "$0")
	cat <<EOT
---
title: "${APP_NAME} (1) user manual"
pubDate: 2023-02-28
author: "R. S. Doiel"
---

# NAME

${APP_NAME}

# SYNOPSIS

${APP_NAME} [OPTIONS] [setup|export|import] [keys|key_list|full] 

# DESCRIPTION

${APP_NAME} can harvest records from a EPrints repository based on
settings in the environment (e.g. REPO_ID, EPRINT_USER, EPRINT_PASSWORD,
C_NAME).  ${APP_NAME} will generate datasets for each record
EPrints repository harvested. It can also import the harvested records
into Invenio RDM.

If you are missing configuration ${APP_NAME} will prompt and create
the necessary configuration based on the repository name you're
harvesting (e.g. caltechauthors configuration file would be
'caltechauthors.env').

If "full" is passed as a parameter it will harvest EPrint records
created since 2008.

# OPTIONS

-h,-help,--help
: Display this help page

setup
: configure the export and import working directory

export
: export an EPrints repository into a dataset collection

import
: import the dataset collection into Invenio-RDM

keys
: works with export, harvest all keys from the EPrints repository

key_list
: works with export and import, the filename containing the ids to harvest

full
: export a list of all keys available from EPrints, then export or import the collections from those keys.


# ENVIRONMENT

The following environment varaibles are relied on.

EPRINT_HOST
: This hostname used for the EPrints repository, e.g. authors.library.caltech.edu

EPRINT_USER
: This is the username to access the EPrints repository with eprint2rdm

EPRINT_PASSWORD
: This is the password to access the EPrints repository with eprint2rdm

C_NAME
: This dataset collection to save the harvested EPrints content in.


# EXAMPLE

Set things up

~~~
${APP_NAME} setup
~~~


Harvest all records in repositories defined in environment.

~~~
${APP_NAME} full
~~~


EOT
}

function do_eprints_export() {
	FULL="$1"
	KEY_LIST="$2"
	echo "Creating $C_NAME as a Postgres based dataset collection"
	if [ ! -d "$C_NAME" ]; then
		read -r -p 'Enter the MySQL DB username: ' DB_USER
		echo -n 'Enter the MySQL DB user password: '
		read -r -s DB_PASSWORD
		DB_NAME="$(basename "${C_NAME}" .ds)_ds"
		mysql -E "CREATE DATABASE IF NOT EXIST ${DB_NAME};"
    	dataset init "${C_NAME}" "mysql://${DB_USER}:${DB_PASSWORD}@localhost/${DB_NAME}"
	fi
	echo "eprint2rdm respecting resources, people and groups"
	if [ "$FULL" = "true" ]; then
		KEY_LIST="${C_NAME}"_all_ids.txt
    	eprint2rdm -all-ids "$EPRINT_HOST" >"${C_NAME}_all_ids.txt"

	fi
    eprint2rdm -id-list "${KEY_LIST}" -harvest "${C_NAME}" \
	     -resource-map resource_types.csv \
         -contributor-map contributor_types.csv \
	     "$EPRINT_HOST"
}

function do_irdm_import() {
    	
	cat <<EOT

do_rdm_import not scripted yet.

EOT

}

function retrieve_csv_files() {
	curl -L -O https://feeds.library.caltech.edu/people/people.csv
	curl -L -O https://feeds.library.caltech.edu/groups/groups.csv
}

#
# Main processing
#
FULL="false"
SETUP="false"
EXPORT_EPRINTS="false"
IMPORT_IRDM="false"
REPO_ID=""
KEY_LIST=""
for ARG in "$@"; do
	case "${ARG}" in
		-h|-help|--help|help)
			usage
			exit 0
			;;
		full)
			FULL="true"
			;;
		setup)
			SETUP="true"
			;;
		export)
			EXPORT_EPRINTS="true"
			;;
		import)
			IMPORT_IRDM="true"
			;;
		*)
			if [[ "$REPO_ID" = "" ]]; then
				REPO_ID="${ARG}"
			else
				KEY_LIST="${ARG}"
			fi
			;;
	esac
done

if [ "$C_NAME" = "" ] || [ "$EPRINT_HOST" = "" ]; then
	SETUP="true"
fi

if [ "${SETUP}" = "true" ]; then
	read -r -p 'What is the repo id (e.g. caltechauthors)? ' REPO_ID
	read -r -p 'What is the EPrints hostname?  ' EPRINT_HOST
	read -r -p 'What is the EPrints username? ' EPRINT_USER
	echo -r -n 'What is the EPrints password? '
	read -r -s EPRINT_PASSWORD
	read -r -p 'What is the dataset collection name? ' C_NAME
	cat <<EOT >"${REPO_ID}.env"
#!/bin/sh
#
# Setup for $REPO_ID
# This will be sourced from the environment by $APP_NAME
#
REPO_ID="${REPO_ID}"
EPRINT_HOST="${EPRINT_HOST}"
EPRINT_USER="${EPRINT_USER}"
EPRINT_PASSWORD="${EPRINT_PASSWORD}"
C_NAME="${C_NAME}"

export REPO_ID
export EPRINT_HOST
export EPRINT_USER
export EPRINT_PASSWORD
export C_NAME

EOT
	chmod 600 "${REPO_ID}.env"
	cat <<EOT

   Wrote ${REPO_ID}.env configuration file
   If REPO_ID environment is set then this
   file will control how ${APP_NAME} runs.

EOT

fi

if [ "$REPO_ID" = "" ]; then
	read -r -p 'What is the repository id to migrate? ' REPO_ID
	export REPO_ID
	source "${REPO_ID}.env"
else
	echo "Using config ${REPO_ID}.env"
	source "${REPO_ID}.env"
fi

echo "Starting $(date)"
if [ "${EXPORT_EPRINTS}" = "true" ]; then
	retrieve_csv_files
	do_eprints_export "${FULL}" "${KEY_LIST}"
fi
if [ "${IMPORT_IRDM}" = "true" ]; then
	do_irdm_import "${FULL}" "${KEY_LIST}"
fi
echo "Completed $(date)"
