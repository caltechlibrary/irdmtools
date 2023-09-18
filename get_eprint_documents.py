#!/usr/bin/env python3

import sys
import os
from subprocess import Popen, PIPE

def pairtree_encode(eprintid):
    txt = eprintid.zfill(8)
    return '/'.join([txt[i:i+2] for i in range(0, len(txt), 2)])

def get_files(eprint_host, eprint_archives_path, repoid, eprintid):
    target = f'./problem_files/{eprintid}'
    os.makedirs(target, 0o775, exist_ok = True)
    docs = pairtree_encode(eprintid)
    cmd = [
        "scp",
        "-r",
        f'{eprint_host}:{eprint_archives_path}/{repoid}/documents/disk0/{docs}/*',
        f'{target}'
    ]
    print(' '.join(cmd))
    with Popen(cmd, stderr = PIPE) as proc:
        err = proc.communicate()
        exit_code = proc.returncode
        if exit_code > 0:
            if isinstance(err, bytes):
                err = err.decode('utf-8').strip()
            print(f'exit code {exit_code}, {err}', file = sys.stderr)
            return err
        return None
    return f'''failed to run {' '.join(cmd)}'''




def check_environment():
    cfg = {}
    is_ok = True
    # Pull the required attributes from the environment
    for envar in [ 'REPO_ID', 'EPRINT_ARCHIVES_PATH', 'EPRINT_HOST' ]:
        val = os.getenv(envar, None)
        if val is None:
            is_ok = False
            print(f"missing {envar} from environment", file=sys.stderr)
        else:
            cfg[envar] = val
    return cfg, is_ok

def main():
    app_name = os.path.basename(sys.argv[0])
    config, is_ok = check_environment()
    if not is_ok:
        print(f"Aborting {app_name}, environment not setup", file=sys.stderr)
        sys.exit(1)
    if len(sys.argv) > 2:
        print(f'usage: {app_name} EPRINT_ID')
        sys.exit(1)
    eprintid = sys.argv[1].strip()
    get_files(config['EPRINT_HOST'],
        config['EPRINT_ARCHIVES_PATH'],
        config['REPO_ID'], eprintid)

if __name__ == "__main__":
    main()
