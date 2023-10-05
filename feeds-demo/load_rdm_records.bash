#!/bin/bash

function usage() {
    APP_NAME=$(basename "$0")
    cat <<EOT
% ${APP_NAME}(1) user manual
% R. S. Doiel
% 2022-10-20

# NAME

${APP_NAME}

# SYNOPSIS

${APP_NAME} REPO_ID [YYYY-MM-DD]

# DESCRIPTION

Load a snapshot of an RDM database replacing the existing database
if necessary. It assume your logged in user id has permissions to
load the database and the database name is the same as the REPO_ID.

NOTE: when you load the database a role needs to exist for that database.
This can be done before your run this script using

~~~
psql -c "CREATE ROLE <REPO_ID>"
~~~

Where <REPO_ID> is the name of the database (e.g. caltechdata in the
examples).

# EXAMPLE

~~~
    ${APP_NAME} caltechdata 2022-10-20
~~~

EOT

}

function restore_postgres_from() {
	DB_NAME="$1"
	BACKUP_FILE="$2"
	ZCAT="zcat"
	# Handle case for macOS where zcat is called gzcat
	if command -v gzcat &>/dev/null; then
		ZCAT=gzcat
	fi
	PG_VERSION=$(psql --version | cut -d\  -f 3)
	echo "Dropping stale ${DB_NAME} if it exists"
	dropdb --if-exists "${DB_NAME}"
	echo "Creating fresh ${DB_NAME}"
	createdb "${DB_NAME}"
	echo "Loading ${BACKUP_FILE}, this can take a while"
	if "$ZCAT" "${BACKUP_FILE}" | psql "${DB_NAME}" >/dev/null; then 
		echo "Success!"
	else
		echo "Something went wrong. :-("
		exit 10
	fi
}

function run_restore() {
	REPO_ID="$1"
	SQL_FILE="$2"
	#
	# Sanity check our requiremented environment
	#
	SCRIPTNAME="$(readlink -f "$0")"
	DNAME="$(dirname "${SCRIPTNAME}")"
	cd "${DNAME}" || exit 1
	restore_postgres_from "$REPO_ID" "$SQL_FILE"
}


#
# Main process
#
for ARG in "$@"; do
    case $ARG in
    -h | -help | --help)
        usage
        exit 0
        ;;
    esac
done

case "$#" in
	"0")
    usage
    exit 1
	;;
	"1")
	REPO_ID="$1"
	SNAPSHOT="$(date +%Y-%m-%d)"
	SQL_FILE="rdm-sql-dumps/${REPO_ID}-dump_${SNAPSHOT}.sql.gz"
	;;
	*)
	REPO_ID="$1"
	SNAPSHOT="$2"
	SQL_FILE="rdm-sql-dumps/${REPO_ID}-dump_${SNAPSHOT}.sql.gz"
	;;
esac
if [ "${REPO_ID}" = "" ]; then
	echo "Missing repo id, aborting"
	exit 1
fi
if [ ! -f "${SQL_FILE}" ]; then
    echo "Can't find ${SQL_FILE}, aborting"
    exit 1
fi
run_restore "$REPO_ID" "$SQL_FILE"
