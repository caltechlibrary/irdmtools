%rdmutil(1) irdmtools user manual | version 0.0.88 94057474
% R. S. Doiel and Tom Morrell
% 2024-09-25

# NAME

rdmutil

# SYNOPSIS

rdmutil [OPTIONS] ACTION [ACTION_PARAMETERS ...]

# DESCRIPTION

__rdmutil__ is way of interacting with Invenio-RDM through its Postgres
database. It does NOT use the OAI-PMH API since that is far too slow.
rdmutil uses environment variables for configuration. For accessing the
JSON API it uses the following.

dataset related environment variables (i.e. for harvest action) 

__rdmutil__ environment variable for storing harvested content.

C_NAME
: A dataset collection name. Collection must exist.
See `dataset help init`

__rdmutil__ PostgreSQL related environment variabels are

REPO_ID
: Repository id (i.e. the name of the Postgres database used by RDM)

RDM_DB_USER
: The username used to access Postgres database

RDM_DB_PASSWORD
: (optional) The password if needed to access the Postgres database.

RDM_DB_HOST
: (optional) The hostname of the database server to access runing Postgres.
by default it assumes localhost running on port 5432.


# OPTIONS

help
: display help

license
: display license

version
: display version

config
: provide a path to an alternate configuration file (e.g. "irdmtools.json")

# ACTION

__rdmutil__ supports the following actions.

setup
: Display an example JSON setup configuration file, if it already exists
then it will display the current configuration file. No optional or required
parameters.  When displaying the JSON configuration a placeholder will be
used for the token value.

get_modified_ids START [END]
: Returns a list of modified record ids (created, updated, deleted) in the
time range listed.  If Postgres is not configured this method uses OAI-PMH
for id retrieval. It is rate limited.  Start and end dates are inclusive
and should be specific in YYYY-MM-DD format.

get_all_ids
: Returns a list of all repository record ids latest versions. The method
requires Postgres database access.

get_all_stale_ids
: Returns a list of public record ids that are NOT the latest version of the
records, useful when prune a dataset collection of stale RDM records.

check_doi DOI
: This takes a DOI and searches the .pids.doi.identifiers for matching rdm
records or drafts. DOI is required.

query QUERY_STRING [size | size sort]
: Returns a result using RDM's search engine. The JSON API access must be
defined. It is limited to about 10K total results. You can use the see RDM's
documentation for query construction.  See <https://inveniordm.docs.cern.ch/customize/search/>,
<https://inveniordm.docs.cern.ch/reference/rest_api_requests/> and
<https://www.elastic.co/guide/en/elasticsearch/reference/current/query-dsl-query-string-query.html#query-string-syntax>.
Query takes one required parameter and two optional.

get_record RECORD_ID
: Returns a specific simplified record indicated by RECORD_ID, e.g. bq3se-47g50.
The RECORD_ID is a required parameter.  If Postgres access is defined then it
does this by SQL assembling an object similar to the one available via 
"/api/records/RECORD_ID" path in the JSON API. If Postgres is not configured
then it retrieves the record via the JSON API.

get_raw_record RECORD_ID
: Returns a specific map/dictionary record indicated by RECORD_ID, e.g. bq3se-47g50.
The RECORD_ID is a required parameter. It uses the JSON API.

get_record_versions RECORD_ID
: Get the records versions as a JSON array. It requires Postgres DB access and
returns the versions found in the rdm_records_metadata_versions table.

get_files RECORD_ID
: Return a list of files for record with RECORD_ID.  RECORD_ID is required.

get_file RECORD_ID FILENAME
: Returns the metadata for a file indicated by RECORD_ID and FILENAME,
e.g. bq3se-47g50 is a record id and article.pdf is a filename. RECORD_ID
and FILENAME are required parameters.

retrieve_file RECORD_ID FILENAME [OUTPUT_FILENAME]
: Retrieves the file indicated by RECORD_ID and FILENAME, e.g. bq3se-47g50
is a record id and article.pdf is a filename. RECORD_ID and FILENAME arei
required parameters. OUTPUT_FILENAME is optional, if present then the file
will be save to this name on the file system.

get_versions
: Retrieve metadata about versions for RECORD_ID. RECORD_ID is required.

get_latest_version
: Retrieve the latest version metadata for a RECORD_ID. RECORD_ID is required.

set_version RECORD_ID VERSION_STRING
: This sets the version string in .metadata.version of a draft. RECORD_ID and
VERSION_STRING are required.

set_publication_date RECORD_ID PUBICLATION_DATE
: This sets the publication date, .metadata.publication_date, in draft.
RECORD_ID and PUBLICATION_DATE required. PUBLICATION_DATE can be in YYYY,
YYYY-MM, YYYY-MM-DD format. 

new_record [FILENAME]
: Create a new record from JSON source. If FILENAME is set then json source
is read from FILENAME otherwise it reads from standard input.

new_version RECORD_ID
: This will create a new version of the record. RECORD_ID is required.
NOTE: When you create a new version .metadata.publication_date and 
.metadata.version are removed.  These need to be you recreated in the draft
of the new version as well as to handle any files, access restrictions before
publishing the new version draft.

publish_version RECORD_ID [VERSION_LABEL] [PUBLICATION_DATE]
: This will publish a new draft version. RECORD_ID is required. VERSION_LABEL
is optional. PUBLICATION_DATE is optional. PUBLICATION_DATE is in YYYY,
YYYY-MM or YYYY-MM-DD format.  Will need to set .metadata.publication_date,
doi and version name in the versioned draft before publish_version is called.
RECORD_ID is required.

new_draft RECORD_ID
: Create a new draft for an existing record. RECORD_ID is required. 

get_draft RECORD_ID
: Retrieve an existing draft record for RECORD_ID. RECORD_ID is required.
If draft of RECORD_ID does not exist you will see a 404 error.

update_draft RECORD_ID [FILENAME]
: Update a draft record. RECORD_ID is required. FILENAME is optional, if
one is provided the JSON document is used to update RDM, otherwise standard
input is used to get the JSON required to do the update.

set_files_enable RECORD_ID true|false
: This will flip the files.enabled value to true and update the draft.
RECORD_ID is required. The one of the values true or false are required.

upload_files RECORD_ID FILENAME [FILENAME ...]
: Upload files to a draft record. RECORD_ID is required as are one or more
filenames.

get_files RECORD_ID
: Retrieve the list of files attached to a draft. RECORD_ID is required.

delete_files RECORD_ID FILENAME [FILENAME ...]
: Delete files in a draft record. RECORD_ID is required as are one or more
filenames.

discard_draft
: Discard (delete) a draft record from RDM. RECORD_ID is required.

review_comment RECORD_ID [FILENAME]
: Submit a comment to a review. RECORD_ID is required. If FILENAME is provided
the comment is read from a file otherwise it is read from standard input.

send_to_community RECORD_ID COMMUNITY_ID
: Submit a draft record to a community for review. RECORD_ID and COMMUNITY_ID
are required.

get_review
: Get review requests associated with RECORD_ID. RECORD_ID is required.

review_request RECORD_ID accept|decline|cancel|"" [COMMENT]
: Review a submitted draft record. the values "accept", "decline" or ""
and an optional COMMENT.

get_access RECORD_ID [ACCESS_TYPE]
: This will return the JSON for the access attribute in the record. If you
include ACCESS_TYPE of "files" or "records" it will return just that attribute.
RECORD_ID is always required.

set_access RECORD_ID ACCESS_TYPE ACCESS_VALUE
: This will update a record with metadata access to the record. RECORD ID is
required. ACCESS_TYPE is required and can be either "record" or "files".
ACCESS_VALUE is required and can be "restricted" or "public".

harvest KEY_JSON
: harvest takes a JSON file containing a list of keys and harvests each record
into the dataset collection indicated by the environment variable C_NAME.


get_endpoint PATH
: Perform a GET to the end point indicated by PATH. PATH is required.

post_endpoint PATH [FILENAME]
: Perform a POST to the end point indicated by PATH. PATH is required. If
FILENAME is provided then JSON source is read file the file otherwise it is
read from standard input.

put_endpoint PATH [FILENAME]
: Perform a PUT to the end point indicated by PATH. PATH is required. If
FILENAME is provided then JSON source is read file the file otherwise it
is read from standard input.

patch_endpoint PATH [FILENAME]
: Perform a PATCH to the end point indicated by PATH. PATH is required.
If FILENAME is provided then JSON source is read file the file otherwise
it is read from standard input.

# ACTION_PARAMETERS

Action parameters are the specific optional or required parameters need
to complete an action.


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

