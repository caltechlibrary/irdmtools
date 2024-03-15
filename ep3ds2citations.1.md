%ep3ds2citations(1) irdmtools user manual | version 0.0.71 f03d3547
% R. S. Doiel and Tom Morrell
% 2024-03-14

# NAME

ep3ds2citations

# SYNOPSIS

ep3ds2citations [OPTIONS] EPRINT_DS CITATION_DS [RECORD_ID]

# DESCRIPTION

ep3ds2citations is a Caltech Library oriented command line application
that takes an dataset collection of eprint records and converts then
to a citations dataset collection. It can do so for a single record id
or read a JSON list of record ids to migrate.

EPRINT_DS is the dataset collection holding the eprint records.

CITATION_DS is the dataset collection where the citation formatted
objects will be written.

# OPTIONS

-help
: display help

-license
: display license

-version
: display version

-config
: provide a path to an alternate configuration file (e.g. "irdmtools.json")

-ids JSON_ID_FILE
: read ids from a file.

-host
: Set the base url to use for the records (e.g. authors.library.caltech.edu)

-resource-types
: Use YAML file to map resouce types

-contributor-types
: Use YAML file to map contributor types

# EXAMPLE

Example of a dataset collection called "authors.ds" of EPrint records
and a "citations.ds" target that will hold citation records.

~~~shell
REPO_HOST="__HOST_NAME_OF_REPOSITORY__"
ep3ds2citations authors.ds citations.ds k3tpc-ga970
ep3ds2citations thesis.ds citations.ds 1233
~~~


