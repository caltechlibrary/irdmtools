#!/usr/bin/env python3
#
"""eprints_to_rdm.py implements our Migration workflow for CaltechAUTHORS
from EPrints 3.3 to RDM 11."""

import sys
import os
import json
import requests
from caltechdata_api import caltechdata_edit
from ames.harvesters import get_restricted_records
from reload_public_version import reload_public_version


def check_environment():
    """Check to make sure all the environment variables have values and are avialable"""
    varnames = [
        "EPRINT_HOST",
        "EPRINT_USER",
        "EPRINT_PASSWORD",
        "EPRINT_DOC_PATH",
        "RDM_URL",
        "RDMTOK",
        "RDM_COMMUNITY_ID",
    ]
    config = {}
    is_ok = True
    for varname in varnames:
        val = os.getenv(varname, None)
        if val is None:
            print(f"missing enviroment {varname}", file=sys.stderr)
            is_ok = False
        else:
            config[varname] = val
    return config, is_ok


def main():
    """main program entry point. I'm avoiding global scope on variables."""
    app_name = os.path.basename(sys.argv[0])
    config, is_ok = check_environment()
    if is_ok:
        identifiers = {}
        records = get_restricted_records(config['RDMTOK'])
        for record in records:
            rdm_id = record["id"]
            metadata = record['metadata']
            if "identifiers" in metadata:
                for identifier in metadata['identifiers']:
                    if identifier['scheme'] == 'eprintid':
                        idv = identifier['identifier']
                        if idv in identifiers:
                            print(f"Duplicate {idv} {identifiers[idv]} {rdm_id}")
                        else:
                            identifiers[idv] = rdm_id
        for idv in identifiers:
            reload_public_version(idv,identifiers[idv])
            exit()
    else:
        print(f"Aborting {app_name}, environment not setup", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
