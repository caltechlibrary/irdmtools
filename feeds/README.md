
# feeds demo

This is a demonstration of generating a feeds like static website 
using current irdmtools and dataset version v2.1.4 or better.

The primise is we can substantially decrease the build time if we leverage
dataset collections usingthe SQL storage enginge (e.q. Postgres and SQLite3).
The general approach is to split the whole process in half. A process which
populates or updates the dataset collections as performed by make_datasets.bash
and a second process called make_site.bash which renders the htdocs directory
tree based on the dataset collections being processed.

Ideally you should run one process then the next but if we leverage a SQL store
such as Postgres we can actually run them concurrently since the database engines takes care of resolving writes and reads cleanly. Likewise you could have a "full" harvest running update our dataset collections and initiate a recent harvest along side it. I do not recommend running these processes concurrently but they should not be self destructive if they happen to run concurrently due to the process isolation provided by using Postgres as the storage engine.

This approach makes use of five dataset collections initialized using Postgres
as a the storage engine.

- authors.ds holds EPrint shaped RDM record content
- data.ds holds EPrint shaped RDM record content
- thesis.ds holds actual EPrint content
- groups.ds holds our groups metadata based on groups.csv
- people.ds holds our people metadata based on people.sv

The cloned copes of authors, data and thesis retain their old names in the htdocs tree, i.e. "CaltechAUTHORS.ds", "CaltechDATA.ds", and "CaltechTHESIS.ds".




The primise is to recreate feeds.library.caltech.edu (version 1) by first harvesting an RDM implementation of our EPrints 
repositories into a dataset collection(s) using the EPrints datastruct crosswalked from RDM. This is done with rdmutil to
get a list of keys in the RDM repository, then uses rdm2eprints to harvest the contents into a dataset collection. Note the
dataset collections are defined to use Postgres as the JSON store running on localhost.  The next step is to generate the 
directory structure and populate all the JSON files used to build the site.  This is done by ysing dsquery taking advantage
of Postgres's SQL dialect to create lists of JSON objects that make up feeds. 
