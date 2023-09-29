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
	if [ -f "${REPO}.env" ]; then
		# shellcheck disable=SC1090
		. "${REPO}.env"
	else
		echo "Skipping harvest for ${REPO}, no ${REPO}.env found"
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
		python csv_to_dataset.py groups.ds groups.csv
	else
		echo "failed to find groups.csv, skipping"
	fi
}

function harvest_people() {
	if [ -f people.csv ]; then
		if [ ! -d people.ds ]; then
			dataset init people.ds "postgres://$USER@localhost/people?sslmode=disable"
		fi
		python csv_to_dataset.py people.ds people.csv
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

echo "Harvesting from groups.csv"
harvest_groups
echo "Harvesting from people.csv"
harvest_people

echo "Harvesting RDM repositories"
for REPO in authors data; do
	harvest_rdm "${REPO}" "$FULL_HARVEST"
done

echo "Harvesting EPRint repositories"
for REPO in thesis; do
	harvest_eprints "${REPO}" "$FULL_HARVEST"
done