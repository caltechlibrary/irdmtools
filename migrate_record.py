#!/usr/bin/env python3
'''migrate_record.py is a cli for filter the output of eprint2rdm and making 
it ready for Invenio RDM import with rdmutil.'''

import os
import sys
import json
import argparse

from irdm import fixup_record, license_text, version_text

def usage(a_name):
    '''explain cli usage'''
    print(f'''---
title: "{a_name} (1) user manual"
pubDate: 2023-03-03
author: "R. S. Doiel"
---

# NAME

{a_name}

# SYNOPSIS

{a_name} [OPTIONS]

# DESCRIPTION

{a_name} is a filter program. It reads from standard input and
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
      {a_name} | rdmutil new_record
~~~

''')

def read_eprintids(f_name):
    '''read an eprint id list from f_name'''
    keys = []
    with open(f_name, encoding='utf-8') as f_p:
        for line in f_p:
            keys.append(line.strip())
    if len(keys) == 0:
        return [], f'No keys found in {f_name}'
    return keys, ''

def app_setup(a_name):
    '''initialize application and parameter processing'''
    parser = argparse.ArgumentParser(
        prog = a_name,
        description="Retreive a record from a collection and fixes it up for Invenio-RDM."
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
        print(license_text)
        sys.exit(0)
    if args.version:
        print(version_text)
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
print(json.dumps(fixup_record(rec), indent = 4))
