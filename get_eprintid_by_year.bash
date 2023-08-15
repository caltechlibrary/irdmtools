#!/bin/bash
#
YEAR="$1"
mysql --batch --skip-column-names --execute "SELECT eprintid FROM eprint WHERE date_year = '$YEAR' AND eprint_status = 'archive' ORDER BY eprintid" "${REPO_ID}"
