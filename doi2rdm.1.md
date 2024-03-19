%doi2rdm(1) irdmtools user manual | version 0.0.73 a0479527
% R. S. Doiel and Tom Morrell
% 2024-03-18

# NAME

doi2rdm

# SYNOPSIS

doi2rdm [OPTIONS] [OPTIONS_YAML] crossref|datacite DOI

# DESCRIPTION

doi2rdm is a Caltech Library oriented command line application
that takes a DOI, queries the CrossRef or DataCite API then returns a
JSON document suitable for import into Invenio RDM. The DOI can be
in either their canonical form or URL form (e.g. "10.1021/acsami.7b15651" or
"https://doi.org/10.1021/acsami.7b15651").

# OPTIONS_YAML

doi2rdm can use an YAML options file to set the behavior of the
crosswalk from CrossRef to RDM. This replaces many of the options
previously required in prior implementations of this tool. See all the
default options setting use the `-show-yaml` command line
options. You can save this to disk, modify it, then use them for
migrating content from CrossRef to RDM.

# OPTIONS

-help
: display help

-license
: display license

-version
: display version

-diff JSON_FILENAME
: compare the JSON_FILENAME contents with record generated from CrossRef or DataCite works record

-show-yaml
: This will display the default YAML configuration file. You can save this and customize to suit your needs.

# EXAMPLES

Example generating a JSON document for a single DOI. The resulting
text file is called "article.json". In this example "options.yaml"
is the configuration file for setup for your RDM instance.

~~~
	doi2rdm options.yaml crossref "10.1021/acsami.7b15651" >article.json
~~~

Check to see the difference from the saved "article.json" and
the current metadata retrieved from CrossRef.

~~~
	doi2rdm -diff article.json crossref doi2rdm.yaml "10.1021/acsami.7b15651
~~~

Save the default YAML options to a file. 

~~~
	doi2rdm -show-yaml >options.yaml
~~~

