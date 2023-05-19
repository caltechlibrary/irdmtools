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


function setup_dataset_collection() {
	if [ ! -d "$C_NAME" ]; then
		echo "Need to create $C_NAME using MySQL JSON store"
		if [ "$DB_USER" = "" ]; then
			read -r -p 'MySQL DB username: ' DB_USER
		fi
		if [ "$DB_PASSWORD" = "" ]; then
			echo -n 'MySQL DB user password: '
			read -r -s DB_PASSWORD
		fi
		DB_NAME="$(basename "${C_NAME}" .ds)_ds"
		CONN="mysql://${DB_USER}:${DB_PASSWORD}@/${DB_NAME}"
		echo "Creating MySQL database $DB_NAME"
		echo "with $CONN"
		mysql -e "CREATE DATABASE IF NOT EXIST ${DB_NAME};"
		echo "Initializing $_CNAME using MySQL JSON Store"
    	if ! dataset init "${C_NAME}" "${CONN}"; then
			cat <<EOT
Failed:

	dataset init "${C_NAME}" "${CONN}"

EOT
			exit 1
		fi
		if [ ! -d "$C_NAME" ]; then
			echo "Something went wrong creating $_CNAME, aborting"
			exit 1
		fi
		if ! dataset count "$C_NAME"; then
			echo "Something went wrong access $C_NAME, aborting"
			exit 1
		fi
	fi
	echo "Using $C_NAME"
}

function do_eprints_export() {
	FULL="$1"
	KEY_LIST="$2"
	if [ "$FULL" = "true" ]; then
		echo "eprint2rdm -all-ids $EPRINT_HOST >${REPO_ID}_all_ids.txt"
		KEY_LIST="${REPO_ID}_all_ids.txt"
    	if eprint2rdm -all-ids "$EPRINT_HOST" >"${REPO_ID}_all_ids.txt"; then
			# NOTE need a trailing new line for while loops
			echo '' >>"${REPO_ID}_all_ids.txt"
		else
			exit 1
		fi

	fi
	if [ "${KEY_LIST}" = "" ] || [ ! -f "${KEY_LIST}" ]; then
		echo 'Missing key list for export'
		exit 1
	fi
	echo "eprint2rdm respecting resources, people and groups"
    if eprint2rdm -id-list "${KEY_LIST}" -harvest "${C_NAME}" \
	     -resource-map resource_types.csv \
         -contributor-map contributor_types.csv \
	     "${EPRINT_HOST}"; then
		exit 1
	fi

}

function do_rdm_import() {
	FULL="$1"
	KEY_LIST="$2"
	if [ "${FULL}" = "true" ]; then
		echo 'Setting up to import all ids'
		KEY_LIST="${REPO_ID}_all_ids.txt"
		if [ -f "${KEY_LIST}" ]; then
			echo "WARNING: using existing ids in ${KEY_LIST}"
		fi
	fi
	if [ "$KEY_LIST" = "" ]; then
		echo 'Missing key list name'
		exit 1
	fi
	# NOTE: need to get path to dataset collection from settings.json
	if [ ! -f "${KEY_LIST}" ]; then
		echo "Getting all keys in ${C_NAME}, saving in ${KEY_LIST}.txt"
		dataset keys "$C_NAME" >"${KEY_LIST}"
	fi
	echo "$(wc -l "${KEY_LIST}") keys found in ${KEY_LIST}"
	echo ""
	if [ -f "${KEY_LIST}" ]; then
		echo "running migrate_records.py -exit_on_error -eprintids ${KEY_LIST} ${C_NAME}"
		if ! python3 migrate_records.py -exit_on_error -eprintids "${KEY_LIST}" "$C_NAME"; then
			echo "Failed to process ${KEY_LIST} using ${C_NAME}"
			exit 1
		fi
	else
		echo "${KEY_LIST} appears to be empty for $C_NAME"
	fi
}

function retrieve_csv_files() {
	if [ ! -f people.csv ]; then
	curl -L -o people.csv https://feeds.library.caltech.edu/people/people.csv
	fi
	if [ ! -f groups.csv ]; then
	curl -L -o groups.csv  https://feeds.library.caltech.edu/groups/groups.csv
	fi
}

#
# Main processing
#
FULL="false"
SETUP="false"
EXPORT_EPRINTS="false"
IMPORT_IRDM="false"
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
			KEY_LIST="${ARG}"
			;;
	esac
done

if [ "$REPO_ID" = "" ]; then
	read -r -p 'What is the repository id to migrate? ' REPO_ID
	export REPO_ID
fi

if [ "$REPO_ID" = "" ] || [ "$C_NAME" = "" ] || [ "$EPRINT_HOST" = "" ]; then
	SETUP="true"
fi



if [ "${SETUP}" = "true" ] && [ ! -f "${REPO_ID}.env" ]; then
    echo 'Setup EPrints REST API access'
	echo '----------------------------------'
	read -r -p 'What is the repo id (e.g. caltechauthors)? ' REPO_ID
	read -r -p 'What is the EPrints hostname?  ' EPRINT_HOST
	read -r -p 'What is the EPrints username? ' EPRINT_USER
	echo -n 'What is the EPrints password? (will not be shown)'
	read -r -s EPRINT_PASSWORD
	ehco ''
    echo 'Setup dataset collection to hold exported records'
	echo '--------------------------------------------------'
	read -r -p 'What is the dataset collection name? ' C_NAME
    read -r -p 'What is the DB_USER for SQL Store? ' DB_USER
	echo -n 'What is the DB_PASSWORD for SQL Store? (will not be shown)'
	read -r -s DB_PASSWORD
	echo 'Setup RDM access'
	read -r -p 'What is the RDM_URL? ' RDM_URL
	echo -r -s 'What is the RDM Access Token? (will not be shown)'
	cat <<EOT >"${REPO_ID}.env"
#!/bin/sh
#
# Setup for $REPO_ID
# This will be sourced from the environment by $APP_NAME
#
REPO_ID="${REPO_ID}"

# EPrints REST API Access
EPRINT_HOST="${EPRINT_HOST}"
EPRINT_USER="${EPRINT_USER}"
EPRINT_PASSWORD="${EPRINT_PASSWORD}"

# Dataset collection setup
C_NAME="${C_NAME}"
DB_USER="${DB_NAME}"
DB_PASSWORD="${DB_PASSWORD}"

# RDM API Access
RDM_URL="${RDM_URL}"
RDMTOK="${RDMTOK}"

export REPO_ID
export EPRINT_HOST
export EPRINT_USER
export EPRINT_PASSWORD
export C_NAME
export DB_USER
export DB_PASSWORD
export RDM_URL
export RDMTOK

EOT
	chmod 600 "${REPO_ID}.env"
	cat <<EOT

   Wrote ${REPO_ID}.env configuration file
   If REPO_ID environment is set then this
   file will control how ${APP_NAME} runs.

EOT

fi

if [ ! -f "${REPO_ID}.env" ]; then
	echo "Can't find ${REPO_ID}.env, aborting"
	exit 1
fi
# shellcheck disable=SC1090
source "${REPO_ID}.env"

retrieve_csv_files
setup_dataset_collection

if [ "${EXPORT_EPRINTS}" != "true" ] && [ "${IMPORT_IRDM}" != "true" ]; then
    echo "Nothing left to do."
    exit 0
fi

echo "Starting eprint export=${EXPORT_EPRITNS}, rdm import=${IMPORT_IRDM} $(date)"
if [ "${EXPORT_EPRINTS}" = "true" ]; then
	do_eprints_export "${FULL}" "${KEY_LIST}"
fi
if [ "${IMPORT_IRDM}" = "true" ]; then
	do_rdm_import "${FULL}" "${KEY_LIST}"
fi
echo "Completed eprint export=${EXPORT_EPRITNS}, rdm import=${IMPORT_IRDM} $(date)"
