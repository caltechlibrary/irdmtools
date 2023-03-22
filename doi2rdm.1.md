% doi2rdm(1) doi2rdm user manual | Version 0.0.1
% R. S. Doiel and Tom Morrell
% 2023-03-22

# NAME

doi2rdm

# SYNOPSIS

doi2rdm [OPTIONS] DOI_OR_FILENAME

# DESCRIPTION

doi2rdm is a Caltech Library centric application that takes one or
more DOI, queries the CrossRef API and if that fails the DataCite API
and returns a JSON document suitable for import into Invenio RDM. The
DOI can be in either their canonical form or URL form 
(e.g. "10.1021/acsami.7b15651" or "https://doi.org/10.1021/acsami.7b15651").

If you pass a filename instead of a DOI then the DOI will be read from
the the file expecting one DOI per line.

# OPTIONS

-help
: display help

-license
: display license

-version
: display version

-crossref
: only search CrossRef API for DOI records

-datacite
: only search DataCite API for DOI records

-dot-initials
: Add period to initials in given name

-download
: attempt to download the digital object if object URL provided

-mailto
: (string) set the mailto value for CrossRef API access (default "helpdesk@library.caltech.edu")

# EXAMPLES

Example generating a JSON document for a single DOI. The resulting
text file is called "article.json".

~~~
	doi2rdm "10.1021/acsami.7b15651" >article.json
~~~

Example generating a JSON document for two DOI. The resulting
text file is called "articles.json".

~~~
	doi2rdm "10.1021/acsami.7b15651" "10.1093/mnras/stu2495" >articles.json
~~~

Example processing a list of DOIs in a text file called "doi-list.txt" and
writing the output to a JSON document called "articles.json".

~~~
	doi2rdm doi-list.txt >articles.json
~~~


