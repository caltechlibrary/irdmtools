
Institutional Repository Data Management Tools
==============================================

This is a proof of concept set tools and Go packages for
working with institutional repositories. Initial target is
Invenio RDM.

The proof of concept is being developed around RDM's web services 
(e.g. REST API and OAI-PMH), PostgreSQL database and external metadata
services (e.g. CrossRef, DataCite).

Caltech Library is using testing the prototype with continuous content
migration, aggregation and metadata analysis.

## Tools

### `rdmutil`

This tool is for interacting with an Invenio RDM repository.

- get_all_ids uses the OAI-PMH service to harvest all the current record ids in an Invenio RDM instance (very slow due to rate limits)
- get_modified_ids uses the OAI-PMH service with the "from" and "until" attributes to get a list of modified record ids (very slow due to rate limits)
- get_record retrieves a specific RDM record based on the id (quick, uses the RDM REST API)
- query can retrieve a selection of records from the RDM REST API, it is limited to 10K total returned records by RDM/Elasticsearch's configuration
- harvest reads a JSON array of record ids from a file and harvests the RDM records into a dataset v2 collection

`rdmutil` configuration is read either from the envinronment or a JSON formated configuration file. See the [man page](rdmutil.1.md) for details.

### `doi2rdm`

This tool is for exporting metadata from either CrossRef or DataCite and
mapping it into a JSON document suitable to import into Invenio RDM via
RDM's REST API.  See the [man page](doi2rdm.1.md) for details.


## Requirements

- An Invenio RDM deployment
- To building the software and documentation
    - git
    - Go >= 1.20.1
    - Make (e.g. GNU Make)
    - Pandoc >= 3
- For harvesting content
    - [dataset](https://github.com/caltechlibrary/dataset/) >= 2

## Installation

This codebase is speculative. It is likely to change and 
as issues are identified. To install you need to download
the source code and compile it.  Here's the steps I take to
install irdmtools.

~~~
git clone git@github.com:caltechlibrary/irdmtools
cd irdmtools
make
make test
make install
~~~


