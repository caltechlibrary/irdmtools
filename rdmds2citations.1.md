%rdmds2citations(1) irdmtools user manual | version 0.0.84 be45fd52
% R. S. Doiel and Tom Morrell
% 2024-07-10

# NAME

rdmds2citations

# SYNOPSIS

rdmds2citations [OPTIONS] RDM_DS CITATION_DS [RECORD_ID]

# DESCRIPTION

rdmds2citations is a Caltech Library oriented command line application
that takes an dataset collection of RDM records and converts then
to a citations dataset collection. It can do so for a single record id
or read a JSON list of record ids to migrate.


RDM_DS is the dataset collection holding the eprint records.

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

-keys
: works from a key list, one per line. Maybe file or standard input (use filename as "-")

-prefix
: Applies a prefix before the provided key when saving a record. E.g. `-prefix authors" will 
prefix keys with "authors:"

-host
: Set the hostname of base url to for reference records (e.g. authors.library.caltech.edu). Can also be set via the environment as RDM_URL.

# ENVIRONMENT 

Some settings can be picked from the environment.

RDM_URL
: Sets the URL to the RDM is installed (e.g. "https://authors.library.caltech.edu").

# EXAMPLE

Example of a dataset collection called "authors.ds" and "data.ds"
RDM records aggregated into a "citation.ds" dataset
collection using prefixes and the source repository ids.

~~~shell
rdmds2citations -prefix authors \
           -host authors.library.caltech.edu \
		   authors.ds citations.ds k3tpc-ga970
rdmds2citations -prefix data \
           -host data.caltech.edu \
		   data.ds citations.ds zzj7r-61978
~~~


