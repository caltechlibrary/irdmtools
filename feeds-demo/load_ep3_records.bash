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

${APP_NAME} REPO_ID YYYY-MM-DD

# DESCRIPTION

Load a snapshot of the EPrints database replacing the existing database
if necessary.

# EXAMPLE

~~~
    ${APP_NAME} caltechthesis 2022-10-20
~~~

EOT

}

#
# Main process
#
if [ "$HOSTNAME" = "eprints.library.caltech.edu" ]; then
	echo "Not needed on eprints.library.caltech.edu, it already has the databases"
	exit 1
fi
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
	SQL_FILE="ep3-sql-dumps/${REPO_ID}-dump_${SNAPSHOT}.sql.gz"
	;;
	*)
	REPO_ID="$1"
	SNAPSHOT="$2"
	SQL_FILE="ep3-sql-dumps/${REPO_ID}-dump_${SNAPSHOT}.sql.gz"
	;;
esac

if [ ! -f "${SQL_FILE}" ]; then
    echo "Can't find ${SQL_FILE}, aborting"
    exit 1
fi
mysql --execute "DROP DATABASE IF EXISTS ${REPO_ID};"
mysql --execute "CREATE DATABASE IF NOT EXISTS ${REPO_ID};"
gzcat "ep3-sql-dumps/${REPO_ID}-dump_${SNAPSHOT}.sql.gz" | mysql "${REPO_ID}"
