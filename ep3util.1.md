%ep3util(1) irdmtools user manual | version 0.0.57-dev 0a28fdc
% R. S. Doiel and Tom Morrell
% 2023-10-04

# NAME

ep3util

# SYNOPSIS

ep3util [OPTIONS] ACTION [ACTION_PARAMETERS ...]

# DESCRIPTION

__ep3util__ provides a quick wrapper around EPrints 3.3 REST API.
By default ep3util looks for five environment variables.

REPO_ID
: the EPrints repository id (name of database and archive subdirectory).

EPRINT_HOST
: the hostname for EPrint's.

EPRINT_USER
: the username having permissions to access the EPrint REST API.

EPRINT_PASSWORD
: the password for the username with access to the EPrint REST API.

C_NAME
: If harvesting the dataset collection name to harvest the records to.

EPRINT_DB_HOST
: The MySQL hostname holding the EPrints repository database

EPRINT_DB_USER
: The MySQL username used to access EPrints repository database 

EPRINT_DB_PASSWORD
: The MySQL password used to access EPrints repository database

The environment provides the default values for configuration. They
maybe overwritten by using a JSON configuration file. The corresponding
attributes are "repo_id", "eprint_host", "c_name", "eprint_db_host",
"eprint_db_user", and "eprint_db_password".

If the environment variables for MySQL access are set then the results
reflect direct access to the database instead of the EPrint REST API.


# OPTIONS

help
: display help

license
: display license

version
: display version

config
: provide a path to an alternate configuration file (e.g. "rdmtools.json")

# ACTION

__ep3util__ supports the following actions.

setup
: Display an example JSON setup configuration file, if it already exists then it will display the current configuration file. No optional or required parameters. When displaying the JSON configuration a placeholder will be used for the token value.

get_all_ids
: Returns a list of all repository record ids. The method uses OAI-PMH for id retrieval. It is rate limited and will take come time to return all record ids. A test instance took 11 minutes to retrieve 24000 record ids.

get_modified_ids START [END]
: Return a list of records created or modified in the START and END date range. If END is not provided it is assume to be today.

get_record RECORD_ID
: Returns a specific simplified record indicated by RECORD_ID, e.g. 23808. The REORCID_ID is a required parameter.

harvest [HARVEST_OPTIONS] [KEY_LIST_JSON]
: harvest takes a JSON file containing a list of keys and harvests each record into a dataset collection. If combined
with one of the options, e.g. `-all`, you can skip providing the KEY_LIST_JSON file.

# HARVEST_OPTIONS

-all
: Harvest all records

-modified START [END]
: Harvest records modified between start and end dates.

# ACTION_PARAMETERS

Action parameters are the specific optional or required parameters need to complete an aciton.


# EXAMPLES

Setup for __ep3util__ by writing an example JSON configuration file.
"nano" is an example text editor program, you need to edit the sample
configuration appropriately.

~~~
ep3util setup >eprinttools.json
nano eprinttools.json
~~~

Get a list of all EPrint record ids.

~~~
ep3util get_all_ids
~~~

Get a specific EPrint record. Record is validated
against irdmtool EPrints data model.

~~~
ep3util get_record 23808
~~~

Harvest all records

~~~
ep3util harvest -all
~~~

Harvest records created or modified in the month of September, 2023.

~~~
ep3util harvest -modified 2023-09-01 2023-09-30
~~~

