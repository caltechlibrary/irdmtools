% irdmutil(1) user manual
% R. S. Doiel
% 2023-03-16

# NAME

irdmutil

# SYNOPSIS

irdmutil [OPTIONS] ACTION [ACTION_PARAMETERS ...]

# DESCRIPTION

__irdmutil__ provides a quick wrapper around access Invenio-RDM
JSON API. By default irdmutil looks in the current working directory
for a JSON configuration file that holds "base_url" to the Invenio-RDM
instance and any authentication information need to access the API.

__irdmutil__ normally expects a configuration file but if one
is not found then it can read its configuration from the environment.
The two environment variables used as INVENIO_API holding the string
that points to the Invenio JSON API and INVENIO_TOKEN needed to 
authenticate and use the API.  If one or both a missing an error will
be returned. If the environment INVENIO_API is set and you run the 
setup action they it will be used to populate the JSON configuration
example displayed by setup. If a configuration file exists it will
display the configuration with the invenio token overwritten with
a placeholder value.

# OPTIONS

help
: display help

license
: display license

version
: display version

config
: provide a path to an alternate configuration file (default is irdmtools.json)

# ACTION

__irdmutil__ supports the following actions.

setup
: Display an example JSON setup configuration file, if it already exists then it will display the current configuration file. No optional or required parameters.

get_record_ids
: Returns a list of all repository record ids (can take a while). No optional or required parameters.

get_record RECORD_ID
: Returns a specific record indicated by RECORD_ID, e.g. bq3se-47g50. The REORCID_ID is a required parameter.

# ACTION_PARAMETERS

Action parameters are the specific optional or required parameters need to complete an aciton.


# EXAMPLES

Setup for __irdmutil__ by writing an example JSON configuration file.
"nano" is an example text editor program, you need to edit the sample
configuration appropriately.

~~~
irdmutil setup >irdmtools.json
nano irdmtools.json
~~~

Get a list of Invenio-RDM record ids.

~~~
irdmutil get_record_ids
~~~

Get a specific Invenio-RDM record.

~~~
irdmutil get_record bq3se-47g50
~~~



irdmutil 0.0.0
