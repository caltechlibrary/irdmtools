#!/usr/bin/env python3
#
'''eprints_to_rdm.py implements our Migration workflow for CaltechAUTHORS
from EPrints 3.3 to RDM 11.'''

import sys
import os
import json
from urllib.parse import urlparse
from subprocess import Popen, PIPE
from irdm import RdmUtil, eprint2rdm, fixup_record

class WorkObject:
    '''create a working object from dict for managing state in complete function.'''
    def __init__(self, working_object):
        self.eprintid = working_object.get('eprintid', None)
        self.community_id = working_object.get('rdm_community_id', None)
        self.root_rdm_id = working_object.get('root_rdm_id', None)
        self.rdm_id = working_object.get('rdm_id', None)
        self.version_record = working_object.get('version_record', None)
        self.rec = working_object.get('record', None)
        self.restriction = working_object.get('restriction', None)

    def display(self):
        '''return a JSON version of object contents.'''
        return json.dumps(self)

    def as_dict(self):
        '''return object as a dict'''
        return {
            'eprintid': self.eprintid,
            'community_id': self.community_id,
            'root_rdm_id': self.root_rdm_id,
            'rdm_id': self.rdm_id,
            'version_record': self.version_record,
            'rec': self.rec,
            'restrictions': self.restriction,
        }

def check_environment():
    '''Check to make sure all the environment variables have values and are avialable'''
    varnames = [
        'REPO_ID',
        'EPRINT_HOST', 'EPRINT_USER', 'EPRINT_PASSWORD', 'EPRINT_DOC_PATH',
        'DB_NAME', 'DB_USER', 'DB_PASSWORD',
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

def get_restrictions(obj):
    '''return any restrictins indicated in .access attribute'''
    restrict_record = False
    restrict_files = False
    if 'access' in obj and 'record' in obj['access']:
        restrict_record = obj['access']['record'] == 'restricted'
    if 'access' in obj and 'files' in obj['access']:
        restrict_files = obj['access']['files'] == 'restricted'
    return restrict_record, restrict_files

def is_metadata_only(rec):
    '''look at the files attribute and determine if this is a metadata only record'''
    if 'files' in rec and 'enable' in rec['files']:
        return rec['files']['enable'] is True
    return False

def set_restrictions(rdmutil, rdm_id, rec):
    '''set the restrictions for a draft using rec'''
    restrict_record, restrict_files = get_restrictions(rec)
    if restrict_files:
        err = rdmutil.set_access(rdm_id, 'files', 'restricted')
        if err is not None:
            return err
    if restrict_record:
        err = rdmutil.set_access(rdm_id, 'record', 'restricted')
        if err is not None:
            return err
    return None

def pairtree(txt):
    '''take a text string and generate a pairtree path from it.'''
    return '/'.join([txt[i:i+2] for i in range(0, len(txt), 2)])

def url_to_scp(config, url, target_filename):
    '''turn an EPrint file URL into a scp command'''
    doc_path = config.get('EPRINT_DOC_PATH', '')
    _u = urlparse(url)
    hostname = _u.hostname
    parts = _u.path.split('/')[1:]
    eprintid = parts[0].zfill(8)
    version = parts[1].zfill(2)
    filename = parts[2]
    host_path =  '/'.join([doc_path, 'documents', 'disk0', pairtree(eprintid), version, filename])
    return [ 'scp', f'{hostname}:{host_path}', target_filename]

def run_scp(cmd):
    '''take the scp command built iwth url_to_scp and run it.'''
    with Popen(cmd, stdout = PIPE, stderr = PIPE) as proc:
        out, err = proc.communicate()
        exit_code = proc.returncode
        if exit_code > 0:
            print(f'exit code {exit_code}, {err}', file = sys.stderr)
            return err
        if out is not None:
            print(f'out: {out}')
        return None
    return f'''failed to run {' '.join(cmd)}'''

def get_file_list(config, rec, security):
    '''given a record get the internal files as
    list of objects where each object is a filename and a path/url to the file.'''
    file_list = []
    if 'files' in rec:
        files = rec.get('files')
        if files is not None:
            entries = files.get('entries')
            if entries is not None:
                for filename in entries:
                    print(f'DEBUG filename {filename}')
                    file = files.get(filename, None)
                    print(f'DEBUG  file -> {file}')
                    if file is not None:
                        metadata = file.get('metadata', None)
                        if metadata is not None:
                            _security = metadata.get('security', None)
                        if _security is not None and security == _security:
                            file_url = file['file_id']
                            cmd = url_to_scp(config, file_url, filename)
                            file_list.append({'filename': filename, 'file_url': file_url, 'cmd': cmd})
    if len(file_list) == 0:
        return None
    return file_list

def update_record(config, rdmutil, obj):
    '''update draft record handling versioning if needed'''
    file_list = None
    err = None
    if obj.restriction in [ 'internal', 'staffonly' ]:
        restrict_record = restrict_files = 'restricted'
    else:
        restrict_record = restrict_files = 'public'

    if obj.version_record:
        # We need to create a new draft version to work with.
        draft, err = rdmutil.get_draft(obj.rdm_id)
        if err is not None:
            print(f'failed ({obj.eprintid}): get_draft {obj.rdm_id}, {err}', file = sys.stderr)
            sys.exit(1) # DEBUG, maybe should be return!
        publication_date = draft['metadata']['publication_date']
        obj.rdm_id, err = rdmutil.new_version(obj.root_rdm_id)
        if err is not None:
            print(f'failed ({obj.eprintid}), new_version {obj.root_rdm_id}', file = sys.stderr)
            return err
        draft, err = rdmutil.get_draft(obj.rdm_id)
        if err is not None:
            print(f'failed ({obj.eprintid}, new version): get_draft {obj.rdm_id}, {err}',
                  file = sys.stderr)
            sys.exit(1) # DEBUG, maybe should be return!
        # Need to re-populate the publication date of new version.
        draft['metadata']['publication_date'] = publication_date
        # Need to give it a version label.
        draft['metadata']['version'] = obj.restriction
        # Save the updated draft before proceeding.
        err = rdmutil.update_draft(rdmutil, draft)
        if err is not None:
            print(f'failed ({obj.eprintid}, new version): update_draft {rdm_id} data, {err}',
                  file = sys.stderr)

    err = rdmutil.set_access(obj.rdm_id, 'files', restrict_files)
    if err is not None:
        print(f'failed ({obj.eprintid}), set access {obj.rdm_id} files {restrict_files}, {err}',
              file = sys.stderr)
    err = rdmutil.set_access(obj.rdm_id, 'record', restrict_record)
    if err is not None:
        print(f'failed ({obj.eprintid}), set access {obj.rdm_id} record {restrict_record}, {err}',
              file = sys.stderr)

    file_list = get_file_list(config, obj.rec, obj.restriction)
    if file_list is not None:
        for file in file_list:
            # Copy file with scp.
            cmd = file['cmd']
            filename = file['filename']
            err = run_scp(cmd)
            if err is not None:
                print(f'failed ({obj.eprintid}): scp {filename}, {err}', file = sys.stderr)
                break
            if obj.restrictions == 'validuser':
                err = rdmutil.upload_campusonly_file(obj.rdm_id, filename)
                if err is not None:
                    print(f'failed ({obj.eprintid}): update_campusonly_file' +
                          ' {rdm_id} {filename}, {err}', file = sys.stderr)
            else:
                err = rdmutil.upload_files(obj.rdm_id, filename)
                if err is not None:
                    print(f'failed ({obj.eprintid}): update_file' +
                          '{obj.rdm_id} {filename}, {err}', file = sys.stderr)
    if obj.version_record:
        # Save version
        err = rdmutil.publish_version(obj.rdm_id, obj.community_id)
    else:
        # send to community and accept first draft
        err = rdmutil.send_to_community(obj.rdm_id, obj.community_id)
        if err is not None:
            print(f'failed ({obj.eprintid}): send_to_community' +
                  ' {obj.rdm_id} {obj.community_id}, {err}', file = sys.stderr)
        else:
            err = rdmutil.review_request(obj.rdm_id, 'accept')
            if err is not None:
                print(f'failed ({obj.eprintid}): review_request' +
                      ' {obj.rdm_id} accepted, {err}', file = sys.stderr)
    obj.version_record = True
    return obj.rdm_id, obj.version_record, err

def migrate_record(config, eprintid):
    '''Migrate a single record from EPrints to RDM using the document security model
to guide versioning.'''
    rdmutil = RdmUtil(config)
    eprint_host = config.get('EPRINT_HOST', None)
    rdm_community_id = config.get('RDM_COMMUNITY_ID', None)
    if rdm_community_id is None or eprint_host is None:
        print(f'failed ({eprintid}): missing configuration, ' +
              'eprint host or rdm community id, aborting', file = sys.stderr)
        sys.exit(1)
    rdm_id = None
    root_rdm_id = None
    rec, err = eprint2rdm(eprint_host, eprintid)
    if err is None:
        rdm_id, err  = rdmutil.new_record(fixup_record(rec))
    if err is None:
        print(f'Creating RDM record {rdm_id} from eprint {eprintid} as draft')
        root_rdm_id = rdm_id

    version_record = False
    for restriction in [ 'internal', 'staffonly', 'validuser', 'public' ]:
        obj = WorkObject({
            'community_id': rdm_community_id,
            'eprintid': eprintid,
            'root_rdm_id': root_rdm_id,
            'rdm_id': rdm_id,
            'version_record': version_record,
            'record': rec,
            'restriction': restriction,
        })
        rdm_id, version_record, err = update_record(config, rdmutil, obj)

    if err is None:
        err = 'migrate_record() not fully implemented'
    return err

def process_document_and_eprintids(config, eprintids):
    '''Process and array of EPrint Ids and migrate those records.'''
    for i, _id in enumerate(eprintids):
        err = migrate_record(config, _id)
        if err is not None:
            print(f'error processing {_id}, row {i}, {err}')
            return err
        if i > 5: # DEBUG
            sys.exit(0) # DEBUG
    return None

def get_eprint_ids():
    '''review the command line parameters and get a list of eprint ids'''
    eprint_ids = []
    if len(sys.argv):
        arg = sys.argv[1]
        if os.path.exists(arg):
            with open(arg, encoding = 'utf-8') as _f:
                for line in _f:
                    eprint_ids.append(line.strip())
        elif arg.isdigit():
            args = sys.argv[:]
            for eprint_id in args[1:]:
                eprint_ids.append(eprint_id.strip())
    return eprint_ids

#
# Migrate a records using eprint2rdm, ./migrate_record.py and rdmutil.
#
def main():
    '''main program entry point. I'm avoiding global scope on variables.'''
    app_name = os.path.basename(sys.argv[0])
    config, is_ok = check_environment()
    if is_ok:
        err = process_document_and_eprintids(config, get_eprint_ids())
        if err is not None:
            print(f'Aborting {app_name}, {err}', file = sys.stderr)
            sys.exit(1)
    else:
        print(f'Aborting {app_name}, environment not setup')
        sys.exit(1)

if __name__ == '__main__':
    main()
