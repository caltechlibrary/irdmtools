
Institutional Repository Data Management Tools
==============================================

This is a proof of concept Go package fork working with Invenio-RDM's JSON API as well as via direct access to the Postgres database backing the Invenio-RDM instance.

Requirements
------------

- Go >= 1.20.1
- Make (e.g. GNU Make)
- Pandoc >= 3
- Git

Installation
------------

This codebase is speculative as a proof of concept. Installation requires
downloading the source code and compiler. Here's the steps I take to
install irdmtools.

~~~
git clone git@github.com:caltechlibrary/irdmtools
cd irdmtools
make
make test
make install
~~~


