
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

## Required software

- dataset (latest release)
- py_dataset (latest release)
- irdmtools (latest release)
- datatools (reldate is required)
- Bash >= 3 (or equivalent POSIX shell)
- python >= 3.10 (including pip for the version you have installed)
- pandoc >= 3
- Postgres >= 14
- s5cmd >= 2.2 (see https://github.com/peak/s5cmd)

## Installation and setup

1. Clone github.com/caltechlibrary/irdmtools
2. Copy the feeds demo directory to where you want to stage things on your system and change into it
3. Install the latest tools
    a. `curl -L https://caltechlibrary.github.io/dataset/installer.sh | sh`
    b. `curl -L https://caltechlibrary.github.io/irdmtools/installer.sh | sh`
    c. `curl -L https://caltechlibrary.github.io/datatools/installer.sh | sh`
4. Install needed Python repositories
    a. Change into the the feeds folder you copied
    b. `python -m pip install -r requirements.txt`
5. Make sure Postgres is installed (see https://postgres.org for details)
6. Make sure Pandoc is installed (see https://pandoc.org for details)
    a. Test to confirm it is running and installed
    b. Create an appropriate account if neccessary with admin provilleges
7. Run setup_database.bash 
8. Create the needed environment files, e.g. authors.env, data.env and thesis.env
    
At this point you should be able to run the following scripts to harvest
and build the feeds

1. ./setup_databases.bash (only need to run this the first time)
2. ./make_datasets.bash (this is run each to to refresh the data)
3. ./make_site.bash (this is done to stage the website from current state of databases)
4. Use the s5cmd to copy/sync to your S3 bucket

The primise is to recreate feeds.library.caltech.edu (version 1) by first harvesting an RDM implementation of our EPrints 
repositories into a dataset collection(s) using the EPrints datastruct crosswalked from RDM. This is done with rdmutil to
get a list of keys in the RDM repository, then uses rdm2eprints to harvest the contents into a dataset collection. Note the
dataset collections are defined to use Postgres as the JSON store running on localhost.  The next step is to generate the 
directory structure and populate all the JSON files used to build the site.  This is done by ysing dsquery taking advantage
of Postgres's SQL dialect to create lists of JSON objects that make up feeds. 
