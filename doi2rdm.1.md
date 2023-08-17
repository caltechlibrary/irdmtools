%doi2rdm(1) irdmtools user manual | version 0.0.38 78797cd
% R. S. Doiel and Tom Morrell
% 2023-08-17

# NAME

doi2rdm

# SYNOPSIS

doi2rdm [OPTIONS] DOI

# DESCRIPTION

doi2rdm is a Caltech Library centric command line application
that takes a DOI, queries the CrossRef API then returns a JSON document
suitable for import into Invenio RDM. The DOI can be in either their
canonical form or URL form (e.g. "10.1021/acsami.7b15651" or
"https://doi.org/10.1021/acsami.7b15651").

# OPTIONS

-help
: display help

-license
: display license

-version
: display version

-diff JSON_FILENAME
: compare the JSON_FILENAME contents with record generated from CrossRef works record

-dot-initials
: Add period to initials in given name

-download
: attempt to download the digital object if object URL provided

-mailto
: (string) set the mailto value for CrossRef API access (default "helpdesk@library.caltech.edu")

-resource-map
: Use this two column CSV file (no header row) to map resource types in CrossRef to RDM

-contributor-map
: Use this two column CSV file (no header row) to map contributor types from CrossRef (e.g.
"author", "translator", "editor", "chair") to RDM roles.

# EXAMPLES

Example generating a JSON document for a single DOI. The resulting
text file is called "article.json".

~~~
	doi2rdm "10.1021/acsami.7b15651" >article.json
~~~

Check to see the difference from the saved "article.json" and
the current metadata retrieved from CrossRef.

~~~
	doi2rdm -diff article.json "10.1021/acsami.7b15651
~~~


