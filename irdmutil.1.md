% irdmutil(1) user manual
% R. S. Doiel
% 2022-10-27

# NAME

irdmutil

# SYNOPSIS

irdmutil [OPTIONS] ACTION [ACTION_PARAMETERS ...]

# DESCRIPTION

__irdmutil__ provides a quick wrapper around access Invenio-RDM
JSON API. By default irdmutil looks in the current working directory
for a JSON configuration file that holds "base_url" to the Invenio-RDM
instance and any authentication information need to access the API.

# OPTIONS

help
: display help

license
: display license

version
: display version

config
: provide a path to an alternate configuration file

dataset
: store ivenio-rdm JSON record in a dataset collection (dataset v2)

# EXAMPLES

Get a list of Invenio-RDM record ids.

~~~
irdmutil get_record_ids
~~~

Get a specific Invenio-RDM record.

~~~
irdmutil get_record bq3se-47g50
~~~



irdmutil 0.0.0
