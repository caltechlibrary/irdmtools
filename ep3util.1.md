%ep3util(1) irdmtools user manual | version 0.0.55 3ade190
% R. S. Doiel and Tom Morrell
% 2023-09-25

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


The environment provides the default values for configuration. They
maybe overwritten by using a JSON configuration file. The corresponding
attributes are "repo_id", "eprint_host" and "c_name".


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

get_record RECORD_ID
: Returns a specific simplified record indicated by RECORD_ID, e.g. 23808. The REORCID_ID is a required parameter.

harvest KEY_LIST_JSON
: harvest takes a JSON file containing a list of keys and harvests each record into the dataset collection.


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


