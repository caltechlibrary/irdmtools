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

{app_name} [OPTIONS]

# DESCRIPTION

{app_name} is a filter program. It reads from standard input and
writes fixed up RDM records to standard output. It is intended to be
used in a pipe line with eprint2rdm and rdmutil.

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
    eprint2rdm authors.library.caltech.edu 85542 |\
      {app_name} | rdmutil new_record
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

#
# Main processing
#
app_name = os.path.basename(sys.argv[0])

app_setup(app_name)
# Read the JSON from standard input
src = sys.stdin.read()
try: 
    rec = json.loads(src)
except Exception as err:
    print(err, file = sys.stderr)
    sys.exit(1)
print(json.dumps(fixup_record(rec, files = None), indent = 4))
