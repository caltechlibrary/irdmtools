#!/usr/bin/env python3

import sys
import os
import csv


def main(argv):
    if len(argv) == 1:
        csv_name = "migrated_records.csv"
    else:
        csv_name = argv[1]
    with open(csv_name, newline = "", encoding ="utf-8") as csvfile:
        rows = csv.DictReader(csvfile, fieldnames = [ "eprintid", "rdm_id", "status" ])
        last_eprint_id = ""
        last_rdm_id = ""
        for row in rows:
            eprintid = row.get('eprintid', '')
            rdm_id = row.get('rdm_id', '')
            if eprintid == last_eprint_id:
                print(f'{eprintid}, {rdm_id}, {last_rdm_id}, duplicate')
            last_eprint_id = eprintid
            last_rdm_id = rdm_id

if __name__ == "__main__":
    main(sys.argv)
