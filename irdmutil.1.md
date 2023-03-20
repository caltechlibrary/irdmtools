% irdmutil(1) user manual
% R. S. Doiel
% 2023-03-20

# NAME

irdmutil

# SYNOPSIS

irdmutil [OPTIONS] ACTION [ACTION_PARAMETERS ...]

# DESCRIPTION

__irdmutil__ provides a quick wrapper around Invenio-RDM's OAI-PMH
and REST API. By default irdmutil looks for two environment variables.

- RDM_INVENIO_API
- RDM_INVENIO_TOKEN

These are use to acces the Invenio RDM REST API and OAI-PMH services.

You may specify a JSON configuration file holding the attributes of 
"invenio_api" and "invenio_token" instead of using environment variables.

irdmutil uses the OAI-PMH service to retrieve record ids. This maybe
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
: provide a path to an alternate configuration file (default is irdmtools.json)

# ACTION

__irdmutil__ supports the following actions.

setup
: Display an example JSON setup configuration file, if it already exists then it will display the current configuration file. No optional or required parameters. When displaying the JSON configuration a placeholder will be used for the token value.

get_modified_ids START [END]
: Returns a list of modified record ids (created, updated, deleted) in the time range listed.  This method uses OAI-PMH for id retrieval. It is rate limited. Start and end dates are inclusive and should be specific in YYYY-MM-DD format.

get_all_ids
: Returns a list of all repository record ids. The method uses OAI-PMH for id retrieval. It is rate limited and will take come time to return all record ids. A test instance took 11 minutes to retrieve 24000 record ids.

query QUERY_STRING [size | size sort]
: Returns a result using RDM's search engine. It is limited to about 10K total results. You can use the see RDM's documentation for query construction.  See <https://inveniordm.docs.cern.ch/customize/search/>, <https://inveniordm.docs.cern.ch/reference/rest_api_requests/> and https://www.elastic.co/guide/en/elasticsearch/reference/current/query-dsl-query-string-query.html#query-string-syntax. Query takes one required parameter and two option.


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

Get a list of Invenio-RDM record ids modified from
Jan 1, 2023 to Jan 31, 2023.

~~~
irdmutil get_modified_ids 2023-01-01 2023-01-31
~~~

Get a list of all Invenio-RDM record ids.

~~~
irdmutil get_all_ids
~~~

Get a specific Invenio-RDM record.

~~~
irdmutil get_record bq3se-47g50
~~~



irdmutil 0.0.0
