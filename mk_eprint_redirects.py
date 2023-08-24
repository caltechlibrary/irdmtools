#!/usr/bin/env python3
'''Convert an eprints_to_rdm.py log into a set of redirect map instructions
for NginX'''

#
# Make NginX friendly redirects for migrated records in log
# from eprints_to_rdm.py.
#
import sys
import os
import csv

from irdm import RdmUtil

def check_environment():
    '''Check to make sure all the environment variables have values and are avia
lable'''
    varnames = [
        'RDM_URL',
        'RDMTOK',
        'RDM_COMMUNITY_ID'
    ]
    config = {}
    is_ok = True
    for varname in varnames:
        val = os.getenv(varname, None)
        if val is None:
            print(f'missing enviroment {varname}', file = sys.stderr)
            is_ok = False
        else:
            config[varname] = val
    return config, is_ok

def rdm_is_accessible(rdmutil, rdm_id):
    '''Check the rdm id to make sure it is available to be mapped'''
    rec, err = rdmutil.get_record(rdm_id)
    if err is not None:
        return False
    access = rec.get('access', None)
    if access is None:
        return False
    record_access = access.get('record', '')
    if record_access != "public":
        return False
    return True

def process_log(rdmutil, log_name):
    '''Read the eprints_to_rdm.py log and generate a rewrite
    rule for each entry containing migrated'''
    with open(log_name, newline='', encoding ='utf-8') as csvfile:
        reader = csv.DictReader(csvfile, fieldnames = [ 'eprintid', 'rdm_id', 'status' ], restval = '')
        eprint_id_list = []
        duplicate_eprintid = []
        redirects = {}
        # Build our map of eprintid to rdm_id. We're favoring the
        # the first encountered.
        for row in reader:
            eprintid = row.get('eprintid', '').strip()
            rdm_id = row.get('rdm_id', '').strip()
            status = row.get('status', '').strip()
            if status == 'migrated' and rdm_is_accessible(rdmutil, rdm_id):
                if eprintid in eprint_id_list:
                    duplicate_eprintid.append(eprintid)
                    print(f'# duplicate, skipping {eprintid} -> {rdm_id}', file = sys.stderr)
                else:
                    eprint_id_list.append(eprintid)
                    print(f'    /{eprintid}      /records/{rdm_id}/latest;')

def main():
    '''Main processing'''
    app_name = os.path.basename(sys.argv[0])
    if len(sys.argv) != 2:
        print(f'{app_name} requires a log filename to process.', file = sys.stderr)
        sys.exit(1)
    config, is_ok = check_environment()
    if is_ok:
        rdmutil = RdmUtil(config)
        process_log(rdmutil, sys.argv[1])

if __name__ == '__main__':
    main()
