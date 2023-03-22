
Institutional Repository Data Management Tools
==============================================

This is a proof of concept set tools and Go packages for
working with institutional repositories. Initial target is
Invenio RDM's REST API and OAI-PMH service.

The concept is being developed around a cli called `rdmutil` which
can talk the web APIs supported by Invenio RDM. Caltech Library is using
this prototype in our migration from EPrints to Invenio RDM and for
content migration, content analysis and aggregation.

Tools
-----

### `rdmutil`

- get_all_ids uses the OAI-PMH service to harvest all the current record ids in an Invenio RDM instance (very slow due to rate limits)
- get_modified_ids uses the OAI-PMH service with the "from" and "until" attributes to get a list of modified record ids (very slow due to rate limits)
- get_record retrieves a specific RDM record based on the id (quick, uses the RDM REST API)
- query can retrieve a selection of records from the RDM REST API, it is limited to 10K total returned records by RDM/Elasticsearch's configuration
- harvest reads a JSON array of record ids from a file and harvests the RDM records into a dataset v2 collection

`rdmutil` configuration is read either from the envinronment or a JSON formated configuration file. See the [man page](rdmutil.1.md) for details.

Requirements
------------

- Go >= 1.20.1
- Make (e.g. GNU Make)
- Pandoc >= 3
- git
- SQLite3, MySQL 8 or PostgreSQL 14

This codebase is speculative as a proof of concept. Installation requires
downloading the source code and compiler. Here's the steps I take to
install irdmtools.

### Installation

~~~
git clone git@github.com:caltechlibrary/irdmtools
cd irdmtools
make
make test
make install
~~~


