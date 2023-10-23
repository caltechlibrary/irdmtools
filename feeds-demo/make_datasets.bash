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

FROM_DATE=$(reldate -- -4 day)

function harvest_rdm() {
	REPO="$1"
	FULL="$2"
	if [ -f "${REPO}.env" ]; then
		# shellcheck disable=SC1090
		. "${REPO}.env"
		if [ "${FULL}" = "full" ]; then
			echo "Harvesting $REPO all ids with rdmutil"
			KEY_LIST="${REPO}_all_ids.json"
			if ! rdmutil get_all_ids >"${KEY_LIST}"; then
				echo "   configuration ${REPO}.env may have problems"
				exit 11
			fi
		else
			echo "Harvesting updated since ${FROM_DATE} with rdmutil"
			KEY_LIST="${REPO}_modified.json"
			if ! rdmutil get_modified_ids "${FROM_DATE}" >"${KEY_LIST}"; then
				echo "    configuration ${REPO}.env may have problems or no new records were available"
				return
			fi
		fi
		echo "Harvesting records with rdm2eprint"
		if ! rdm2eprint -harvest "${C_NAME}" -ids "${KEY_LIST}"; then
			echo "   configuration ${REPO}.env may have problems or no records avaialble to harvest"
		fi
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
		ep3util harvest -modified "${FROM_DATE}"
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
	# NOTE: creating groups.ds from groups.csv is temporary, groups.ds should
	# originate and be managed from cold
	if [ -f groups.csv ]; then
		if [ ! -d groups.ds ]; then
			dataset init groups.ds "postgres://$USER@localhost/groups?sslmode=disable"
		fi	
		dsimporter -overwrite groups.ds groups.csv key
	else
		echo "failed to find groups.csv, skipping"
		return
	fi
}


function harvest_people() {
	# NOTE: creating people.ds from people.csv is temporary, people.ds should
	# originate and be managed from cold
	if [ -f people.csv ]; then
		if [ ! -d people.ds ]; then
			dataset init people.ds "postgres://$USER@localhost/people?sslmode=disable"
		fi
		dsimporter -overwrite people.ds people.csv cl_people_id
	else
		echo "failed to find people.csv, skipping"
	fi

	echo "Retrieving authors_count from authors.ds"
	dsquery -pretty -sql authors_count.sql authors.ds \
	>authors_count.json
	echo "Retrieving editor_count from authors.ds"
	dsquery -pretty -sql editor_count.sql authors.ds \
	>editor_count.json
	echo "Retrieving thesis_count from thesis.ds"
	dsquery -pretty -sql thesis_count.sql thesis.ds \
	>thesis_count.json
	echo "Retrieving advisor_count from thesis.ds"
	dsquery -pretty -sql advisor_count.sql thesis.ds \
	>thesis_count.json
	echo "Retrieving data_count from data.ds"
	dsquery -pretty -sql data_count.sql data.ds \
	>data_count.json
	echo "Mapping orcid to clpid for data.ds"
	dsquery -pretty -sql orcid_to_clpid.sql people.ds \
	>orcid_to_clpid.json
 	./update_people_counts.py
}

function harvest_rdm_fixup_data() {
	# generate a CSV file of RDM id and local_group values
	psql caltechdata -f fixup_data_local_group.sql >fixup_data_local_group.csv
	./fixup_data_local_group.py data.ds groups.csv fixup_data_local_group.csv
}

#
# Main processing steps
#
check_for_required_programs

# Harvest our RDM based repositories
FULL_HARVEST=""
# Check if we're doing a limited run
LIMITED=""
for ARG in "$@"; do
	case "${ARG}" in
		full)
		FULL_HARVEST="full"
		;;
		????-??-??)
		FROM_DATE="${ARG}"
		echo "Harvest will be from $FROM_DATE"
		;;
		authors)
		LIMITED="true"
		echo "Harvesting authors ${FULL_HARVEST}"
		 if ! harvest_rdm authors "${FULL_HARVEST}"; then
			echo "something went wrong"
			exit 64
		fi
		;;
		thesis)
		LIMITED="true"
		echo "Harvesting thesis ${FULL_HARVEST}"
		if ! harvest_eprints thesis "${FULL_HARVEST}"; then
			echo "something went wrong"
			exit 64
		fi
		;;
		data)
		LIMITED="true"
		echo "Harvesting data ${FULL_HARVEST}"
		if ! harvest_rdm data "${FULL_HARVEST}"; then
			echo "something went wrong"
			exit 64
		fi
		;;
		groups)
		LIMITED="true"
		echo "Harvesting groups ${FULL_HARVEST}"
		if ! harvest_groups; then
			echo "something went wrong"
			exit 64
		fi
		;;
		people)
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
# shellcheck disable=SC2043
for REPO in thesis; do
	echo "Harvesting ${REPO} EPrint repository"
	harvest_eprints "${REPO}" "${FULL_HARVEST}"
done

for REPO in data authors; do
	echo "Harvesting ${REPO} RDM repository"
	harvest_rdm "${REPO}" "${FULL_HARVEST}"
done

echo "Harvesting from groups.csv"
harvest_groups
echo "Harvesting from people.csv"
harvest_people

# Fixups can require use of the groups and people data harvested.
echo "Doing fix up for local groups (e.g. CaltechDATA)"
harvest_rdm_fixup_data
