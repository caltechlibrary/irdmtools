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
from eprints_to_rdm import get_file_list


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
    community_id = config.get("RDM_COMMUNITY_ID", None)

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
    rec_copy, err = fixup_record(dict(rec), reload, token=config["RDMTOK"])
    if err is not None:
        print(
            f"{eprintid}, {rdm_id}, failed ({eprintid}): rdmutil new_record, fixup_record failed {err}"
        )

    file_list = get_file_list(config, eprintid, rec, restriction)
    file_description = ""
    file_types = set()
    campusonly_description = "The files for this record are restricted to users on the Caltech campus network:<p><ul>\n"
    campusonly_files = False
    if len(file_list) > 0:
        for file in file_list:
            filename = file.get("filename", None)
            content = file["content"]
            if content:
                file_description += f'<p>{content} - <a href="/records/{rdm_id}/files/{filename}?download=1">{filename}</a></p>'
                file_types.add(content)
            if restriction == "validuser":
                # NOTE: We want to put the files in place first, then update the draft.
                campusonly_description += f'     <li><a href="https://campus-restricted.library.caltech.edu/{rdm_id}/{filename}">{filename}</a></li>\n'
                campusonly_files = True
                file_types.add("campus only")
    if file_description != "" or campusonly_files:
        additional_descriptions = rec["metadata"].get("additional_descriptions", [])
        if file_description != "" and restriction == "public":
            # Add file descriptions and version string
            additional_descriptions.append(
                {"type": {"id": "attached-files"}, "description": file_description}
            )
        # Add campusonly descriptions
        if campusonly_files:
            additional_descriptions.append(
                {"type": {"id": "files"}, "description": campusonly_description}
            )
        rec["metadata"]["additional_descriptions"] = additional_descriptions
        rec["metadata"]["version"] = " + ".join(file_types)

    # We want to use the access currently in RDM
    rec.pop("access")

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


def update_from_eprints(eprintid, rdm_id, reload=False):
    config, is_ok = check_environment()
    if is_ok:
        err = fix_record(config, eprintid, rdm_id, reload)
        if err is not None:
            print(f"Aborting update_from_eprints, {err}", file=sys.stderr)
            sys.exit(1)


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
        restriction = sys.argv[3]
        err = fix_record(config, eprintid, rdm_id, restriction)
        if err is not None:
            print(f"Aborting {app_name}, {err}", file=sys.stderr)
            sys.exit(1)
    else:
        print(f"Aborting {app_name}, environment not setup", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
