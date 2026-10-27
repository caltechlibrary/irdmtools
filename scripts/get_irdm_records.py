#!/usr/bin/env python3

import requests
import os
import sys
import json
import random
import argparse

import progressbar
from irdm import IRDM_Client
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

{app_name} [OPTIONS] C_NAME [RECORD_ID]

# DESCRIPTION

{app_name} retrieves a record via Invenio-RDM's REST API saving
the result in a dataset collection.

{app_name} relies on two environment variables 

- RDMTOK
- RDM_URL

to access a remote Invenio-RDM instance and migrate an
EPrint records saved in simple record format in a dataset
collection to the Invenio-RDM instance indicated by the 
environment.

C_NAME
: The name of the dataset collection to read records from

RECORD_ID
: This is the "key" portion of the Invenio-RDM URL (last part of the path) to the metadata record.

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


# EXAMPLES

The following example assumes that RDMTOK and RDM_URL have been
set in the shell's environment. It retrieves record "gt6ct-jxj21"
and save the result in irdm_authors.ds

~~~
    {app_name} irdm_authors.ds gt6ct-jxj21
~~~

Iterating over the whole collection. If you have a collection
of Invenio-RDM record keys retrieve all the records in bulk
saving them in a dataset collection called irdm_authors.ds

~~~
    {app_name} irdm_authors.ds < key_list.txt
~~~

''')

def app_setup(app_name):
    api_url = ''
    token = '' 
    community = ''
    c_name = None
    dsn = None
    keys = None
    # Get access token as environment variable
    if "RDM_URL" in os.environ:
        api_url = os.environ["RDM_URL"]
    if "RDMTOK" in os.environ:
        token = os.environ["RDMTOK"]
    if "RDM_COMMUNITY" in os.environ:
        community = os.environ['RDM_COMMUNITY']
    parser = argparse.ArgumentParser(
        prog = app_name,
        description="This program retreives records from an Invenio-RDM repository at Caltech Library and stores them in a dataset collection"
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
        "-token",
        nargs=1,
        default = "",
        help = "Set the access token for the API URL provided",
    )
    parser.add_argument(
        "-community",
        nargs=1,
        default = community,
        help = 'Set the community key for the repository',
    )
    parser.add_argument(
        "-dsn",
        nargs=1,
        default = "",
        help = 'if initializing a dataset colleciton use this dsn',
    )
    parser.add_argument(
        "c_name",
        nargs="?",
        default = "",
        help = 'set the name of the dataset collection to use',
    )
    parser.add_argument(
        "keys",
        nargs='*',
        default = [],
        help = 'retrieve records for each key and store in dataset collection',
    )
    args = parser.parse_args()
    if args.help:
        usage(app_name)
        sys.exit(0)
    if args.license:
        print(irdm.licenseText)
        sys.exit(0)
    if args.version:
        print(irdm.versionText)
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
    #if keys == None:
    #    print(f'missing key to migrate') # DEBUG
    return api_url, token, community, c_name, dsn, keys

#
# Main processing
#
app_name = os.path.basename(sys.argv[0])

production = False
publish = True
authors = True
community = "d0b46a93-0293-4689-a5c6-5ded7b7b4fd8"

api_url, token, community, c_name, dsn, keys = app_setup(app_name)

if not os.path.exists(c_name):
    err = dataset.init(c_name, dsn)
    if err != None and err != "":
        print(f'{c_name} count not be create, {err}')
        sys.exit(1)

client = IRDM_Client(
    api_url = api_url,
    repo = 'authors', 
    token = token, 
    community = community, 
    production = production,
)

if keys == None:
    keys = []
    for key in sys.stdin:
        key = key.strip()
        if key != '':
            keys.append(key)

if isinstance(keys, list) and len(keys) > 0:
    pid = os.getpid()
    tot = len(keys)
    bar = progressbar.ProgressBar(
        max_value=tot,
        widgets = [
            f'{app_name} {c_name} (pid:{pid})',
            ' ', progressbar.Counter(), f'/{tot}',
            ' ', progressbar.Percentage(),
            ' ', progressbar.AdaptiveETA(),
        ]
    )
    for i, key in enumerate(keys):
        if key != '':
            try:
                record = client.read(key)
            except Exception as err:
                print(f'failed to read {key}, {err}')
                continue
            if dataset.has_key(c_name, key) == False:
                err = dataset.create(c_name, key, record)
                if err != '':
                    print(f'{err}')
        if (i % 100) == 0:
            bar.update(i)
    bar.finish()
