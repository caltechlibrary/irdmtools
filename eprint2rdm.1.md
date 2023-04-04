% eprint2rdm(1) eprint2rdm user manual | Version 0.0.2
% R. S. Doiel and Tom Morrell
% 2023-03-30

# NAME

eprint2rdm

# SYNOPSIS

eprint2rdm [OPTIONS] EPRINT_HOSTNANE EPRINT_ID

# DESCRIPTION

eprint2rdm is a Caltech Library centric command line application
that takes an EPrint hostname and EPrint ID and returns a JSON
document suitable to import into Invenio RDM. It relies on
access to EPrint's REST API. It uses EPRINT_USER and EPRINT_PASSWORD
environment variables to access the API. Using the "-keys" options
you can get a list of keys available from the EPrints REST API.

# OPTIONS

-help
: display help

-license
: display license

-version
: display version

-all-ids
: return a list of EPrint ids, one per line.

-resource-map FILENAME
: use this comma delimited resource map from EPrints to RDM resource types.
The resource map file is a comma delimited file without a header row.
first column is the EPrint resource type string, the second is the
RDM resource type string.


# EXAMPLE


Example generating a JSON document for from the EPrints repository
hosted as "eprints.example.edu" for EPrint ID 118621.  Access to
the EPrint REST API is configured in the environment.  The result
is saved in "article.json".

~~~
EPRINT_USER="__USERNAME_GOES_HERE__"
EPRINT_PASSWORD="__PASSWORD_GOES_HERE__"
eprint2rdm eprints.example.edu 118621 \
	>article.json
~~~

Generate a list of EPrint ids from a repository (e.g. eprints.example.edu).

~~~
eprint2rdm -all-ids eprints.example.edu >eprintids.txt
~~~

Generate a JSON document from the EPrints repository
hosted as "eprints.example.edu" for EPrint ID 118621 using a
resource map file to map the EPrints resource type to an
Invenio RDM resource type.

~~~
eprint2rdm --resource-map resource-types.csv \
      eprints.example.edu 118621 \
	  >article.json
~~~

Putting it together in the to harvest an EPrints repository
saving the results in a dataset collection for analysis or
migration.

1. create a dataset collection
2. get the EPrint ids to harvest applying a resource type map, "resource-types.csv"
3. Harvest the eprint records and save in our dataset collection

~~~
dataset init example_edu.ds
eprint2rdm -all-ids eprints.example.edu >eprintids.txt
while read EPRINTID; do
    eprint2rdm -resource-map resource-types.csv \
       eprints.example.edu "${EPRINTID}" |\
	   dataset create -i - example_edu.ds "${EPRINTID}"
done <eprintids.txt
~~~

At this point you would be ready to improve the records in
example_edu.ds before migrating them into Invenio RDM.
