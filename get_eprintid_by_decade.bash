#!/bin/bash
#
START_YEAR="$1"
END_YEAR="$2"
mysql --batch --skip-column-names --execute "SELECT eprintid FROM eprint WHERE date_year >= '$START_YEAR' AND date_year <= "$END_YEAR" AND eprint_status = 'archive' ORDER BY date_year, date_month, date_day, eprintid" "${REPO_ID}" >"migrate-ids-${START_YEAR}s.txt"
