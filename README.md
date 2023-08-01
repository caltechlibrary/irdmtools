
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

This tool is for interacting with an Invenio RDM repository via RDM's
REST and OAI-PMH API. It covers most the JSON API documented at <https://inveniordm.docs.cern.ch/>. This includes listing, submitting and managing records and draft records.

`rdmutil` configuration is read either from the envinronment or a JSON formated configuration file.  See the [man page](rdmutil.1.md) for details.

### `eprint2rdm`

This tool is migrating content from an EPrints repository via the EPrint REST API.  It will retrieve an EPrint XML representation of the EPrint record  and transform it into a JSON encded simplified record nearly compatible with Invenio RDM.  See the [man page](eprint2rdm.1.md) for details.

### `doi2rdm`

This tool will query the CrossRef API and convert a works record into a JSON structure compatible with an RDM record (e.g. to be inserted via an RDM API call).  See the [man page](eprint2rdm.1.md) for details.

## Requirements

- An Invenio RDM deployment
- To building the software and documentation
    - git
    - Go >= 1.20.4
    - Make (e.g. GNU Make)
    - Pandoc >= 3
- For harvesting content
    - [dataset](https://github.com/caltechlibrary/dataset/) >= 2

## Quick install

If you're running on Linux, macOS or Raspberry Pi OS you may be able to installed irdmtools with the following curl command

~~~
curl https://caltechlibrary.github.io/irdmtools/installer.sh | sh
~~~

## Installation from source

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


