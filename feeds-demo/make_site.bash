#!/bin/bash
#

#
# populate the htdocs/recent folder with JSON content
#
function make_recent_folder() {
	echo "Populating recent folder"
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
	echo "Making ${REPO} folder"
	# Cleanup stale stuff
	if [ -d "htdocs/$REPO" ]; then
		rm -fR "htdocs/$REPO"
	fi
    # Setup clean repository directory	
	mkdir -p "htdocs/$REPO"
	# FIXME: Need to create a pairtree version of the repository as a Zip file here
	dsquery -csv "_Key,updated" -sql "${REPO}-updated.sql" "$REPO.ds" >"htdocs/${REPO}/updated.csv"
	# Clone the repository and zip it up.
	FEEDS_C_NAME="Caltech$(printf '%s\n' "$REPO" | awk '{ print toupper($0) }').ds"
	FEEDS_KEY_LIST="Caltech$(printf '%s\n' "$REPO" | awk '{ print toupper($0) }').keys"
	dataset clone -all -i "htdocs/${REPO}/${FEEDS_KEY_LIST}" "${REPO}.ds" "htdocs/${REPO}/${FEEDS_C_NAME}"
	# Make sure we have collection to Zip up
	if [ -d "htdocs/${REPO}/${FEEDS_C_NAME}" ]; then
		CWD=$(pwd)
		cd "htdocs/${REPO}" || exit
		if zip -r "${FEEDS_C_NAME}.zip" "${FEEDS_C_NAME}"; then
			echo "Zipping complete."
			#sleep 30
			#echo "Removing ${FEEDS_C_NAME} and ${FEEDS_KEY_LIST}"
			#rm -fR "${FEEDS_C_NAME}"
			#rm "${FEEDS_KEY_LIST}"
		fi
		cd "${CWD}" || exit
	fi
}

function check_for_required_programs() {
	IS_MISSING=""
	for CLI in dataset dsquery python pandoc jsonrange jq; do
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

function make_groups_index_md() {
	echo '---'
	echo 'groups:'
	dsquery -sql groups-index-md.sql groups.ds | json2yaml | sed -E 's/^/    /g'
	echo '---'
}

function make_groups() {
	echo "Populating groups folder"
	mkdir -p htdocs/groups
	dataset keys groups.ds >htdocs/groups/index.keys
	dsquery -pretty -sql groups-index-json.sql groups.ds >htdocs/groups/index.json
	# Now build index.md for groups
	make_groups_index_md | pandoc -f markdown -t markdown \
					  --template templates/groups-index-md.tmpl \
					  >htdocs/groups/index.md
	# Now build the old group_list.json (this get used by CL.js and the widget stuff)
	
#FIXME: Need to generate these files
# groups.csv
# updated.csv

	# Clean and clone groups.ds collection to CaltechGROUPS.ds.zip
	if [ -d htdocs/groups/CaltechGROUPS.ds ]; then
		rm -fR htdocs/groups/CaltechGROUPS.ds
	fi
	dataset clone -all groups.ds htdocs/groups/CaltechGROUPS.ds
	if [ -d "htdocs/groups/CaltechGROUPS.ds" ]; then
		CWD=$(pwd)
		cd "htdocs/groups" || exit
		if zip -r "CaltechGROUPS.ds.zip" "CaltechGROUPS.ds"; then
			echo "Zipping complete."
		fi
		cd "${CWD}" || exit
	fi
}

function make_people() {
	echo "Populating people folder"
	mkdir -p htdocs/people
#FIXME: Need to determine the files to generate here.
}

function make_root_folder_grids() {
	echo "Populating root folder"
	mkdir -p htdocs
	for REPO in authors data thesis; do
		if [ -f "${REPO}.env" ]; then
			dsquery -pretty -grid='_Key,date,date_type,title,creators,local_group,type,url' \
			        -sql "${REPO}-grid.sql" "${REPO}.ds" \
					>"htdocs/caltech${REPO}-grid.json"
		else
			echo "missing ${REPO}.env, skipped making htdocs/caltech${REPO}-grid.json"
		fi
	done
}

function make_repo_folders() {
	# Build the recent folder for each repository's content
	for REPO in authors thesis data; do
		if [ -f "${REPO}.env" ]; then
			# shellcheck disable=SC1090
			. "${REPO}.env"
			make_recent_folder
			make_repo_folder "${REPO}"
		else
			echo "Missing ${REPO}.env, skipping"
		fi
	done
}

#
# Main processing loop to generate our website.
#
check_for_required_programs

if [ "$1" != "" ]; then
	for cmd in $@; do
		echo "Running $cmd"
		if ! $cmd; then
			echo 'Something went wrong'
			exit 10
		fi
	done
	exit 0
fi

# Build root folder contents.
make_root_folder_grids

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

