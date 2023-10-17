#!/bin/bash
#

#
# populate the htdocs/recent folder with JSON content
#
function make_recent() {
	echo "Populating recent folder"
	mkdir -p htdocs/recent
	# Populate recent folder for CaltechAUTHORS
	if [ -f authors.env ]; then
		# shellcheck disable=SC1090,SC1091
		. authors.env
		dsquery -pretty -sql recent_authors_object_types.sql authors.ds >htdocs/recent/object_types.json
		dsquery -pretty -sql recent_authors_combined.sql authors.ds >htdocs/recent/combined.json
		for T in article audiovisual book book_section collection combined_data conference_item data_object_types data_pub_types dataset image interactiveresource model monograph object_types patent pub_types software teaching_resource text thesis video workflow; do
			dsquery -pretty -sql recent_authors_for_type.sql authors.ds "${T}" >"htdocs/recent/$T.json"
		done
	fi

	# NOTE: We don't do recent lists for CaltechTHESIS, doesn't really make sense.

	# Populate recent folder for CaltechDATA
	if [ -f data.env ]; then
		# shellcheck disable=SC1090,SC1091
		. data.env
		dsquery -pretty -sql recent_data_object_types.sql data.ds >htdocs/recent/data_object_types.json
		dsquery -pretty -sql recent_data_combined.sql data.ds >htdocs/recent/data_combined.json
		for T in collection dataset image image_map interactive_resource model other publication software video workflow; do
			dsquery -pretty -sql recent_data_for_type.sql data.ds "${T}" >"htdocs/recent/data_$T.json"
		done
	fi
}

#
# Populate the repository folder (e.g. authors, thesis, data)
#
function make_repo() {
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
	for CLI in dataset dsquery pandoc jsonrange jq python3; do
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
	dsquery -sql groups_index_md.sql groups.ds | json2yaml | sed -E 's/^/    /g'
	echo '---'
}

function make_group_list_json() {
		if [ ! -f groups.csv ]; then
			echo "failed to find groups.csv, skipping making groups"
			return
		fi
		dsquery -csv 'id,type,pub_date,local_group,collection' \
		        -sql get_authors_by_group.sql authors.ds \
				>htdocs/groups/group_authors.csv
		if [ ! -f "htdocs/groups/group_authors.csv" ]; then
			echo "failed to find group_authors.csv, skipping making groups"
			return
		fi
		dsquery -csv 'id,thesis_type,pub_date,local_group,collection' \
		        -sql get_thesis_by_group.sql thesis.ds \
				>htdocs/groups/group_thesis.csv
		if [ ! -f "htdocs/groups/group_thesis.csv" ]; then
			echo "failed to find group_thesis.csv, skipping making groups"
			return
		fi
		dsquery -csv 'id,type,pub_date,local_group,collection' \
		        -sql get_data_by_group.sql data.ds \
				>htdocs/groups/group_data.csv
		if [ ! -f "htdocs/groups/group_data.csv" ]; then
			echo "failed to find group_data.csv, skipping making groups"
			return
		fi
		python3 aggregate_resource_types.py groups.csv \
		    htdocs/groups/group_authors.csv \
		    htdocs/groups/group_thesis.csv \
		    htdocs/groups/group_data.csv \
			>htdocs/groups/group_list.json
}

function make_group_authors() {
	GROUP_ID="$1"
	if [ "$GROUP_ID" = "" ]; then
		echo "Missing GROUP_ID aborting"
		exit 11
	fi
	echo "Processing combined authors.ds and groups.ds for $GROUP_ID"
	# Write out htdocs/groups/<GROUP_ID>/combined.json
	if [ -f "authors_group_combined_types.sql" ]; then
		GROUP_JSON=$(printf '[%s]' "\"${GROUP_ID}\"")
		GROUP_FILE="htdocs/groups/$GROU_ID/combined.json"
		dsquery -pretty -sql "authors_group_combined_types.sql" \
		    authors.ds \
			"$GROUP_JSON" \
			>"$GROUP_FILE"
		L=$(jq '. | length' "$GROUP_FILE")
		if [ "$L" = "0" ]; then
			rm "$GROUP_FILE"
		else
			echo "Wrote ($L) $GROUP_FILE"
		fi
	fi
}

function make_group_thesis() {
	GROUP_ID="$1"
	if [ "$GROUP_ID" = "" ]; then
		echo "Missing GROUP_ID aborting"
		exit 11
	fi
	echo "Processing combined thesis.ds and groups.ds for $GROUP_ID"
	# Write out htdocs/groups/<GROUP_ID>/combined_thesis.json
	if [ -f "thesis_group_combined_types.sql" ]; then
		GROUP_JSON=$(printf '[%s]' "\"${GROUP_ID}\"")
		GROUP_FILE="htdocs/groups/$GROUP_ID/combined_thesis.json"
		dsquery -pretty -sql "thesis_group_combined_types.sql" \
		    thesis.ds \
			"$GROUP_JSON" \
			>"$GROUP_FILE"
		L=$(jq '. | length' "$GROUP_FILE")
		if [ "${L}" = "0" ]; then
			rm "$GROUP_FILE"
		else
			echo "Wrote ($L) $GROUP_FILE"
		fi
	fi
}

function make_group_data() {
	GROUP_ID="$1"
	if [ "$GROUP_ID" = "" ]; then
		echo "Missing GROUP_ID aborting"
		exit 11
	fi
	echo "Processing combined data.ds and groups.ds for $GROUP_ID"
	# Write out htdocs/groups/<GROUP_ID>/combined_data.json
	if [ -f "data_group_combined_types.sql" ]; then
		GROUP_JSON=$(printf '[%s]' "\"${GROUP_ID}\"")
		GROUP_FILE="htdocs/groups/$GROUP_ID/combined_data.json"
		dsquery -pretty -sql "data_group_combined_types.sql" \
		    data.ds \
			"$GROUP_JSON" \
			>"$GROUP_FILE"
		L=$(jq '. | length' "$GROUP_FILE")
		if [ "$L" = "0" ]; then
			rm "$GROUP_FILE"
		else
			echo "Wrote ($L) $GROUP_FILE"
		fi
	fi
}

function make_group_folders() {
	for GROUP_ID in $(jsonrange -values -i htdocs/groups/index.json | jq -r); do
		if [ "$GROUP_ID" = "" ]; then
			echo "error: GROUP_ID is not set from htdocs/groups/index.json"
			exit 11
		fi
		# Create the group directory if neccessary
		GROUP_DIR="htdocs/groups/$GROUP_ID"
		if [ ! -d "${GROUP_DIR}" ]; then
			mkdir -p "$GROUP_DIR"
		fi
		# Process CaltechAUTHORS
		make_group_authors "$GROUP_ID"
		# Process CaltechTHESIS
		make_group_thesis "$GROUP_ID"
		# Process CaltechDATA
		make_group_data "$GROUP_ID"
	done
}

function make_groups() {
	echo "Populating groups folder"
	mkdir -p htdocs/groups
	dataset keys groups.ds >htdocs/groups/index.keys
	dsquery -pretty -sql groups_index_json.sql groups.ds >htdocs/groups/index.json
	# Now build index.md for groups
	make_groups_index_md | pandoc -f markdown -t markdown \
					  --template templates/groups_index_md.tmpl \
					  >htdocs/groups/index.md
	make_group_list_json 
	python3 generate_group_json.py htdocs/groups/group_list.json
}

function clone_groups_ds() {
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

function clone_people_ds() {
	# Clean and clone people.ds collection to CaltechPEOPLE.ds.zip
	if [ -d htdocs/people/CaltechPEOPLE.ds ]; then
		rm -fR htdocs/people/CaltechPEOPLE.ds
	fi
	dataset clone -all people.ds htdocs/people/CaltechPEOPLE.ds
	if [ -d "htdocs/people/CaltechPEOPLE.ds" ]; then
		CWD=$(pwd)
		cd "htdocs/people" || exit
		if zip -r "CaltechPEOPLE.ds.zip" "CaltechPEOPLE.ds"; then
			echo "Zipping complete."
		fi
		cd "${CWD}" || exit
	fi
}

function make_root() {
	echo "Populating root folder"
	mkdir -p htdocs
	for REPO in authors data thesis; do
		if [ -f "${REPO}.env" ]; then
			dsquery -pretty -grid='_Key,date,date_type,title,creators,local_group,type,url' \
			        -sql "${REPO}_grid.sql" "${REPO}.ds" \
					>"htdocs/caltech${REPO}-grid.json"
		else
			echo "missing ${REPO}.env, skipped making htdocs/caltech${REPO}_grid.json"
		fi
	done
}

function clone_thesis() {
	REPO="thesis"
	if [ ! -f "${REPO}.env" ]; then
		echo "Missing ${REPO}.env, skipping"
		return
	fi
	# shellcheck disable=SC1090
	. "${REPO}.env"
	make_repo "${REPO}"
}

function clone_data() {
	REPO="data"
	if [ ! -f "${REPO}.env" ]; then
		echo "Missing ${REPO}.env, skipping"
		return
	fi
	# shellcheck disable=SC1090
	. "${REPO}.env"
	make_repo "${REPO}"
}

function clone_authors() {
	REPO="authors"
	if [ ! -f "${REPO}.env" ]; then
		echo "Missing ${REPO}.env, skipping"
		return
	fi
	# shellcheck disable=SC1090
	. "${REPO}.env"
	make_repo "${REPO}"
}

#
# Main processing loop to generate our website.
#
check_for_required_programs

if [ "$1" != "" ]; then
	param=""
	cmd=""
	for arg in "$@"; do
		case "$arg" in
		clone_*)
			cmd="${arg}"
			;;
		root|recent|groups|people|group_folders|group_authors|group_thesis|group_data)
			cmd="make_${arg}"
			;;
		groups_*)
			echo "did you mean group_authors, group_thesis or group_data?"
			exit 1
			;;
		*)
			param="$arg"
			;;
		esac
	done
	if [ "$param" != "" ]; then
		echo "Running -> '$cmd' '$param'"
		if ! "$cmd" "$param"; then
			echo 'Something went wrong'
			exit 10
		fi
	else
		echo "Running -> $cmd"
		if ! "${cmd}"; then
			echo 'Something went wrong'
			exit 10
		fi
	fi
	exit 0
fi

##   START_TIME=$(date)
# Build root folder contents.
make_root
# Build  recent folder
make_recent
# Build out the groups tree
make_groups
make_group_folders
# Build out the people tree
make_people

##   echo "Starting to clone dataset collections (takes a while)"
##   # Clone groups.ds
##   clone_groups
##   # Build thesis folder
##   clone_thesis
##   # Build data folder
##   clone_data
##   # Build authors folder
##   clone_authors
##   END_TIME=$(date)
##   echo "Completed, start ${START_TIME}, finished ${END_TIME}"
