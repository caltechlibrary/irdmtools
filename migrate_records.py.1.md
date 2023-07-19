---
title: "migrate_records.py (1) user manual"
pubDate: 2023-02-28
author: "R. S. Doiel"
---

# NAME

migrate_records.py

# SYNOPSIS

migrate_records.py [OPTIONS] C_NAME [KEY]

# DESCRIPTION

migrate_records.py relies on two environment variables

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
    migrate_records.py CaltechAUTHORS.ds 12211
~~~

Iterating over the whole collection.

~~~
    migrate_records.py CaltechAUTHORS.ds
~~~


