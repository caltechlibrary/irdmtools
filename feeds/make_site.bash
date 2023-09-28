#!/bin/bash
#

#
# populate the htdocs/recent folder with JSON content
#
function make_recent_folder() {
	mkdir -p htdocs/recent
	dsquery -pretty -sql recent-object-types.sql authors.ds >htdocs/recent/object_types.json
	dsquery -pretty -sql recent-combined.sql authors.ds >htdocs/recent/combined.json
	for T in article audiovisual book book_section collection combined_data conference_item data_object_types data_pub_types dataset image interactiveresource model monograph object_types patent pub_types software teaching_resource text thesis video workflow; do
		dsquery -pretty -sql recent-for-type.sql authors.ds "${T}" >"htdocs/recent/$T.json"
	done
}

#
# Populate the repository folder (e.g. authors, thesis, data)
#
function make_repo_folder() {
	REPO="$1"
	mkdir -p "htdocs/$REPO"
	# FIXME: Need to create a pairtree version of the repository as a Zip file here
	dsquery -pretty -sql "${REPO}-updated.sql" "$REPO.ds" >"htdocs/${REPO}/updated.json"
	# Convert updated.json to CSV
	python jsonlist_to_csv.py "htdocs/${REPO}/updated.json" >"htdocs/${REPO}/updated.csv"
}

for CFG in authors thesis data; do
	if [ -f "${CFG}.env" ]; then
		. "${CFG}.env"
		make_repo_folder "${CFG}"
		make_recent_folder
	else
		echo "Missing ${CFG}.env, skipping"
	fi
done
