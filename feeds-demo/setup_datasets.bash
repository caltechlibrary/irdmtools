#!/bin/bash

#
# setup-datasets.bash creetes the Postgres databases and
# initialized the datasets need for feeds processing.
#
for DB_NAME in people groups authors data thesis; do
	if psql "${DB_NAME}" -c "" 2>/dev/null; then
		echo "${DB_NAME} exists, skipping setup"
	else
		echo "creating ${DB_NAME}"
		if createdb "$DB_NAME"; then
			dataset init "${DB_NAME}.ds" "postgres://$USER@localhost/$DB_NAME?sslmode=disable"
		else
			echo "Problem setting up for $DB_NAME, aborting setup"
		fi
	fi
done
