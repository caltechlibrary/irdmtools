% rdmutil(1) rdmutil user manual | Version 0.0.10
% R. S. Doiel and Tom Morrell
% 2023-04-13

# NAME

rdmutil

# SYNOPSIS

rdmutil [OPTIONS] ACTION [ACTION_PARAMETERS ...]

# DESCRIPTION

__rdmutil__ provides a quick wrapper around Invenio-RDM's OAI-PMH
and REST API. By default rdmutil looks for three environment variables.

RDM_INVENIO_API
: the URL of the Invenio RDM API and OAI-PMH services

RDM_INVENIO_TOKEN
: the token needed to access the Invenio RDM API and OAI-PMH services


RDM_C_NAME
: A dataset collection name. Collection must exist. See `dataset help init`

The environment provides the default values for configuration. They
maybe overwritten by using a JSON configuration file. The corresponding
attributes are "invenio_api", "invenio_token" and "c_name".

rdmutil uses the OAI-PMH service to retrieve record ids. This maybe
slow due to rate limits. Also provided is a query service and record
retrieval using Invenio RDM's REST API. These are faster but the query
services limited the total number of results to 10K records.

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

__rdmutil__ supports the following actions.

setup
: Display an example JSON setup configuration file, if it already exists then it will display the current configuration file. No optional or required parameters. When displaying the JSON configuration a placeholder will be used for the token value.

get_modified_ids START [END]
: Returns a list of modified record ids (created, updated, deleted) in the time range listed.  This method uses OAI-PMH for id retrieval. It is rate limited. Start and end dates are inclusive and should be specific in YYYY-MM-DD format.

get_all_ids
: Returns a list of all repository record ids. The method uses OAI-PMH for id retrieval. It is rate limited and will take come time to return all record ids. A test instance took 11 minutes to retrieve 24000 record ids.

query QUERY_STRING [size | size sort]
: Returns a result using RDM's search engine. It is limited to about 10K total results. You can use the see RDM's documentation for query construction.  See <https://inveniordm.docs.cern.ch/customize/search/>, <https://inveniordm.docs.cern.ch/reference/rest_api_requests/> and https://www.elastic.co/guide/en/elasticsearch/reference/current/query-dsl-query-string-query.html#query-string-syntax. Query takes one required parameter and two option.

get_record RECORD_ID
: Returns a specific simplified record indicated by RECORD_ID, e.g. bq3se-47g50. The REORCID_ID is a required parameter.

get_raw_record RECORD_ID
: Returns a specific map/dictionary record indicated by RECORD_ID, e.g. bq3se-47g50. The REORCID_ID is a required parameter.

harvest KEY_JSON
: harvest takes a JSON file containing a list of keys and harvests each record into the dataset collection.

# ACTION_PARAMETERS

Action parameters are the specific optional or required parameters need to complete an aciton.


# EXAMPLES

Setup for __rdmutil__ by writing an example JSON configuration file.
"nano" is an example text editor program, you need to edit the sample
configuration appropriately.

~~~
rdmutil setup >rdmtools.json
nano rdmtools.json
~~~

Get a list of Invenio-RDM record ids modified from
Jan 1, 2023 to Jan 31, 2023.

~~~
rdmutil get_modified_ids 2023-01-01 2023-01-31
~~~

Get a list of all Invenio-RDM record ids.

~~~
rdmutil get_all_ids
~~~

Get a specific Invenio-RDM record. Record is validated
against irdmtool model.

~~~
rdmutil get_record bq3se-47g50
~~~

Get a specific Invenio-RDM record as it is returned by
the RDM API.

~~~
rdmutil get_raw_record bq3se-47g50
~~~

