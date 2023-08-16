'''irdmtools.py wraps the Go based command line tools integrating
them in as Python object and functions.'''

from distutils.file_util import move_file
import os
import sys
import json
from subprocess import Popen, PIPE

class RdmUtil:
    '''RdmUtil creates a wrapper around the Go cli rdmutil for using in Python'''
    def __init__(self, config):
        '''grab environment needed by rdmutil'''
        self.rdm_url = config.get('RDM_URL', '')
        self.rdmtok = config.get('RDMTOK', '')
        self.rdm_community_id = config.get('RDM_COMMUNITY_ID', '')

    def new_record(self, draft):
        '''Run rdmutil new_record, return the generated record as dict along with error'''
        src = json.dumps(draft)
        if not isinstance(src, bytes):
            src = src.encode('utf-8')
        cmd = ["rdmutil", "new_record"]
        with Popen(cmd, stdin = PIPE, stdout = PIPE, stderr = PIPE) as proc:
            src, err = proc.communicate(input = src)
            exit_code = proc.returncode
            if exit_code > 0:
                return None, err
            if not isinstance(src, bytes):
                src = src.encode('utf-8')
            obj = json.loads(src)
            if 'id' in obj:
                return obj['id'], None
            return None, f'unable to extract record id for new record {obj}'
        return None, 'failed to run new_record'

    def get_record(self, rdm_id):
        '''get a published RDM record'''
        cmd = ["rdmutil", "get_record", rdm_id]
        with Popen(cmd, stdout = PIPE, stderr = PIPE) as proc:
            src, err = proc.communicate()
            exit_code = proc.returncode
            if exit_code > 0:
                return None, err
            if not isinstance(src, bytes):
                src = src.encode('utf-8')
            return json.loads(src), None
        return None, f'failed to get record {rdm_id}'

    def get_draft(self, rdm_id):
        '''retreives an existing draft record for the record id provided'''
        cmd = ["rdmutil", "get_draft", rdm_id]
        with Popen(cmd, stdout = PIPE, stderr = PIPE) as proc:
            src, err = proc.communicate()
            exit_code = proc.returncode
            if exit_code > 0:
                return None, err
            if not isinstance(src, bytes):
                src = src.encode('utf-8')
            return json.loads(src), None
        return None, f'failed to get draft {rdm_id}'

    def set_access(self, rdm_id, access_type, access_value):
        '''set the access of a draft'''
        cmd = ["rdmutil", "set_access", rdm_id, access_type, access_value]
        with Popen(cmd, stdout = PIPE, stderr = PIPE) as proc:
            src, err = proc.communicate()
            exit_code = proc.returncode
            if exit_code > 0:
                return None, err
            if not isinstance(src, bytes):
                src = src.encode('utf-8')
            return json.loads(src), None
        return None, f'failed to set_access {rdm_id} {access_type} {access_value}'
    
    def set_files_enable(self, rdm_id, enable):
        '''set files enable for draft'''
        enable_value = 'false'
        if enable:
            enable_value = 'true'
        cmd = ["rdmutil", "set_files_enable", rdm_id, enable_value]
        with Popen(cmd, stdout = PIPE, stderr = PIPE) as proc:
            src, err = proc.communicate()
            exit_code = proc.returncode
            if exit_code > 0:
                return None, err
            if not isinstance(src, bytes):
                src = src.encode('utf-8')
            return json.loads(src), None
        return None, f'failed to set_file_enable {rdm_id} {enable_value}'

    def upload_file(self, rdm_id, filename):
        cmd = ["rdmutil", "upload_files", rdm_id, filename ]
        with Popen(cmd, stdout = PIPE, stderr = PIPE) as proc:
            src, err = proc.communicate()
            exit_code = proc.returncode
            if exit_code > 0:
                return None, err
            if not isinstance(src, bytes):
                src = src.encode('utf-8')
            return json.loads(src), None
        return None, f'failed to upload_file {rdm_id} {filename}'

    def upload_files(self, rdm_id, filenames):
        cmd = ["rdmutil", "upload_files", rdm_id ]
        for filename in filenames:
            cmd.append(filename)
        with Popen(cmd, stdout = PIPE, stderr = PIPE) as proc:
            src, err = proc.communicate()
            exit_code = proc.returncode
            if exit_code > 0:
                return None, err
            if not isinstance(src, bytes):
                src = src.encode('utf-8')
            return json.loads(src), None
        return None, f'failed to upload_file {rdm_id} {filename}'

#     def upload_campusonly_file(self, rdm_id, filename):
#         '''Upload the campus only file to the S3 bucket,
#         then generate HTML to link to it in record'''
#         draft, err = self.get_draft(rdm_id)
#         if err is not None:
#             return None, err

#         # NOTE: We want to put the files in place first, then update the draft.
#         staging_dir = f's3_uploads/{rdm_id}'
#         dest = os.path.join(staging_dir, filename)
#         if not os.path.exists(staging_dir):
#             os.makedirs(staging_dir, 0o775, exist_ok = True )
#         move_file(filename, dest, verbose = True)
#         # NOTE: Now that we saved the file for S3 upload, we can update the metadata.additi
#         metadata = draft.get("metadata", {})
#         additional_descriptions = metadata.get("additional_descriptions", [])
#         print(f'DEBUG additional_descriptions -> {additional_descriptions}')
# # additional_descriptions: [
# # {
# # description: "The files for this record are restricted to users on the Caltech campus network:",
# # type: {
# # id: "files",
# # }
# # }
# # ],
# # The S3 URL can be embedded in the description, save the file to a subfolder for upload in S3:
#         campus_file = {
#             "description": 'The files for this record are restricted to users on the Caltech campus network:',
#             "type": {
#                 "id": "files",
#             },
#         }
#         additional_descriptions.append(campus_file)
#         print(f'DEBUG additional_descriptions -> {additional_descriptions}')
#         metadata["additional_descriptions"] = additional_descriptions
#         draft['metadata'] = metadata
#         print(f'DEBUG draft after additoinals\n{draft}\n')
#         sys.exit(1) # DEBUG
#         draft, err = self.update_draft(rdm_id, draft)
#         if err is not None:
#             return None, err
#         return draft, None

    def send_to_community(self, rdm_id, community_id = None):
        '''send a draft to the community'''
        if community_id is None:
            community_id = self.rdm_community_id
        cmd = ["rdmutil", "send_to_community", rdm_id, community_id]
        with Popen(cmd, stdout = PIPE, stderr = PIPE) as proc:
            src, err = proc.communicate()
            exit_code = proc.returncode
            if exit_code > 0:
                return None, err
            if not isinstance(src, bytes):
                src = src.encode('utf-8')
            return json.loads(src), None
        return None, f'failed to send_to_community draft {rdm_id} {community_id}'

    def review_request(self, rdm_id, decision):
        '''review a draft request'''
        cmd = ["rdmutil", "review_request", rdm_id, decision ]
        with Popen(cmd, stdout = PIPE, stderr = PIPE) as proc:
            src, err = proc.communicate()
            exit_code = proc.returncode
            if exit_code > 0:
                return None, err
            if not isinstance(src, bytes):
                src = src.encode('utf-8')
            return json.loads(src), None
        return None, f'failed to review_request {rdm_id} {decision}'

    def new_version(self, rdm_id):
        '''create a new version draft'''
        cmd = ["rdmutil", "new_version", rdm_id ]
        with Popen(cmd, stdout = PIPE, stderr = PIPE) as proc:
            src, err = proc.communicate()
            exit_code = proc.returncode
            if exit_code > 0:
                return None, err
            if not isinstance(src, bytes):
                src = src.encode('utf-8')
            obj = json.loads(src)
            new_rdm_id = obj.get('id', None)
            if new_rdm_id is not None:
                # Return the new version id.
                return new_rdm_id, None
            return None, f'failed to find id in new version of {rdm_id}'
        return None, f'failed to new_version {rdm_id}'

    def set_version(self, rdm_id, version):
        '''set .metadata.version string'''
        cmd = [ "rdmutil", "set_version", rdm_id, version]
        with Popen(cmd, stdout = PIPE, stderr = PIPE) as proc:
            src, err = proc.communicate()
            exit_code = proc.returncode
            if exit_code > 0:
                return None, err
            if not isinstance(src, bytes):
                src = src.encode('utf-8')
            return json.loads(src), None
        return None, f'failed to set_version {rdm_id} {version}'

    def publish_version(self, rdm_id, version = None, publication_date = None):
        '''create a new version draft'''
        cmd = ["rdmutil", "publish_version", rdm_id ]
        if version is not None:
            cmd.append(version)
        if publication_date is not None:
            cmd.append(publication_date)
        with Popen(cmd, stdout = PIPE, stderr = PIPE) as proc:
            src, err = proc.communicate()
            exit_code = proc.returncode
            if exit_code > 0:
                return None, err
            if not isinstance(src, bytes):
                src = src.encode('utf-8')
            obj = json.loads(src)
            if 'id' in obj:
                # Return the new version id.
                return obj['id'], None
            return rdm_id, None
        return None, f'failed to publish_version {rdm_id}'

    def update_draft(self, rdm_id, draft):
        src = json.dumps(draft)
        if not isinstance(src, bytes):
            src = src.encode('utf-8')
        cmd = ["rdmutil", "update_draft", rdm_id]
        with Popen(cmd, stdin = PIPE, stdout = PIPE, stderr = PIPE) as proc:
            src, err = proc.communicate(input = src)
            exit_code = proc.returncode
            if exit_code > 0:
                return None, err
            if src is None:
                return None, f'error, rdmutil update_draft failed to return object'
            if not isinstance(src, bytes):
                src = src.encode('utf-8')
            obj = json.loads(src)
            
            return obj, None
        return None, 'failed to run update_draft'


def eprint2rdm(eprint_id):
    '''Run the eprint2rdm command and get back a converted eprint record'''
    cmd = ["eprint2rdm"]
    cmd.append(eprint_id)
    with Popen(cmd, stdout = PIPE, stderr = PIPE) as proc:
        src, err = proc.communicate()
        exit_code = proc.returncode
        if exit_code > 0:
            return None, err
        if not isinstance(src, bytes):
            src = src.encode('utf-8')
        rec = json.loads(src)
        return rec, None
    return None, f'failed to run command eprint2rdm {eprint_id}.'
