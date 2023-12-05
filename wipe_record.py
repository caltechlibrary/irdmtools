#!/usr/bin/env python3
#
"""eprints_to_rdm.py implements our Migration workflow for CaltechAUTHORS
from EPrints 3.3 to RDM 11."""

from distutils.file_util import move_file
import sys
import os
import json
import requests
from caltechdata_api import caltechdata_edit


def check_environment():
    """Check to make sure all the environment variables have values and are avialable"""
    varnames = [
        "RDMTOK",
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


def wipe_ind_record(config, rdm_id, message=None):

    with open(f"deleted_records/{rdm_id}.json", "w") as outfile:
        response = requests.get(
            f"https://authors.library.caltech.edu/api/records/{rdm_id}"
        )
        outfile.write(json.dumps(response.json(), indent=4))

    with open("blank.json", "r") as infile:
        metadata = json.load(infile)

        if message:
            metadata["metadata"]["description"] = message

        caltechdata_edit(
            rdm_id,
            metadata=metadata,
            token=config["RDMTOK"],
            production=True,
            publish=True,
            authors=True,
        )


def wipe_record(rdm_id):
    config, is_ok = check_environment()
    if is_ok:
        err = wipe_ind_record(config, rdm_id)
        if err is not None:
            print(f"Aborting update_from_eprints, {err}", file=sys.stderr)
            sys.exit(1)


#
# Wipe record
#
def main():
    """main program entry point. I'm avoiding global scope on variables."""
    config, is_ok = check_environment()
    if is_ok:
        rdm_id = sys.argv[1]
        message = sys.argv[2]
        err = wipe_ind_record(config, rdm_id, message)
        if err is not None:
            print(f"Aborting {app_name}, {err}", file=sys.stderr)
            sys.exit(1)
    else:
        print(f"Aborting {app_name}, environment not setup", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
