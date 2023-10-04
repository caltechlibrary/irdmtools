%eprint2rdm(1) irdmtools user manual | version 0.0.57-dev 0a28fdc
% R. S. Doiel and Tom Morrell
% 2023-10-04

# NAME

eprint2rdm

# SYNOPSIS

eprint2rdm [OPTIONS] [EPRINT_HOST] EPRINT_ID

# DESCRIPTION

eprint2rdm is a Caltech Library centric command line application
that takes an EPrint hostname and EPrint ID and returns a JSON
document suitable to import into Invenio RDM. It relies on
access to EPrint's REST API. It uses EPRINT_USER, EPRINT_PASSWORD
and EPRINT_HOST environment variables to access the API. Using
the "-all-ids" options you can get a list of keys available from
the EPrints REST API.

eprint2rdm can havest a set of eprint ids into a dataset collection
using the "-id-list" and "-harvest" options. You map also provide
customized resource type and person role mapping for the content
you harvest. This will allow you to be substantially closer to the
final record form needed to crosswalk EPrints data into Invenio RDM.

# OPTIONS

-help
: display help

-license
: display license

-version
: display version

-all-ids
: return a list of EPrint ids, one per line.

-harvest DATASET_NAME
: Harvest content to a dataset collection rather than standard out

-id-list ID_FILE_LIST
: (used with harvest) Retrieve records based on the ids in a file,
one line per id.

-resource-map FILENAME
: use this comma delimited resource map from EPrints to RDM resource types.
The resource map file is a comma delimited file without a header row.
The First column is the EPrint resource type string, the second is the
RDM resource type string.

-contributor-map FILENAME
: use this comma delimited contributor type map from EPrints to RDM
contributor types.  The contributor map file is a comma delimited file
without a header row. The first column is the value stored in the EPrints
table "eprint_contributor_type" and the second value is the string used
in the RDM instance.

# EXAMPLE


Example generating a JSON document for from the EPrints repository
hosted as "eprints.example.edu" for EPrint ID 118621.  Access to
the EPrint REST API is configured in the environment.  The result
is saved in "article.json". EPRINT_USER, EPRINT_PASSWORD and
EPRINT_HOST (e.g. eprints.example.edu) via the shell environment.

~~~
EPRINT_USER="__USERNAME_GOES_HERE__"
EPRINT_PASSWORD="__PASSWORD_GOES_HERE__"
EPRINT_HOST="eprints.example.edu"
eprint2rdm 118621 >article.json
~~~

Generate a list of EPrint ids from a repository 

~~~
eprint2rdm -all-ids >eprintids.txt
~~~

Generate a JSON document from the EPrints repository
hosted as "eprints.example.edu" for EPrint ID 118621 using a
resource map file to map the EPrints resource type to an
Invenio RDM resource type and a contributor type map for
the contributors type between EPrints and RDM.

~~~
eprint2rdm -resource-map resource_types.csv \
      -contributor-map contributor_types.csv \
      eprints.example.edu 118621 \
	  >article.json
~~~

Putting it together in the to harvest an EPrints repository
saving the results in a dataset collection for analysis or
migration.

1. create a dataset collection
2. get the EPrint ids to harvest applying a resource type map, "resource_types.csv"
   and "contributor_types.csv" for contributor type mapping
3. Harvest the eprint records and save in our dataset collection

~~~
dataset init eprints.ds
eprint2rdm -all-ids >eprintids.txt
eprint2rdm -id-list eprintids.txt -harvest eprints.ds
~~~

At this point you would be ready to improve the records in
eprints.ds before migrating them into Invenio RDM.

