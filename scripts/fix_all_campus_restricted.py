#!/usr/bin/env python3
#
"""Update RDM records based on eprinte metadata"""

import sys
import os
import json, csv
import requests
from caltechdata_api import caltechdata_edit
from datetime import datetime
from urllib.parse import urlparse, unquote_plus
from subprocess import Popen, PIPE
from irdm import RdmUtil, eprint2rdm, fixup_record
from eprints_to_rdm import get_file_list

# Global list of DOIs aded during run, which are used to manage slow indexing times when records
# are updated
DOI_LIST = []


class WorkObject:
    """create a working object from dict for managing state in complete function."""

    def __init__(self, working_object):
        self.eprintid = working_object.get("eprintid", None)
        self.community_id = working_object.get("community_id", None)
        self.root_rdm_id = working_object.get("root_rdm_id", None)
        self.rdm_id = working_object.get("rdm_id", None)
        self.version_record = working_object.get("version_record", None)
        self.rec = working_object.get("record", None)
        self.restriction = working_object.get("restriction", None)
        self.version = working_object.get("version", "")
        self.publication_date = working_object.get("publication_date", None)

    def display(self):
        """return a JSON version of object contents."""
        return json.dumps(
            {
                "eprintid": self.eprintid,
                "community_id": self.community_id,
                "root_rdm_id": self.root_rdm_id,
                "rdm_id": self.rdm_id,
                "version": self.version,
                "restriction": self.restriction,
                "version_record": self.version_record,
                "publication_date": self.publication_date,
            }
        )

    def as_dict(self):
        """return object as a dict"""
        return {
            "eprintid": self.eprintid,
            "community_id": self.community_id,
            "root_rdm_id": self.root_rdm_id,
            "rdm_id": self.rdm_id,
            "version_record": self.version_record,
            "rec": self.rec,
            "restrictions": self.restriction,
        }


def check_environment():
    """Check to make sure all the environment variables have values and are avialable"""
    varnames = [
        "EPRINT_HOST",
        "EPRINT_USER",
        "EPRINT_PASSWORD",
        "EPRINT_DOC_PATH",
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


def get_restrictions(obj):
    """return any restrictins indicated in .access attribute"""
    restrict_record = False
    restrict_files = False
    if "access" in obj and "record" in obj["access"]:
        restrict_record = obj["access"]["record"] == "restricted"
    if "access" in obj and "files" in obj["access"]:
        restrict_files = obj["access"]["files"] == "restricted"
    return restrict_record, restrict_files


def set_restrictions(rdmutil, rdm_id, rec):
    """set the restrictions for a draft using rec"""
    restrict_record, restrict_files = get_restrictions(rec)
    if restrict_files:
        _, err = rdmutil.set_access(rdm_id, "files", "restricted")
        if err is not None:
            return err
    if restrict_record:
        _, err = rdmutil.set_access(rdm_id, "record", "restricted")
        if err is not None:
            return err
    return None


def pairtree(txt):
    """take a text string and generate a pairtree path from it."""
    return "/".join([txt[i : i + 2] for i in range(0, len(txt), 2)])


def update_record(config, rec, rdmutil, rdm_id):

    caltechdata_edit(
        rdm_id,
        metadata=rec,
        token=config["RDMTOK"],
        production=True,
        publish=True,
        authors=True,
    )


def fix_record(config, eprintid, rdm_id, restriction, reload=False):
    """Migrate a single record from EPrints to RDM using the document security model
    to guide versioning."""
    rdmutil = RdmUtil(config)
    eprint_host = config.get("EPRINT_HOST", None)

    token = config["RDMTOK"]
    headers = {"Authorization": f"Bearer {token}"}
    existing = requests.get(
        "https://authors.library.caltech.edu/api/records/" + rdm_id, headers=headers
    ).json()

    rec, err = eprint2rdm(eprintid)
    if err is not None:
        print(f"{eprintid}, None, failed ({eprintid}): eprint2rdm {eprintid}")
        sys.stdout.flush()
        return err  # sys.exit(1)

    file_list = get_file_list(config, eprintid, rec, restriction)

    file_description = ""
    file_types = set()
    campusonly_description = """<p><strong>Files attached to this record are
    restricted to users connected to the Caltech campus
    network:</strong></p><ul>"""
    campusonly_files = False

    if len(file_list) > 0:
        for file in file_list:
            filename = file.get("filename", None)
            description = file["description"]
            if restriction == "validuser":
                campusonly_description += f'     <li>{description} -  <a href="https://campus-restricted.library.caltech.edu/{rdm_id}/{filename}">{filename}</a></li>'
                campusonly_files = True
                file_types.add("campus only")
    campusonly_description += "</ul>"
    existing["metadata"]["version"] = "Campus-Access Only"
    if file_description != "" or campusonly_files:
        additional_descriptions = existing["metadata"].get(
            "additional_descriptions", []
        )
        for desc in additional_descriptions:
            if desc["type"]["id"] == "attached-files" or desc["type"]["id"] == "files":
                additional_descriptions.remove(desc)
        additional_descriptions.append(
                    {"type": {"id": "files"}, "description": campusonly_description}
                )
        existing["metadata"]["additional_descriptions"] = additional_descriptions

    update_record(config, existing, rdmutil, rdm_id)
    return None


def process_status(app_name, tot, cnt, started):
    if (cnt % 10) == 0:
        # calculate the duration in minutes.
        now = datetime.now()
        duration = (now - started).total_seconds()
        x = cnt / duration
        minutes_remaining = round((tot - cnt) * x)
        percent_completed = round((cnt / tot) * 100)
        if cnt == 0 or duration == 0:
            print(
                f'# {now.isoformat(" ", "seconds")} {app_name}: {cnt}/{tot} {percent_completed}%  eta: unknown',
                file=sys.stderr,
            )
        else:
            print(
                f'# {now.isoformat(" ", "seconds")} {app_name}: {cnt}/{tot} {percent_completed}%  eta: {minutes_remaining} minutes',
                file=sys.stderr,
            )


def display_status(app_name, cnt, started, completed):
    # calculate the duration in minutes.
    duration = round((completed - started).total_seconds() / 60) + 1
    x = round(cnt / duration)
    print(f"#    records processed: {cnt}", file=sys.stderr)
    print(f"#             duration: {duration} minutes", file=sys.stderr)
    print(f"#   records per minute: {x}")
    print(
        f'#   {app_name} started: {started.isoformat(" ", "seconds")}, completed: {completed.isoformat(" ", "seconds")}',
        file=sys.stderr,
    )


def process_document_and_eprintids(config, app_name, eprintids):
    """Process and array of EPrint Ids and migrate those records."""
    started = datetime.now()
    tot = len(eprintids)
    print(
        f'# Processing {tot} eprintids, started {started.isoformat(" ", "seconds")}',
        file=sys.stderr,
    )
    for i, _id in enumerate(eprintids):
        err = migrate_record(config, _id)
        if err is not None:
            print(f"error processing {_id}, row {i}, {err}", file=sys.stderr)
        process_status(app_name, tot, i, started)
    completed = datetime.now()
    display_status(app_name, len(eprintids), started, completed)
    return None


def get_eprint_ids():
    """review the command line parameters and get a list of eprint ids"""
    eprint_ids = []
    if len(sys.argv) > 1:
        arg = sys.argv[1]
        if os.path.exists(arg):
            with open(arg, encoding="utf-8") as _f:
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
    """main program entry point. I'm avoiding global scope on variables."""
    app_name = os.path.basename(sys.argv[0])
    config, is_ok = check_environment()
    if is_ok:
        with open("migrated_records.csv", "r") as f:
            token = config["RDMTOK"]
            headers = {"Authorization": f"Bearer {token}"}
            migrated = csv.DictReader(f)
            records = {}
            to_update = {}
            for row in migrated:
                if row['record_status'] == 'public':
                    if row['eprintid'] not in records:
                        records[row['eprintid']] = row['rdmid']
                    else:
                        api_url = f"https://authors.library.caltech.edu/api/records/"
                        response = requests.get(api_url + row['rdmid'],headers=headers)
                        if 'versions' in response.json():
                            latest = response.json()['versions']['is_latest']
                        else:
                            print(row['rdmid'])
                            print(response.json())
                        if latest==True:
                            to_update[records[row['eprintid']]]=row['eprintid']
                        else:
                            to_update[row['rdmid']]=row['eprintid']
        for rdm_id, eprintid in to_update.items():
            restriction = "validuser"
            err = fix_record(config, eprintid, rdm_id, restriction)
            if err is not None:
                print(f"Aborting {app_name}, {err}", file=sys.stderr)
                sys.exit(1)
            print(f"Updated {rdm_id}")
    else:
        print(f"Aborting {app_name}, environment not setup", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
