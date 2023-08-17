#!/usr/bin/env python3
'''Convert an eprints_to_rdm.py log into a set of rewrite instructions
for NginX'''

#
# Make NginX friendly redirects for migrated records in log
# from eprints_to_rdm.py.
#
import sys
import os
import csv


def process_log(log_name):
    '''Read the eprints_to_rdm.py log and generate a rewrite
    rule for each entry containing migrated'''
    with open(log_name, newline='', encoding ='utf-8') as csvfile:
        reader = csv.DictReader(csvfile, fieldnames = [ 'eprintid', 'rdm_id', 'status' ], restval = '')
        for row in reader:
            eprintid = row.get('eprintid', '').strip()
            rdm_id = row.get('rdm_id', '').strip()
            status = row.get('status', '').strip()
            if status == 'migrated':
                print(f'rewrite ^/{eprintid}$ /records/{rdm_id}')

def main():
    '''Main processing'''
    app_name = os.path.basename(sys.argv[0])
    if len(sys.argv) != 2:
        print(f'{app_name} requires a log filename to process.', file = sys.stderr)
        sys.exit(1)
    process_log(sys.argv[1])

if __name__ == '__main__':
    main()
