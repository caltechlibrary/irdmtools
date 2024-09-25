%ep3ds2citations(1) irdmtools user manual | version 0.0.88 94057474
% R. S. Doiel and Tom Morrell
% 2024-09-25

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

-keys
: works from a key list, one per line. Maybe file or standard input (use filename as "-")

-prefix
: Applies a prefix before the provided key when saving a record. E.g. `-prefix caltechauthors"

-host
: Set the base url to use for the records (e.g. authors.library.caltech.edu)

# EXAMPLE

Example of a dataset collection called "authors.ds", "data.ds" and
"thesis.ds" of EPrint records aggregated into a "citation.ds" dataset
collection using prefixes and the source repository ids.

~~~shell
ep3ds2citations -prefix caltechauthors \
           -host authors.library.caltech.edu \
           authors.ds citation.ds k3tpc-ga970
ep3ds2citations -prefix caltechdata \
           -host data.caltech.edu \
           data.ds citation.ds zzj7r-61978
ep3ds2citations -prefix caltechthesis \
           -host thesis.library.caltech.edu \
           thesis.ds citation.ds 1233
~~~


