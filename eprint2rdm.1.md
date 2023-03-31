% eprint2rdm(1) eprint2rdm user manual | Version 0.0.1
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
environment variables to access the API.

# OPTIONS

-help
: display help

-license
: display license

-version
: display version

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


