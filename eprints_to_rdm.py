#!/usr/bin/env python3
#
'''eprints_to_rdm.py implements our Migration workflow for CaltechAUTHORS
from EPrints 3.3 to RDM 11.'''

import sys
from subprocess import Popen, PIPE
import os
import csv
import json
import pymysql
import pymysql.cursors
import irdm


def check_environment():
    '''Check to make sure all the environment variables have values and are avialable'''
    varnames = [
        'REPO_ID',
        'EPRINT_HOST', 'EPRINT_USER', 'EPRINT_PASSWORD',
        'DB_NAME', 'DB_USER', 'DB_PASSWORD',
        'RDM_URL',
        'RDMTOK',
        'RDM_COMMUNITY_ID'
    ]
    config = {}
    is_ok = True
    for varname in varnames:
        val = os.getenv(varname)
        if val is None:
            print(f'missing enviroment {varname}', file = sys.stderr)
            is_ok = False
        else:
            config[varname] = val
    return config, is_ok

def generate_document_and_eprintids(config, f_name):
    '''Generate a tab delimited file of eprintid and document informaition'''
    documents = []
    conn = pymysql.connect(host = 'localhost',
                             user = config['DB_USER'],
                             password = config['DB_PASSWORD'],
                             database = config['DB_NAME'],
                             charset='utf8mb4',
                             cursorclass=pymysql.cursors.DictCursor)
    with conn:
        with conn.cursor() as cursor:
            sql = '''
SELECT eprint.eprintid AS eprintid,
       IFNULL(document.security, 'metadata_only') AS security
FROM eprint LEFT JOIN document ON (eprint.eprintid = document.eprintid)
WHERE eprint.eprint_status = 'archive' AND
      ((IFNULL(document.formatdesc, '') NOT LIKE 'Generate %') AND
       (IFNULL(document.formatdesc, '') NOT LIKE 'Thumbnail%'))
ORDER BY eprint.eprintid,
     CASE IFNULL(document.security, 'metadata_only')
         WHEN 'internal' THEN 1
         WHEN 'staffonly' THEN 1
         WHEN 'validuser' THEN 2
         WHEN 'public' THEN 4
         WHEN 'metadata_only' THEN 5
         ELSE 6
     END ASC
'''
            cursor.execute(sql)
            obj = {}
            last_eprint_id = None
            row = cursor.fetchone()
            while row is not None:
                eprint_id = row['eprintid']
                if last_eprint_id != eprint_id:
                    if 'eprintid' in obj:
                        print(f'DEBUG obj -> {obj}')
                        documents.append(obj)
                    obj = {
                        'eprintid': eprint_id,
                        'internal': (row['security'] == 'internal'),
                        'staffonly': (row['security'] == 'staffonly'),
                        'campus_only': (row['security'] == 'validuser'),
                        'public': (row['security'] == 'public'),
                        'metadata_only': (row['security'] == ''),
                    }
                    last_eprint_id = eprint_id
                else:
                    if row['security'] == 'internal':
                        obj['internal'] = True
                        obj['metadata_only'] = False
                    elif row['security'] == 'staffonly':
                        obj['staffonly'] = True
                        obj['metadata_only'] = False
                    elif row['security'] == 'validuser':
                        obj['campus_only'] = True
                        obj['metadata_only'] = False
                    elif row['security'] == 'public':
                        obj['public'] = True
                        obj['metadata_only'] = False
                row = cursor.fetchone()
    # After collecting eprintids and document types, write our CSV file.
    if len(documents) == 0:
        print('nothing found to process', file = sys.stderr)
        return False
    with open(f_name, 'w', newline = '', encoding = 'utf-8') as csvfile:
        fieldnames = [
            'eprintid', 'metadata_only', 'internal',
            'staffonly', 'campus_only', 'public'
        ]
        _w = csv.DictWriter(csvfile, fieldnames = fieldnames)
        _w.writeheader()
        for obj in documents:
            _w.writerow(obj)
    return True

def eprint2rdm(eprint_host, eprint_id, doc_files = None):
    '''Run the eprint2rdm command and get back a converted eprint record'''
    cmd = ["eprint2rdm"]
    if doc_files is not None:
        cmd.append('-doc-files')
        cmd.append(doc_files)
    cmd.append(eprint_host)
    cmd.append(eprint_id)
    print(f"DEBUG cmd {cmd}")
    with Popen(cmd, stdout = PIPE, stderr = PIPE) as proc:
        src, err = proc.communicate()
        exit_code = proc.returncode
        if exit_code > 0:
            print(f'error {err}', file=sys.stderr)
            return None, err
        if not isinstance(src, bytes):
            src = src.encode('utf-8')
        rec = json.loads(src)
        return rec, None
    return None, 'failed to run command.'


def migrate_record(config, obj):
    '''Migrate a single record from EPrints to RDM using the document security model
to guide versioning.

This processing approach per Tom's slack description:

    Yes. Here's a more detailed possible logic: If a eprints record has internal
    or staffonly files, create a fully restricted record with those files.
    If the eprints record also has public files, create an open version
    with the public files. Otherwise, create a public metadata only version
    with the metadata
'''
    eprint_id = config['eprintid']
    eprint_host = config['eprint_host']
#FIXME: Need to take the eprint2rdm records and follow Tom's algorythm.
    return 'migrate_record() not implemented'

def process_document_and_eprintids(config, f_name):
    '''Reads in the tab separate value file and runs the
    migration using those documents/eprintids listed'''
    is_ok = True
    with open(f_name, newline = '', encoding = 'utf-8') as csvfile:
        _r = csv.DictReader(csvfile)
        for i, obj in enumerate(_r):
            err = migrate_record(cfg, obj)
            if err is not None:
                print(f'error processing {f_name}, row {i}, {err}')
                is_ok = False
            if i > 5: # DEBUG
                sys.exit(0) # DEBUG
    return is_ok

#
# Migrate a records using eprint2rdm, ./migrate_record.py and rdmutil.
#
if __name__ == '__main__':
    app_name = os.path.basename(sys.argv[0])
    DOCUMENT_AND_EPRINTIDS = 'document_and_eprintids.csv'
    cfg, ok = check_environment()
    if not ok:
        print(f'Aborting {app_name}, environment not setup')
        sys.exit(1)
    #if not os.path.exists(DOCUMENT_AND_EPRINTIDS):
    if generate_document_and_eprintids(cfg, DOCUMENT_AND_EPRINTIDS):
        print(f'{DOCUMENT_AND_EPRINTIDS} generated')
    else:
        print(f'{DOCUMENT_AND_EPRINTIDS} not generated, aborting')
        sys.exit(1)
    if not process_document_and_eprintids(cfg, DOCUMENT_AND_EPRINTIDS):
        print(f'Aborting {app_name}', file = sys.stderr)
        sys.exit(1)
