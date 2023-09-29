#!/bin/bash
#

#
# populate the htdocs/recent folder with JSON content
#
function make_recent_folder() {
	mkdir -p htdocs/recent
	if [ -f authors.env ]; then
		# shellcheck disable=SC1090
		. authors.env
		dsquery -pretty -sql recent-authors-object-types.sql authors.ds >htdocs/recent/object_types.json
		dsquery -pretty -sql recent-authors-combined.sql authors.ds >htdocs/recent/combined.json
		for T in article audiovisual book book_section collection combined_data conference_item data_object_types data_pub_types dataset image interactiveresource model monograph object_types patent pub_types software teaching_resource text thesis video workflow; do
			dsquery -pretty -sql recent-authors-for-type.sql authors.ds "${T}" >"htdocs/recent/$T.json"
		done
	fi
	if [ -f data.env ]; then
		# shellcheck disable=SC1090
		. data.env
		dsquery -pretty -sql recent-data-object-types.sql data.ds >htdocs/recent/data_object_types.json
		dsquery -pretty -sql recent-data-combined.sql data.ds >htdocs/recent/data_combined.json
		for T in collection dataset image image_map interactive_resource model other publication software video workflow; do
			dsquery -pretty -sql recent-data-for-type.sql data.ds "${T}" >"htdocs/recent/data_$T.json"
		done
	fi
}

#
# Populate the repository folder (e.g. authors, thesis, data)
#
function make_repo_folder() {
	REPO="$1"
	# Cleanup stale stuff
	if [ -d "htdocs/$REPO" ]; then
		rm -fR "htdocs/$REPO"
	fi
    # Setup clean repository directory	
	mkdir -p "htdocs/$REPO"
	# FIXME: Need to create a pairtree version of the repository as a Zip file here
	dsquery -pretty -sql "${REPO}-updated.sql" "$REPO.ds" >"htdocs/${REPO}/updated.json"
	# Convert updated.json to CSV
	if python jsonlist_to_csv.py "htdocs/${REPO}/updated.json" >"htdocs/${REPO}/updated.csv"; then
		rm "htdocs/${REPO}/updated.json"
	fi
	# Clone the repository and zip it up.
	FEEDS_C_NAME="Caltech$(printf '%s\n' "$REPO" | awk '{ print toupper($0) }').ds"
	FEEDS_KEY_LIST="Caltech$(printf '%s\n' "$REPO" | awk '{ print toupper($0) }').keys"
	dataset keys "${REPO}.ds" >"htdocs/${REPO}/${FEEDS_KEY_LIST}"
	dataset clone -i "htdocs/${REPO}/${FEEDS_KEY_LIST}" "${REPO}.ds" "htdocs/${REPO}/${FEEDS_C_NAME}"
	# Make sure we have collection to Zip up
	if [ -d "htdocs/${REPO}/${FEEDS_C_NAME}" ]; then
		CWD=$(pwd)
		cd "htdocs/${REPO}" || exit
		if zip -r "${FEEDS_C_NAME}.zip" "${FEEDS_C_NAME}"; then
			echo "Zipping completed waiting 30 seconds for disk to settle."
			sleep 30
			echo "Removing ${FEEDS_C_NAME} and ${FEEDS_KEY_LIST}"
			rm -fR "${FEEDS_C_NAME}"
			rm "${FEEDS_KEY_LIST}"
		fi
		cd "${CWD}" || exit
	fi
}

function check_for_required_programs() {
	IS_MISSING=""
	for CLI in dsquery python pandoc; do
		if ! command -v "${CLI}" &>/dev/null; then
			IS_MISSING="true"
			echo "Missing ${CLI}"
		fi
	done
	if [ "$IS_MISSING" != "" ]; then
		echo "Missing requirements, aborting"
		exit 10
	fi
}

function make_groups() {
	mkdir -p htdocs/groups
# make index.json
# index.keys
# index.md
# group_list.json
# groups.csv
# updated.csv
}

function make_people() {
	mkdir -p htdocs/people
}

#
# Main processing loop to generate our website.
#
check_for_required_programs
# Build the recent folder for each repository's content
for REPO in authors thesis data; do
	if [ -f "${REPO}.env" ]; then
		# shellcheck disable=SC1090
		. "${REPO}.env"
		make_recent_folder
		#make_repo_folder "${REPO}"
	else
		echo "Missing ${REPO}.env, skipping"
	fi
done

# Build out the groups tree
make_groups
# Build out the people tree
make_people

# Build the repo folder for each repository
## for REPO in authors thesis data; do
## 	if [ -f "${REPO}.env" ]; then
## 		# shellcheck disable=SC1090
## 		. "${REPO}.env"
## 		# make_repo_folder "${REPO}"
## 	else
## 		echo "Missing ${REPO}.env, skipping"
## 	fi
## done

