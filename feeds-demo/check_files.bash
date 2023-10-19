#!/bin/bash
#


#
# check_recent checks the recent folder for expected files
#
function check_recent() {
	# checked files for htdocs/recent/*.*
	cat <<EOF >check_list.tmp
htdocs/recent/article.json
htdocs/recent/article.md
htdocs/recent/audiovisual.json
htdocs/recent/audiovisual.md
htdocs/recent/book.json
htdocs/recent/book.md
htdocs/recent/book_section.json
htdocs/recent/book_section.md
htdocs/recent/collection.json
htdocs/recent/collection.md
htdocs/recent/combined.json
htdocs/recent/combined.md
htdocs/recent/combined_data.json
htdocs/recent/combined_data.md
htdocs/recent/conference_item.json
htdocs/recent/conference_item.md
htdocs/recent/data_object_types.json
htdocs/recent/data_pub_types.json
htdocs/recent/dataset.md
htdocs/recent/image.json
htdocs/recent/image.md
htdocs/recent/index.md
htdocs/recent/interactiveresource.json
htdocs/recent/interactiveresource.md
htdocs/recent/model.json
htdocs/recent/model.md
htdocs/recent/monograph.json
htdocs/recent/monograph.md
htdocs/recent/object_types.json
htdocs/recent/patent.json
htdocs/recent/patent.md
htdocs/recent/pub_types.json
htdocs/recent/software.json
htdocs/recent/software.md
htdocs/recent/teaching_resource.json
htdocs/recent/teaching_resource.md
htdocs/recent/text.json
htdocs/recent/text.md
htdocs/recent/thesis.json
htdocs/recent/thesis.md
htdocs/recent/video.json
htdocs/recent/video.md
htdocs/recent/workflow.json
htdocs/recent/workflow.md
EOF

	FAILED=""
	while read -r FNAME; do
		if [ ! -f "$FNAME" ]; then
			echo "Missing ${FNAME}"
			FAILED="true"
		fi
	done <check_list.tmp
	if [ "$FAILED" != "" ]; then
		echo "check_recent() failed"
	fi
}


#
# check_groups checks the groups folder for expected files
#
function check_groups() {
	# checked files for htdocs/groups/*.json
	cat <<EOF >check_list.tmp
htdocs/groups/group_list.json
htdocs/groups/index.json
EOF

	FAILED=""
	while read -r FNAME; do
		if [ ! -f "$FNAME" ]; then
			echo "Missing ${FNAME}"
			FAILED="true"
		fi
	done <check_list.tmp
	if [ "$FAILED" = "true" ]; then
		echo "check_groups() failed, aborting"
	fi
}

#
# check_root checks the htdocs folder's files
#
function check_root() {
	cat <<EOF >check_list.tmp
htdocs/caltechdata-grid.json
htdocs/caltechthesis-grid.json
htdocs/caltechauthors-grid.json
htdocs/favicon.ico
htdocs/index.md
htdocs/formats-and-extensions.md
htdocs/robots.txt
htdocs/about.md
htdocs/error.md
EOF

	FAILED=""
	while read -r FNAME; do
		if [ ! -f "$FNAME" ]; then
			echo "Missing ${FNAME}"
			FAILED="true"
		fi
	done <check_list.tmp
	if [ "$FAILED" != "" ]; then
		echo "check_root() failed, aborting"
	fi
}

#
# check_group_folders checks the groups folder for expected files
#
function check_group_folders() {
	# checked files for htdocs/groups/{GROUP_ID}/*.json
	# For specific goups, e.g. GALCIT, IPAQ, JCAP
	cat <<EOF >check_list.tmp
htdocs/groups/GALCIT/article.json
htdocs/groups/GALCIT/bachelors.json
htdocs/groups/GALCIT/book.json
htdocs/groups/GALCIT/book_section.json
htdocs/groups/GALCIT/combined.json
htdocs/groups/GALCIT/conference_item.json
htdocs/groups/GALCIT/engd.json
htdocs/groups/GALCIT/group.json
htdocs/groups/GALCIT/masters.json
htdocs/groups/GALCIT/monograph.json
htdocs/groups/GALCIT/patent.json
htdocs/groups/GALCIT/phd.json
htdocs/groups/GALCIT/senior_minor.json
htdocs/groups/GALCIT/teaching_resource.json
EOF

	FAILED=""
	while read -r FNAME; do
		if [ ! -f "$FNAME" ]; then
			echo "Missing ${FNAME}"
			FAILED="true"
		fi
	done <check_list.tmp
	if [ "$FAILED" != "" ]; then
		echo "check_group_folders() failed, aborting"
	fi
}



#
# Main
#
if [ "$1" != "" ]; then
        for arg in "$@"; do
                case "$arg" in
                *)
                        cmd="check_${arg}"
                        ;;
                esac
                echo "Running $cmd"
                if ! "${cmd}"; then
                        echo 'Something went wrong'
                        exit 10
                fi
        done
        exit 0
fi

# Check the root htdocs folder for expected files.
check_root
# Check htdocs/recent for at least the files from current production system
check_recent
# Check htdocs/groups/ check the groups main folder
check_groups
# Check a sample of htdocs/groups/*/*.json
check_group_folders

echo 'Success!'
