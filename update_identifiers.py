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
from datetime import datetime
from urllib.parse import urlparse, unquote_plus
from subprocess import Popen, PIPE
from irdm import RdmUtil, eprint2rdm, fixup_record


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
    """update draft record handling versioning if needed"""
    url = "https://authors.library.caltech.edu/"

    headers = {
        "Authorization": "Bearer %s" % config['RDMTOK'],
        "Content-type": "application/json",
    }
    # Get existing metadata
    existing = requests.get(
        url + "/api/records/" + rdm_id,
        headers=headers,
    )
    existing = existing.json()

    rec = rec['metadata']
    new_ids = []
    for idv in rec["identifiers"]:
        new_ids.append(idv["identifier"])

    identifiers = rec["identifiers"]
    if "identifiers" in existing:
        for idv in existing["identifiers"]:
            if idv["identifier"] not in new_ids:
                identifiers.append(idv)
    existing["metadata"]["identifiers"] = identifiers

    new_related = []
    for idv in rec["related_identifiers"]:
        new_related.append(idv["identifier"])

    related_identifiers = rec["related_identifiers"]
    if "related_identifiers" in existing:
        for idv in existing["related_identifiers"]:
            if idv["identifier"] not in new_related:
                related_identifiers.append(idv)
    existing["metadata"]["related_identifiers"] = related_identifiers
    #print(json.dumps(existing))

    caltechdata_edit(
        rdm_id,
        metadata=existing,
        token=config['RDMTOK'],
        production=True,
        publish=True,
        authors=True,
    )


def fix_record(config, eprintid, rdm_id):
    """Migrate a single record from EPrints to RDM using the document security model
    to guide versioning."""
    rdmutil = RdmUtil(config)
    eprint_host = config.get("EPRINT_HOST", None)
    community_id = config.get("RDM_COMMUNITY_ID", None)
    if community_id is None or eprint_host is None:
        print(
            f"failed ({eprintid}): missing configuration, "
            + "eprint host or rdm community id, aborting",
            file=sys.stderr,
        )
        sys.exit(1)
    rec, err = eprint2rdm(eprintid)
    if err is not None:
        print(f"{eprintid}, None, failed ({eprintid}): eprint2rdm {eprintid}")
        sys.stdout.flush()
        return err  # sys.exit(1)
    # Let's save our .custom_fields["caltech:internal_note"] value if it exists, per issue #16
    custom_fields = rec.get("custom_fields", {})
    internal_note = custom_fields.get("caltech:internal_note", "").strip("\n")

    # NOTE: fixup_record is destructive. This is the rare case of where we want to work
    # on a copy of the rec rather than modify rec!!!
    rec_copy, err = fixup_record(dict(rec))
    if err is not None:
        print(
            f"{eprintid}, {rdm_id}, failed ({eprintid}): rdmutil new_record, fixup_record failed {err}"
        )
    #print(json.dumps(rec_copy))
    update_record(config, rec, rdmutil, rdm_id)
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
        eprintid = sys.argv[1]
        rdm_id = sys.argv[2]
        err = fix_record(config, eprintid, rdm_id)
        # err = process_document_and_eprintids(config, app_name, get_eprint_ids())
        if err is not None:
            print(f"Aborting {app_name}, {err}", file=sys.stderr)
            sys.exit(1)
    else:
        print(f"Aborting {app_name}, environment not setup", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
