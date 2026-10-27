#!/usr/bin/env python3

import requests
import os
import sys
import json
import random
import argparse

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

{app_name} [OPTIONS]

# DESCRIPTION

{app_name} uses Invenio-RDM's Elasticsearch to retrieve
a list of URLs for all records.

{app_name} relies on three environment variables 

- RDMTOK
- RDM_URL
- RDM_COMMUNITY

# POSITIONAL ARGUMENTS

c_name
: set the name of the dataset collection to use

keys
: retrieve records for each key and store in dataset collection

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
e.g. 'https://authors.caltechlibrary.dev'

-token TOKEN
: Set the access token for the API URL provided

-community COMMUNITY
: Set the community key for the repository

-dsn DSN
: If initializing a dataset colleciton use this dsn


# EXAMPLES

The following example assumes that RDMTOK and RDM_URL have been
set in the shell's environment. It retrieves a list of URLs
for all records in the repository. It writes the urls to
standard out.

~~~
    {app_name}
    {app_name} > key_list.txt
~~~

''')

def app_setup(app_name):
    api_url, token, community = None, None, None
    # Get access token as environment variable
    if "RDM_URL" in os.environ:
        api_url = os.environ["RDM_URL"]
    if "RDMTOK" in os.environ:
        token = os.environ["RDMTOK"]
    if "RDM_COMMUNITY" in os.environ:
        community = os.environ['RDM_COMMUNITY']
    if community == None:
        community = "d0b46a93-0293-4689-a5c6-5ded7b7b4fd8"

    parser = argparse.ArgumentParser(
        description="This program retreives ids from an Invenio-RDM repository at Caltech Library"
    )
    parser.add_argument(
        "-help",
        action = 'store_true',
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

    if args.api_url != "":
        if isinstance(args.api_url, list):
            api_url = args.api_url[0]
        else:
            api_url = args.api_url
        #print(f'DEBUG args.api_url {type(args.api_url)} -> {args.api_url}')
    if args.token != "":
        if isinstance(args.token, list):
            token = args.tokan[0]
        else:
            token = args.token
    if args.community != "":
        if isinstance(args.community, list):
            community = args.community[0]
        else:
            community = args.community
    if api_url == "":
        print(f'failed to read env RDM_URL or command line')
        sys.exit(1)
    if token == "":
        print(f'Failed to read env RDMTOK for access token')
        sys.exit(1)
    return api_url, token, community

def show_hits(response, remaining = -1):
    hits, total = response["hits"]["hits"], response["hits"]["total"]
    if remaining < 0:
        remaining = total - len(hits)
    for hit in hits:
        print(hit["id"])
        remaining -= 1
    return remaining

#
# Main processing
#
app_name = os.path.basename(sys.argv[0])

api_url, token, community = app_setup(app_name)

client = IRDM_Client(
    api_url = api_url,
    token = token, 
    community = community, 
)

size = 10
page_no = 1
#sort = { "updated" : "desc" }
sort = None # DEBUG need to figure out how sort works.
response = client.query("*", sort = sort, size = size, page = page_no)
total = response["hits"]["total"]
pages = total // size
remaining = show_hits(response, total)
while remaining > 0:
    page_no += 1
    response = client.query(
        "*",
        sort = sort,
        size = size,
        page = page_no,
    )
    remaining = show_hits(response, remaining)


