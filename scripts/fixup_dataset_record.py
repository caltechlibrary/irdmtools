#!/usr/bin/env python3

import os
import sys
import json
import argparse

from irdm import fixup_record
from py_dataset import dataset

def usage(app_name):
    print(f'''---
title: "{app_name} (1) user manual"
pubDate: 2023-03-03
author: "R. S. Doiel"
---

# NAME

{app_name}

# SYNOPSIS

{app_name} [OPTIONS] C_NAME RECORD_ID

# DESCRIPTION

{app_name} retrieves a simple formated JSON record from 
a dataset collection (i.e value of C_NAME) and returns an 
Invenio-RDM JSON data structure.

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

# EXAMPLES

Retrieves record 85542 from caltechauthors.ds and return a JSON
object as an Invenio-RDM structure.

~~~
    {app_name} caltechauthors.ds 85542
~~~

''')

def read_eprintids(f_name):
    keys = []
    with open(f_name) as f:
        for line in f:
            keys.append(line.strip())
    if len(keys) == 0:
        return [], f'No keys found in {f_name}'
    return keys, '' 
    
def app_setup(app_name):
    c_name = None
    keys = None
    parser = argparse.ArgumentParser(
        prog = app_name,
        description="This program retreives a simple record from a dataset collection and fixes it up returning a JSON structure suitable for Invenio-RDM."
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
        "-eprintids",
        action = 'store',
        help = 'read eprint ids from file'
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
    if args.eprintids:
        keys, err = read_eprintids(args.eprintids)
        if err != '':
            print(err, file = os.stderr)
            sys.exit(1)
    return c_name, keys

#
# Main processing
#
app_name = os.path.basename(sys.argv[0])

c_name, keys = app_setup(app_name)

if not c_name:
    print(f'C_NAME not provided')
    sys.exit(1)
if not keys:
    print(f'KEY NOT PROVIDED')
    sys.exit(1)
if not os.path.exists(c_name):
    if os.path.exists(f'{c_name}.ds'):
        c_name = f'{c_name}.ds'
    else:
        print(f'{c_name} not found')
        sys.exit(1)

if keys == None:
    keys = []
    for key in sys.stdin:
        key = key.strip()
        if key != '':
            keys.append(key)

if isinstance(keys, list) and len(keys) > 0:
    for i, key in enumerate(keys):
        if key != '':
            if dataset.has_key(c_name, key) == True:
                rec, err = dataset.read(c_name, key)
                if err != '':
                    print(f'error ({i}) {err}', file = os.stderr)
                    sys.exit(1)
                rec = fixup_record(rec, files = None)
                print(json.dumps(rec, indent = 4))
