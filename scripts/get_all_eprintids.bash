#!/bin/bash
#
REPO_ID="$1"
mysql --batch --skip-column-names \
  --execute "SELECT eprintid FROM eprint WHERE eprint_status = 'archive' ORDER BY date_year, date_month, date_day, eprintid" "${REPO_ID}" \
  >all_ids.txt
