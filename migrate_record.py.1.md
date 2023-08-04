---
title: "migrate_record.py (1) user manual"
pubDate: 2023-03-03
author: "R. S. Doiel"
---

# NAME

migrate_record.py

# SYNOPSIS

migrate_record.py [OPTIONS]

# DESCRIPTION

migrate_record.py is a filter program. It reads from standard input and
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
    eprint2rdm authors.library.caltech.edu 85542 |      migrate_record.py | rdmutil new_record
~~~


