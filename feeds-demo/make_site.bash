#!/bin/bash
#

#
# populate the htdocs/recent folder with JSON content
#
function make_recent() {
	echo "Populating recent folder json"
	mkdir -p htdocs/recent
	# Populate recent folder for CaltechAUTHORS
	if [ -f authors.env ]; then
		# shellcheck disable=SC1090,SC1091
		. authors.env
		dsquery -pretty -sql authors_object_types.sql authors.ds >htdocs/recent/object_types.json
		dsquery -pretty -sql recent_authors_combined.sql authors.ds >htdocs/recent/combined.json
		jsonrange -values <htdocs/recent/object_types.json | while read -r src; do
			field_name=$(echo "${src}" | jq -r .name)
			dsquery -pretty -sql recent_authors_for_type.sql authors.ds "${field_name}" >"htdocs/recent/${field_name}.json"
		done
	fi

	# NOTE: We don't do recent lists for CaltechTHESIS, doesn't really make sense.

	# Populate recent folder for CaltechDATA
	if [ -f data.env ]; then
		# shellcheck disable=SC1090,SC1091
		. data.env
		dsquery -pretty -sql data_object_types.sql data.ds >htdocs/recent/data_object_types.json
		dsquery -pretty -sql recent_data_combined.sql data.ds >htdocs/recent/combined_data.json
		jsonrange -values <htdocs/recent/data_object_types.json | while read -r src; do
			field_name=$(echo "${src}" | jq -r .name)
			dsquery -pretty -sql recent_data_for_type.sql data.ds "${field_name}" >"htdocs/recent/${field_name}.json"
		done
	fi
	echo "Populating recent folder markdown"
	echo "" >htdocs/recent/index.md
	./wrap_recent.py htdocs/recent/object_types.json '{"repository": "CaltechAUTHORS", "combined": "combined"}' |\
		pandoc -f markdown -t markdown --template=templates/recent-index.md >>htdocs/recent/index.md
	./wrap_recent.py htdocs/recent/data_object_types.json '{"repository": "CaltechDATA", "combined": "combined_data"}' |\
		pandoc -f markdown -t markdown --template=templates/recent-index.md >>htdocs/recent/index.md
	
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

function make_group_pages() {
	GROUP_ID="$1"
	GROUP_NAME=$(jq -r .name htdocs/groups/$GROUP_ID/group.json)
	# Build up the index.md for group.
	FNAME="htdocs/groups/${GROUP_ID}/index.md"
	# build up resource type lists for CaltechAUTHORS, CaltechTHESIS and CaltechDATA lists
	./wrap_group_info.py "htdocs/groups/$GROUP_ID/group.json" |\
		pandoc -f markdown -t markdown \
		--template templates/groups-group-index.md \
		>"${FNAME}"
	#FIXME: Need to generate combined pages in Markdown
	#FIXME: from markdown resource documents we can use Pandoc to generate HTML and HTML include
	#FIXME: Need to render BibTeX and RSS from templates
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

function make_groups() {
	echo "Populating groups folder json"
	mkdir -p htdocs/groups
	dataset keys groups.ds >htdocs/groups/index.keys
	dsquery -pretty -sql groups_index_json.sql groups.ds >htdocs/groups/index.json
	dsquery -pretty -sql authors_object_types.sql authors.ds >htdocs/groups/authors_object_types.json
	dsquery -pretty -sql thesis_thesis_types.sql thesis.ds >htdocs/groups/thesis_thesis_types.json
	dsquery -pretty -sql data_object_types.sql data.ds >htdocs/groups/data_object_types.json
	# Now build index.md for groups
	./wrap_groups_a_to_z.py htdocs/groups/index.json | pandoc -f markdown -t markdown \
					  --template templates/groups-index.md \
					  >htdocs/groups/index.md
	make_group_list_json 
	python3 generate_group_files.py htdocs/groups/group_list.json
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
	echo "Populating people folder json"
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
	# generate JSON docs
	echo "Populating root folder json"
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
	# generate Markdown docs
	# generate HTML docs
	echo "Rendering HTML docs"
	for FNAME in index.md about.md error.md formats-and-extensions.md; do
		BNAME=$(basename "${FNAME}" ".md")
		echo "rendering ${FNAME} as ${BNAME}.html"
		pandoc --metadata title="$BNAME" \
			-s --template=templates/page.html \
			"htdocs/${FNAME}" \
			>"htdocs/${BNAME}.html"
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

function make_static() {
	# Copy in static content
	cp -vR static/* htdocs/
}

function page_title_from_path() {
	NAME="${1/htdocs/}"
	DNAME="$(dirname "$NAME")"
	# Strip off the .md extension
	FNAME="$(basename "$NAME" ".md")"
	if [ "$FNAME" =  "" ] || [[ "$FNAME" = "index" && "$DNAME" = "" ]]; then
			echo "Caltech Library Feeds"
	else
		string englishtitle "$(echo "$DNAME $FNAME" | tr '/' ' ')" | sed -E 's/ / > /g;s/-/ /g'
	fi
}

function make_html() {
	find htdocs -type f | grep -E '\.md$' | while read -r FNAME; do
		DNAME=$(dirname "$FNAME")
		HNAME="$DNAME/$(basename "$FNAME" ".md").html"
		INAME="$DNAME/$(basename "$FNAME" ".md").include"
		# NOTE: Need to bread crum the title for better search results ...
		TITLE=$(page_title_from_path "$FNAME")
		#echo "DEBUG title: $TITLE"; 
		echo "Writing $HNAME"
		pandoc --metadata title="${TITLE}" \
				-s --template=templates/page.html \
				$FNAME \
				-o $HNAME
		echo "Writing $INAME"
		pandoc --metadata title="${TITLE}" \
				-f markdown -t html5 \
				$FNAME \
				-o $INAME
	done
}

function make_pagefind() {
	CWD=$(pwd)
	cd htdocs && pagefind --verbose --exclude-selectors="nav,menu,header,footer" --output-path ./pagefind --site .
	cd "$CMD"
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
		html|static|root|recent|groups|people|group_pages|pagefind)
			cmd="make_${arg}"
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
make_static
# Build root folder contents.
make_root
# Build  recent folder
make_recent
# Build out the groups tree
make_groups
# Build out the people tree
make_people

# Find all the markdown files and render .html pages.
make_html

# Setup and run Pagefind
make_pagefind

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
