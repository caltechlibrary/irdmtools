#!/bin/bash
#

# Fetch Today's SQL dumps
YMD=$(date +%Y-%m-%d)
if [ ! -d rdm-sql-dumps ]; then
	mkdir -p rdm-sql-dumps
fi
if [ ! -d ep3-sql-dumps ]; then
	mkdir -p ep3-sql-dumps
fi
for HOST in authors.library.caltech.edu data.caltech.edu; do
	echo "scp ${HOST}:sql-dumps/*_${YMD}.sql.gz ./rdm-sql-dumps/"
	scp "${HOST}:sql-dumps/*_${YMD}.sql.gz" ./rdm-sql-dumps/
done
if [ "$HOSTNAME" != "eprints.library.caltech.edu" ]; then
	REMOTE_HOST="eprints.library.caltech.edu"
	for REPO_ID in thesis; do
		echo "scp ${REMOTE_HOST}:sql-dumps/*${REPO_ID}-dump_${YMD}.sql.gz ./ep3-sql-dumps/"
		scp "${REMOTE_HOST}:sql-dumps/*${REPO_ID}-dump_${YMD}.sql.gz" ./ep3-sql-dumps/
	done
else
	echo "Skipping, we have the data on eprints.library.caltech.edu"
fi

#scp datawork.library.caltech.edu:/Sites/feeds.library.caltech.edu/groups.csv ./
#scp datawork.library.caltech.edu:/Sites/feeds.library.caltech.edu/people.csv ./
scp datawork.library.caltech.edu:/Sites/feeds.library.caltech.edu/htdocs/groups/groups.csv ./
scp datawork.library.caltech.edu:/Sites/feeds.library.caltech.edu/htdocs/people/people.csv ./
