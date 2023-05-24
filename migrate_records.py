#!/usr/bin/env python3

import requests
import os
import sys
import json
import random
import argparse

import progressbar
from irdm import IRDM_Client, license_text, version_text, get_dict_path
from py_dataset import dataset

def usage(app_name):
    print(f'''---
title: "{app_name} (1) user manual"
pubDate: 2023-02-28
author: "R. S. Doiel"
---

# NAME

{app_name}

# SYNOPSIS

{app_name} [OPTIONS] C_NAME [KEY]

# DESCRIPTION

{app_name} relies on two environment variables

- RDMTOK
- RDM_URL

to access a remote Invenio-RDM instance and migrate an
EPrint records saved in simple record format in a dataset
collection to the Invenio-RDM instance indicated by the
environment.

C_NAME
: The name of the dataset collection to read records from

KEY
: The eprintid as string to retrieve the exported record

# OPTIONS

-h, --help
: show this help message and exit

-help
: display help details

-license
: display license details

-version
: display version

-api_url API_URL
: point to a specific Invenio-RDM api url,
e.g.  'https://authors.caltechlibrary.dev'

-token TOKEN
: Set the access token for the API URL provided

-community COMMUNITY
: Set the community key for the repository

-dsn DSN
: if initializing a dataset colleciton use this dsn

# EXAMPLES

The following example assumes that RDMTOK and RDM_URL have been
set in the shell's environment. It retrieves record 12211.

~~~
    {app_name} CaltechAUTHORS.ds 12211
~~~

Iterating over the whole collection.

~~~
    {app_name} CaltechAUTHORS.ds
~~~

''')

def app_setup(app_name):
    api_url = ''
    token = ''
    community = ''
    c_name = None
    dsn = None
    keys = None
    exit_on_error = False

    # Get access token as environment variable
    if "RDM_URL" in os.environ:
        api_url = os.environ["RDM_URL"]
    if "RDMTOK" in os.environ:
        token = os.environ["RDMTOK"]
    if "RDM_COMMUNITY" in os.environ:
        community = os.environ['RDM_COMMUNITY']
    parser = argparse.ArgumentParser(
        prog = app_name,
        description="""This program retreives records from an EPrints repository,
converts them into a dataset collection then sends them trys to import them into
Invenio-RDM repository. It is specific to Caltech Library and its' repositories.
"""
    )
    parser.add_argument(
        "-help",
        action = "store_true",
        help = "display help details",
    )
    parser.add_argument(
        "-license",
        action = 'store_true',
        help = "display license details",
    )
    parser.add_argument(
        "-version",
        action = 'store_true',
        help = "display version",
    )
    parser.add_argument(
        "-api_url",
        default = "",
        help= f"point to a specific Invenio-RDM api url, e.g. '{api_url}'",
    )
    parser.add_argument(
        "-exit_on_error",
        action = "store_true",
        help = "exit on error when trying to migration an error",
    )
    parser.add_argument(
        "-token",
        nargs=1,
        default = "",
        help = "Set the access token for the API URL provided",
    )
    parser.add_argument(
        "-community",
        nargs=1,
        default = community,
        help = "Set the community key for the repository",
    )
    parser.add_argument(
        "-eprintids",
        nargs=1,
        default = "",
        help = "Use a list of record ids in the given file",
    )
    parser.add_argument(
        "-dsn",
        nargs=1,
        default = "",
        help = "if initializing a dataset colleciton use this dsn",
    )
    parser.add_argument(
        "c_name",
        nargs="?",
        default = "",
        help = "set the name of the dataset collection to use",
    )
    parser.add_argument(
        "keys",
        nargs='*',
        default = [],
        help = "retrieve records for each key and store in dataset collection",
    )
    args = parser.parse_args()
    if args.help:
        usage(app_name)
        sys.exit(0)
    if args.license:
        print(license_text)
        sys.exit(0)
    if args.version:
        print(version_text)
        sys.exit(0)

    if args.c_name:
        if isinstance(args.c_name, list):
            c_name = args.c_name[0]
        else:
            c_name = args.c_name
    if args.keys:
        if isinstance(args.keys, list):
            keys = args.keys[:]
        else:
            keys = [args.keys]
    if args.eprintids:
        keys = []
        with open(args.eprintids[0]) as f:
            keys = f.readlines()
    if api_url == "":
        print(f'failed to read env RDM_URL')
        sys.exit(1)
    if token == "":
        print(f'Failed to read env RDMTOK for access token')
        sys.exit(1)
    if community == "":
        community = "d0b46a93-0293-4689-a5c6-5ded7b7b4fd8"
    if c_name == None:
        print(f'missing dataset collection name')
        sys.exit(1)
    if dsn == None:
        dsn = ''
    if args.exit_on_error:
        exit_on_error = True
    else:
        exit_on_error = False
    return api_url, token, community, c_name, dsn, keys, exit_on_error

def stop_on_exception(stop, key, err):
    s = f'{err}'
    key = f'{key}'.strip()
    if s.startswith('{'):
        err_obj = json.loads(s)
        msg = get_dict_path(err_obj, ["errors", 0, "messages", 0])
        if msg != None and msg.endswith(' already exists.'):
            print(f'     ➾ skipping eprintid {key}, already exists')
            return False
    if stop:
       return True
    return False

#
# Main processing
#
app_name = os.path.basename(sys.argv[0])
pid = os.getpid()
production = False
publish = True
authors = True
community = "d0b46a93-0293-4689-a5c6-5ded7b7b4fd8"

api_url, token, community, c_name, dsn, keys, exit_on_error = app_setup(app_name)

if keys == None:
    print(f'fetching all keys from {c_name}', file = sys.stderr)
    keys = dataset.keys(c_name)
if keys != None:
    client = IRDM_Client(api_url = api_url, token = token, community = community, production=production, publish = publish, schema = "")
    tot = len(keys)
    bar = progressbar.ProgressBar(
        maxvalue = tot,
        widgets = [
        f' {app_name} {c_name} (pid: {pid})',
        ' ', progressbar.Counter(), f'/{tot}',
        ' ', progressbar.Percentage(),
        ' ', progressbar.AdaptiveETA(),
        ]
    )
    for key in bar(keys):
        if tot < 120:
            print(f'fetching {key} ', end = '')
        data, err = dataset.read(c_name, key.strip())
        if err != '':
            print(f'error reading {key}, {err}')
            continue
        if data == None:
            print(f'no data for {key} in {c_name}')
            continue
        if 'tombstone' in data:
            print(f'    ⟹ skipping eprintid {key.strip()}, it is a tombstone record')
        else:
            try:
                response = client.create(
                    data
                )
                if tot < 120:
                    print(response)
            except Exception as err:
                if stop_on_exception(exit_on_error, key, err):
                    print(json.dumps(data, indent= 4))
                    print(f'c_name: {c_name}, key: {key}')
                    print(f'Exception: {err}')
                    sys.exit(1)
