#!/bin/bash
#

#
# Make dataset recreates the dataset collections for each of our repositories
# used for feeds v1.  For the RDM repositories (e.g. authors and data) the
# irdmtools converts the RDM records to a EPrints like JSON representation.
#
# The tooling relies on direct access to a snapshot of RDM's Postgres database(s).
# It uses ep3apid for harvestings EPrints data quickly for our thesis repositories.
#
# make_datasets.bash is the first stop in the processing pipeline needed to 
# create a feeds v1 compatible website.
#

function harvest_rdm() {
	REPO="$1"
	FULL="$2"
	if [ -f "${REPO}.env" ]; then
		# shellcheck disable=SC1090
		. "${REPO}.env"
		if [ "${FULL}" = "full" ]; then
			echo "Harvesting $REPO all ids with rdmutil"
			KEY_LIST="${REPO}_all_ids.json"
			rdmutil get_all_ids >"${KEY_LIST}"
		else
			echo "Harvesting last seven days ids with rdmutil"
			KEY_LIST="${REPO}_modified.json"
			rdmutil get_modified_ids "$(reldate -- -1 week)" >"${KEY_LIST}"
		fi
		echo "Harvesting records with rdm2eprint"
		rdm2eprint -harvest "${C_NAME}" -ids "${KEY_LIST}"
	else
		echo "Skipping harvest for ${REPO}, no ${REPO}.env found"
	fi
}

function harvest_eprints() {
	REPO="$1"
	FULL="$2"
	if [ -f "${REPO}.env" ]; then
		# shellcheck disable=SC1090
		. "${REPO}.env"
	else
		echo "Skipping harvest for ${REPO}, no ${REPO}.env found"
	fi
	if [ "${FULL}" = "full" ]; then
		ep3util harvest -all
	else
		ep3util harvest -modified "$(reldate -- -1 week)"
	fi
}

function check_for_required_programs() {
	IS_MISSING=""
	for CLI in rdmutil rdm2eprint; do
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

function harvest_groups() {
	if [ -f groups.csv ]; then
		if [ ! -d groups.ds ]; then
			dataset init groups.ds "postgres://$USER@localhost/groups?sslmode=disable"
	    fi	
		dsimporter -overwrite groups.ds groups.csv key
	else
		echo "failed to find groups.csv, skipping"
		return
	fi
    dsquery -csv 'id,type,pub_date,local_group' \
                -sql get_authors_pubs_by_group.sql authors.ds \
                >group_pubs.csv
	if [ ! -f group_pubs.csv ]; then
		echo "failed to create group_pubs.csv, skipping"
		return
	fi
	if [ ! -f group_pubs.ds ]; then
		dropdb --if-exists group_pubs
		createdb group_pubs
		dataset init group_pubs.ds "postgres://$USER@localhost/group_pubs?sslmode=disable"
	fi
	dsimporter -overwrite group_pubs.ds group_pubs.csv
	# Make group list
	dsquery authors.ds "SELECT src->>'local_group' AS local_group FROM authors GROUP BY src->>'local_group'"/
}

function harvest_people() {
	if [ -f people.csv ]; then
		if [ ! -d people.ds ]; then
			dataset init people.ds "postgres://$USER@localhost/people?sslmode=disable"
		fi
		dsimporter -overwrite people.ds people.csv cl_people_id
	else
		echo "failed to find people.csv, skipping"
	fi
}

#
# Main processing steps
#
check_for_required_programs

# Harvest our RDM based repositories
FULL_HARVEST=""
for ARG in "$@"; do
	case "${ARG}" in
	   "full")
	   FULL_HARVEST="full"
	   ;;
	   *)
	   ;;
	esac
done


# Check if we're doing a limited run
LIMITED=""
for ARG in "$@"; do
	case "${ARG}" in
	  	"authors")
		LIMITED="true"
		echo "Harvesting authors ${FULL_HARVEST}"
	 	if ! harvest_rdm authors "${FULL_HARVEST}"; then
			echo "something went wrong"
			exit 64
		fi
		;;
		"thesis")
		LIMITED="true"
		echo "Harvesting thesis ${FULL_HARVEST}"
		if ! harvest_eprints thesis "${FULL_HARVEST}"; then
			echo "something went wrong"
			exit 64
		fi
		;;
		"data")
		LIMITED="true"
		echo "Harvesting data ${FULL_HARVEST}"
		if ! harvest_rdm data "${FULL_HARVEST}"; then
			echo "something went wrong"
			exit 64
		fi
		;;
		"groups")
		LIMITED="true"
		echo "Harvesting groups ${FULL_HARVEST}"
		if ! harvest_groups; then
			echo "something went wrong"
			exit 64
		fi
		;;
		"people")
		LIMITED="true"
		echo "Harvesting people ${FULL_HARVEST}"
		if ! harvest_people; then
			echo "something went wrong"
			exit 64
		fi
		;;
	esac
done

if [ "$LIMITED" != "" ]; then
	echo "limited run completed"
	exit 0
fi

# We're doing a standard run, build the following in sequence
echo "Harvesting EPrint repositories"
for REPO in thesis; do
	harvest_eprints "${REPO}" "$FULL_HARVEST"
done

echo "Harvesting RDM repositories"
for REPO in data authors; do
	harvest_rdm "${REPO}" "$FULL_HARVEST"
done

echo "Harvesting from groups.csv"
harvest_groups
echo "Harvesting from people.csv"
harvest_people
