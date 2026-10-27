#!/bin/bash
#
YEAR="$1"
mysql --batch --skip-column-names --execute "SELECT eprintid FROM eprint WHERE date_year = '$YEAR' AND eprint_status = 'archive' ORDER BY date_year, date_month, date_day, eprintid" "${REPO_ID}"
