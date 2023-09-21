#!/usr/bin/env python3
#
'''eprints_to_rdm.py implements our Migration workflow for CaltechAUTHORS
from EPrints 3.3 to RDM 11.'''

from distutils.file_util import move_file
import sys
import os
import json
from datetime import datetime
from urllib.parse import urlparse, unquote_plus
from subprocess import Popen, PIPE
from irdm import RdmUtil, eprint2rdm, fixup_record

class WorkObject:
    '''create a working object from dict for managing state in complete function.'''
    def __init__(self, working_object):
        self.eprintid = working_object.get('eprintid', None)
        self.community_id = working_object.get('community_id', None)
        self.root_rdm_id = working_object.get('root_rdm_id', None)
        self.rdm_id = working_object.get('rdm_id', None)
        self.version_record = working_object.get('version_record', None)
        self.rec = working_object.get('record', None)
        self.restriction = working_object.get('restriction', None)
        self.version = working_object.get('version', '')
        self.publication_date = working_object.get('publication_date', None)

    def display(self):
        '''return a JSON version of object contents.'''
        return json.dumps({
            'eprintid': self.eprintid,
            'community_id': self.community_id,
            'root_rdm_id': self.root_rdm_id,
            'rdm_id': self.rdm_id,
            'version': self.version,
            'restriction': self.restriction,
            'version_record': self.version_record,
            'publication_date': self.publication_date,
        })

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
        'EPRINT_HOST', 'EPRINT_USER', 'EPRINT_PASSWORD', 'EPRINT_DOC_PATH',
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

def set_restrictions(rdmutil, rdm_id, rec):
    '''set the restrictions for a draft using rec'''
    restrict_record, restrict_files = get_restrictions(rec)
    if restrict_files:
        _, err = rdmutil.set_access(rdm_id, 'files', 'restricted')
        if err is not None:
            return err
    if restrict_record:
        _, err = rdmutil.set_access(rdm_id, 'record', 'restricted')
        if err is not None:
            return err
    return None

def pairtree(txt):
    '''take a text string and generate a pairtree path from it.'''
    return '/'.join([txt[i:i+2] for i in range(0, len(txt), 2)])

def file_to_scp(config, eprintid, pos, source_name, target_name):
    '''turn an EPrint file URL into a scp command'''
    hostname = config.get('EPRINT_HOST', None)
    doc_path = config.get('EPRINT_DOC_PATH', None)
    if doc_path is None or hostname is None:
        print('failed: EPRINT_HOST and EPRINT_DOC_PATH not set'
              + ' for {target_name}')
        sys.exit(1)
    _eprint_id = eprintid.zfill(8)
    pos = f'{pos}'.zfill(2)
    host_path = os.path.join(
        doc_path, 'documents', 'disk0',
        pairtree(_eprint_id), pos, source_name
    )
    return [ 'scp', f"{hostname}:{host_path}", f"{target_name}" ]

def run_scp(cmd):
    '''take the scp command built iwth url_to_scp and run it.'''
    with Popen(cmd, stdout = PIPE, stderr = PIPE) as proc:
        out, err = proc.communicate()
        exit_code = proc.returncode
        if exit_code > 0:
            if isinstance(err, bytes):
                err = err.decode('utf-8').strip()
            print(f'exit code {exit_code}, {err}', file = sys.stderr)
            return err
        if isinstance(out, bytes):
            out = out.decode('utf-8').strip()
        if out is not None and out != "":
            print(f'out: {out}', file = sys.stderr)
        return None
    return f'''failed to run {' '.join(cmd)}'''

def get_file_list(config, eprintid, rec, security):
    '''given a record get the internal files as
    list of objects where each object is a filename and a path/url to the file.'''
    file_list = []
    if security == 'metadata_only':
        return file_list
    files = rec.get('files', {})
    entries = files.get('entries', [])
    content_mapping = {'accepted': 'Accepted Version', 'archival':
    'Archival Material', 'bibliography': 'Bibliography', 'coverimage':
    'Cover Image', 'discussion': 'Discussion', 'draft': 'Draft', 'erratum': 'Erratum',
    'inpress': 'In Press', 'metadata': 'Additional Metadata', 'other': 'Other',
    'permission': 'Release Permission', 'presentation': 'Presentation',
    'reprint': 'Reprint', 'submitted': 'Submitted', 'supplemental': 'Supplemental Material',
    'updated': 'Updated', 'waiver': 'OA Policy Waiver', 'published': 'Published'}
    for filename in entries:
        file = entries[filename]
        metadata = file.get('metadata', {})
        _security = metadata.get('security', None)
        pos = metadata.get('pos', 1)
        target_name = metadata.get('filename', filename)
        if '&' in target_name:
            target_name = target_name.replace('&', '_')
        if '%' in target_name:
            target_name = target_name.replace('%', '_')
        if '[' in target_name:
            target_name = target_name.replace('[', '_')
        if ']' in target_name:
            target_name = target_name.replace(']', '_')
        if ' ' in target_name:
            target_name = target_name.replace(' ', '_')
        if '(' in target_name:
            target_name = target_name.replace('(', '_')
        if ')' in target_name:
            target_name = target_name.replace(')', '_')
        source_name = filename
        if '&' in source_name:
            source_name = source_name.replace('&', '\&')
        if '[' in source_name:
            source_name = source_name.replace('[', '\[')
        if ']' in source_name:
            source_name = source_name.replace(']', '\]')
        if '(' in source_name:
            source_name = source_name.replace('(', '\(')
        if ')' in source_name:
            source_name = source_name.replace(')', '\)')
        if ' ' in source_name:
            source_name = source_name.replace(' ', '\ ')
        
        content = metadata.get('content', None)
        if content:
            content = content_mapping[content]
        format_des = metadata.get('format_des', None)
        if format_des:
            content = content + format_des
        if _security is not None and security == _security:
            file_url = file['file_id']
            cmd = file_to_scp(config, eprintid, pos, source_name, target_name)
            file_list.append({
                'filename': target_name,
                'file_url': file_url, 
                'cmd': cmd,
                'content': content
            })
    return file_list

def prune_attached_files_description(rec):
    metadata = rec.get('metadata', None)
    if metadata is not None:
        additional_descriptions = metadata.get('additional_descriptions', None)
        if additional_descriptions is not None:
            keep_descriptions = []
            for desc in additional_descriptions:
                desc_type = desc.get("type", "")
                if desc_type != "attached_files":
                    keep_descriptions.append(desc)
            metadata['additional_descriptions'] = keep_descriptions
        rec['metadata'] = metadata
    return rec

def update_record(config, rec, rdmutil, obj, internal_note):
    '''update draft record handling versioning if needed'''
    file_list = None
    err = None
    rec = prune_attached_files_description(rec)
    if obj.version_record:
        # Create the new version after saving the publication_date value
        obj.rdm_id, err = rdmutil.new_version(obj.root_rdm_id)
        if err is not None:
            print(f'failed ({obj.eprintid}), new_version {obj.root_rdm_id}')
            return obj.rdm_id, obj.version_record, err

    file_list = get_file_list(config, obj.eprintid, rec, obj.restriction)
    file_description = ''
    file_types = set()
    campusonly_description = 'The files for this record are restricted to users on the Caltech campus network:<p><ul>\n'
    campusonly_files = False
    if len(file_list) > 0:
        for file in file_list:
            filename = file.get('filename', None)
            content = file['content']
            if content:
                file_description += f'<p>{content} - <a href="/records/{obj.rdm_id}/files/{filename}?download=1">{filename}</a></p>'
                file_types.add(content)
            # Copy file with scp.
            cmd = file.get('cmd', None)
            err = run_scp(cmd)
            if err is not None:
                print(f'failed ({obj.eprintid}): {" ".join(cmd)}, {err}')
                continue # sys.exit(1)
            if obj.restriction == 'validuser':
                # NOTE: We want to put the files in place first, then update the draft.
                staging_dir = f's3_uploads/{obj.rdm_id}'
                dest = os.path.join(staging_dir, filename)
                if not os.path.exists(staging_dir):
                    os.makedirs(staging_dir, 0o775, exist_ok = True )
                move_file(filename, dest, verbose = True)
                campusonly_description += f'     <li><a href="https://campus-restricted.library.caltech.edu/{obj.rdm_id}/{filename}">{filename}</a></li>\n'
                campusonly_files = True
                file_types.add("campus only")
            else:
                _, err = rdmutil.set_files_enable(obj.rdm_id, True)
                if err is not None:
                    print(f'failed ({obj.eprintid}): set_files_enable {obj.rdm_id} true')
                    continue # sys.exit(1)
                _, err = rdmutil.upload_file(obj.rdm_id, filename)
                if err is not None:
                    print(f'failed ({obj.eprintid}): upload_file' +
                            f' {obj.rdm_id} {filename}, {err}')
                    continue # sys.exit(1)
                # NOTE: We want to remove the copied file if successfully uploaded.
                if os.path.exists(filename):
                    os.unlink(filename)
    if file_description != "" or campusonly_files:
        additional_descriptions = rec['metadata'].get('additional_descriptions', [])
        if file_description != "" and obj.restriction == "public":
            # Add file descriptions and version string
            additional_descriptions.append({'type': {'id':'attached-files'}, 'description': file_description})
        # Add campusonly descriptions
        if campusonly_files:
            additional_descriptions.append({'type': {'id':'files'}, 'description': campusonly_description})
        rec['metadata']['additional_descriptions'] = additional_descriptions
        rec['metadata']['version'] = ' + '.join(file_types)
        # Update the draft.
        #print(f'DEBUG dump record ->\n' + json.dumps(rec))
        rec, err = rdmutil.update_draft(obj.rdm_id, rec)
        if err is not None:
            print(f'failed ({obj.eprintid}): update_draft' +
                f' {obj.rdm_id} {rec}, {err}', file = sys.stderr)
            return obj.rdm_id, obj.version_record, err # sys.exit(1)
    else:
        # Set the version string.
        _, err = rdmutil.set_version(obj.rdm_id, obj.restriction)
        if err is not None:
            print(f'failed ({obj.eprintid}): set_version' +
                  f' {obj.rdm_id} {obj.restriction}, {err}')
            return obj.rdm_id, obj.version_record, err # sys.exit(1)

    restrict_record = restrict_files = 'public'
    if obj.restriction == 'internal':
        restrict_record = restrict_files = 'restricted'
    if obj.restriction == 'public' or obj.restriction == 'metadata_only':
        restrict_record = restrict_files = 'public'

    _, err = rdmutil.set_access(obj.rdm_id, 'files', restrict_files)
    if err is not None:
        print(f'failed ({obj.eprintid}), set access {obj.rdm_id} files {restrict_files}, {err}',
            file = sys.stderr)
    _, err = rdmutil.set_access(obj.rdm_id, 'record', restrict_record)
    if err is not None:
        print(f'failed ({obj.eprintid}), set access {obj.rdm_id} record {restrict_record}, {err}',
            file = sys.stderr)

    # Make sure .files.enabled is False for metadata only record(s) and versions
    if obj.restriction in [ "validuser", "metadata_only" ]:
        _, err = rdmutil.set_files_enable(obj.rdm_id, False)
        if err is not None:
            print(f'failed ({obj.eprintid}): set_files_enable {obj.rdm_id} false')
            return obj.rdm_id, obj.version_record, err # sys.exit(1)
    if obj.version_record:
        # Save version
        _, err = rdmutil.publish_version(obj.rdm_id, obj.restriction, obj.publication_date)
        if err is not None:
            print(f'failed ({obj.eprintid}/{obj.root_rdm_id})' +
                  f' publish_version {obj.rdm_id} {obj.restriction} {obj.publication_date}, {err}')
    else:
        # send to community and accept first draft
        _, err = rdmutil.send_to_community(obj.rdm_id, obj.community_id)
        if err is not None:
            print(f'failed ({obj.eprintid}): send_to_community' +
                  f' {obj.rdm_id} {obj.community_id}, {err}')
            return obj.rdm_id, obj.version_record, err # sys.exit(1)
        # NOTE: If internal_note is not empty then we need to append a comment to the review.
##         if internal_note != "":
##             _, err = rdmutil.review_comment(obj.rdm_id, internal_note)
##             if err is not None:
##                 print(f'failed ({obj.eprintid}): review_comment' +
##                     f' {obj.rdm_id} {obj.community_id}, {err}')
##                 return obj.rdm_id, obj.version_record, err # sys.exit(1)
        _, err = rdmutil.review_request(obj.rdm_id, 'accept', internal_note)
        if err is not None:
            print(f'failed ({obj.eprintid}): review_request' +
                f' {obj.rdm_id} accepted, {err}')
            return obj.rdm_id, obj.version_record, err # sys.exit(1)
    obj.version_record = True
    return obj.rdm_id, obj.version_record, err

def get_publication_date(rec):
    '''extract the publication date from the simplified EPRint record, default is today'''
    metadata = rec.get('metadata', None)
    if metadata is not None:
        publication_date = metadata.get('publication_date', None)
        if publication_date is not None:
            return publication_date
    today = datetime.now()
    return today.isoformat()

def get_restriction_list(rec):
    '''can the simplified record and figure out types of files and strictions are needed.'''
    files = rec.get('files', {})
    entries = files.get('entries', [])
    restriction_obj = {}
    for filename in entries:
        file = entries[filename]
        metadata = file.get('metadata', {})
        security = metadata.get('security', None)
        if security is not None:
            restriction_obj[security] = True
    restriction_list = []
    for restriction in [ 'internal', 'validuser', 'public' ]:
        if restriction in restriction_obj:
            restriction_list.append(restriction)
    if (len(restriction_list) == 0) or (not 'public' in restriction_list):
        restriction_list.append('metadata_only')
    return restriction_list

def migrate_record(config, eprintid, rdm_id):
    '''Migrate a single record from EPrints to RDM using the document security model
to guide versioning.'''
    rdmutil = RdmUtil(config)
    eprint_host = config.get('EPRINT_HOST', None)
    community_id = config.get('RDM_COMMUNITY_ID', None)
    if community_id is None or eprint_host is None:
        print(f'failed ({eprintid}): missing configuration, ' +
              'eprint host or rdm community id, aborting', file = sys.stderr)
        sys.exit(1)
    rec, err = eprint2rdm(eprintid)
    if err is not None:
        print(f'{eprintid}, None, failed ({eprintid}): eprint2rdm {eprintid}')
        sys.stdout.flush()
        return err # sys.exit(1)
    # Let's save our .custom_fields["caltech:internal_note"] value if it exists, per issue #16
    custom_fields = rec.get("custom_fields", {})
    internal_note = custom_fields.get("caltech:internal_note", "").strip('\n')

    # NOTE: fixup_record is destructive. This is the rare case of where we want to work
    # on a copy of the rec rather than modify rec!!!
    #print(json.dumps(rec))
    rec_copy, err = fixup_record(dict(rec),has_doi=True)
    if err is not None:
        print(f'{eprintid}, {rdm_id}, failed ({eprintid}): rdmutil new_record, fixup_record failed {err}')
    #print(json.dumps(rec_copy))
    root_rdm_id = rdm_id
    version_record = True
    publication_date = get_publication_date(rec)
    # Only rerunning public records
    for restriction in ['public']:
        obj = WorkObject({
            'community_id': community_id,
            'version': restriction,
            'publication_date': publication_date,
            'eprintid': eprintid,
            'root_rdm_id': root_rdm_id,
            'rdm_id': rdm_id,
            'version_record': version_record,
            'restriction': restriction,
        })
        rdm_id, version_record, err = update_record(config, rec, rdmutil, obj, internal_note)
        if err is not None:
            print(f'{obj.eprintid}, {rdm_id}, failed ({obj.eprintid}): update_record(config, rec, rdmutil, {obj.display()})')
            return err # sys.exit(1)
        print(f'{obj.eprintid}, {rdm_id}, {restriction}')
    print(f'{obj.eprintid}, {root_rdm_id}, migrated')
    with open('migrated_records.csv','a') as outfile:
        print(f"{obj.eprintid},{rdm_id},public",file=outfile)
    sys.stdout.flush()
    return None

def process_status(app_name, tot, cnt, started):
    if (cnt % 10) == 0:
        # calculate the duration in minutes.
        now = datetime.now()
        duration = (now - started).total_seconds()
        x = cnt / duration
        minutes_remaining = round((tot - cnt) * x)
        percent_completed = round((cnt/tot)*100)
        if cnt == 0 or duration == 0:
            print(f'# {now.isoformat(" ", "seconds")} {app_name}: {cnt}/{tot} {percent_completed}%  eta: unknown', file = sys.stderr)
        else:
            print(f'# {now.isoformat(" ", "seconds")} {app_name}: {cnt}/{tot} {percent_completed}%  eta: {minutes_remaining} minutes', file = sys.stderr)

def display_status(app_name, cnt, started, completed):
    # calculate the duration in minutes.
    duration = round((completed - started).total_seconds()/60) + 1
    x = round(cnt / duration)
    print(f'#    records processed: {cnt}', file = sys.stderr)
    print(f'#             duration: {duration} minutes', file = sys.stderr)
    print(f'#   records per minute: {x}')
    print(f'#   {app_name} started: {started.isoformat(" ", "seconds")}, completed: {completed.isoformat(" ", "seconds")}', file = sys.stderr)

def process_document_and_eprintids(config, app_name, eprint_id, rdm_id):
    '''Process eprints id of record with submitted RDM id.'''
    started = datetime.now()
    err = migrate_record(config, eprint_id, rdm_id)
    if err is not None:
        print(f'error processing {eprints_id}, rdm {rdm_id}, {err}', file = sys.stderr)
    return None


def reload_public_version(eprint_id,rdm_id):
    app_name = os.path.basename(sys.argv[0])
    config, is_ok = check_environment()
    if is_ok: 
        err = process_document_and_eprintids(config, app_name, eprint_id, rdm_id)
        if err is not None:
            print(f'Aborting {app_name}, {err}', file = sys.stderr)
            sys.exit(1)
    else:
        print(f'Aborting {app_name}, environment not setup', file = sys.stderr)

#
# Migrate a records using eprint2rdm, ./migrate_record.py and rdmutil.
#
def main():
    '''main program entry point. I'm avoiding global scope on variables.'''
    app_name = os.path.basename(sys.argv[0])
    eprint_id = sys.argv[1]
    rdm_id = sys.argv[2]
    config, is_ok = check_environment()
    if is_ok:
        err = process_document_and_eprintids(config, app_name, eprint_id, rdm_id)
        if err is not None:
            print(f'Aborting {app_name}, {err}', file = sys.stderr)
            sys.exit(1)
    else:
        print(f'Aborting {app_name}, environment not setup', file = sys.stderr)
        sys.exit(1)

if __name__ == '__main__':
    main()
