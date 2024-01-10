%rdm2eprint(1) irdmtools user manual | version 0.0.66 5281a5b2
% R. S. Doiel and Tom Morrell
% 2024-01-10

# NAME

rdm2eprint

# SYNOPSIS

rdm2eprint [OPTIONS] RDM_ID

# DESCRIPTION

rdm2eprint is a Caltech Library oriented command line application
that takes an RDM record ID and returns a EPrint record JSON document.
It was created to allow us migrate our EPrints repositories minimal change
to our feeds system which works with EPrint structured data.
It uses RDM_URL, RDMTOK, RDM_COMMUNITY_ID environment variables for
configuration.  It can read data from a previously harvest RDM record
or directly from RDM via the API url. The tool is intended to run
in a pipe line so have minimal options.

# OPTIONS

-help
: display help

-license
: display license

-version
: display version

-config
: provide a path to an alternate configuration file (e.g. "irdmtools.json")

-harvest C_NAME
: harvest JSON formatted eprint records into the dataset collection 
specified by C_NAME.

-ids JSON_ID_FILE
: read ids from a file.

-xml
: output as EPrint XML rather than JSON, does not work with -harvest.

-pipeline
: read from standard input and write crosswalk to standard out.

# EXAMPLE

Example generating a EPRINT JSON document from RDM would use the following
variables.

the EPrints repository
hosted as "eprints.example.edu" for EPrint ID 118621.  Access to
the EPrint REST API is configured in the environment.  The result
is saved in "article.json". EPRINT_USER, EPRINT_PASSWORD and
EPRINT_HOST (e.g. eprints.example.edu) via the shell environment.

~~~
RDM_URL="__URL_TO_RDM_INSTANCE_HERE__"
RDMTOK="__RDM_ACCESS_TOKEN_HERE__"
RDM_COMMUNITY_ID="rdm.example.edu"
rdm2eprint k3tpc-ga970 >article.json
~~~


