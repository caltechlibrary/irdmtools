%eprintrest(1) irdmtools user manual | version 0.0.68-dev ce0479d4
% R. S. Doiel and Tom Morrell
% 2024-01-12

# NAME

eprintrest

# SYNOPSIS

eprintrest [OPTIONS]

# DESCRIPTION

eprintrest is a Caltech Library oriented localhost web service
that creates a functionally similar replica of the EPrints REST API 
for EPrints 3.3.x based repositories. It uses the path to the 
"archives" directory and a MySQL Database for the repository. 
It only supports "archive" eprint.eprint_status records and
only the complete XML. Start up time is slow because it builds 
the data structures representing the content in memory. This
makes the response times to request VERY fast compared to
the EPrints REST API.

NOTE: the rest API does not enforce user permissions, restrictions
or roles. It is a minimal READ ONLY re-implementation of the EPrints 3.3
REST API!

The application is configured from the environment. The following
environment variables need to be set.

REPO_ID
: The repository id string (e.g. caltechauthors). Also the name of the database for the repository.

EPRINT_ARCHIVES_PATH
: A path to the "archives" directory holding your repository content 
(e.g. /usr/local/eprints/archives)

DB_USER
: The user name needed to access the MySQL database[^1]

DB_PASSWORD
: The password needed to access the MySQL database[^1]

REST_PORT
: The localhost port to use for the read only REST API.

[^1]: MySQL, like this REST service assumes to be running on localhost.


# OPTIONS

-help
: display help

-license
: display license

-version
: display version


# EXAMPLE

This is an example environment

~~~
REPO_ID="caltechauthors"
EPRINT_ARCHIVES_PATH="/code/eprints3.3/archives"
REST_PORT=80
DB_USER="eprints"
DB_PASSWORD="something_secret_here"
~~~

Running the localhost REST API clone

~~~
eprintrest
~~~


