% doi2rdm(1) doi2rdm user manual | Version 0.0.1
% R. S. Doiel and Tom Morrell
% 2023-03-22

# NAME

doi2rdm

# SYNOPSIS

doi2rdm [OPTIONS] DOI_OR_FILENAME

# DESCRIPTION

doi2rdm is a Caltech Library centric command line application
that takes a DOI, queries the CrossRef API and if that fails the DataCite API
before returning a JSON document suitable for import into Invenio RDM. The
DOI can be in either their canonical form or URL form
(e.g. "10.1021/acsami.7b15651" or "https://doi.org/10.1021/acsami.7b15651").

# OPTIONS

-help
: display help

-license
: display license

-version
: display version

-config FILENAME
: use configuration file

-diff JSON_FILENAME
: compare the JSON_FILENAME contents with record generated from CrossRef works record

-dot-initials
: Add period to initials in given name

-download
: attempt to download the digital object if object URL provided

-mailto
: (string) set the mailto value for CrossRef API access (default "helpdesk@library.caltech.edu")

-setup
: Display an example configuration or the configuration

# EXAMPLES

Example generating a configuration example irdmtools saving
the configuration to a text file named "doi2rdm.json".

~~~
doi2rdm -setup >doi2rdm.json
~~~

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


