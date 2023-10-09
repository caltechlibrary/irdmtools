
# feeds demo

This is a demonstration of generating a feeds.library.caltech.edu
style static website. It uses irdmtools, dataset, datatools,
and Pandoc orchestrated via Bash scripts.

The purpose of this demonstration was to establish a hybrid approach
as we transition from EPrints to RDM repositories. The data structures
used to build the website use an EPrints style record as opposed to RDM.
This is true even when we are harvesting from an RDM repository. It

Dataset used for intermediate representation of the collections from
which the website is generated.  By relying on dataset >= 2.1.6 I am
taking advantage of Postgres database as a JSON store. This has proven
to speed up operations considerably. Rather than the harvesting process
being the biggest time consumer it is now website rendering and copying
to the S3 bucket that takes long. 

The goal of the demo is to show three things:

1. These tools can replicate the existing feeds.library.caltech.edu
2. The entire process completes takes less than four hours
3. The process can be reduced to two phase, "harvesting" and "site rendering"

## dataset collections

The start of the process begins with generating and updating five dataset
collections using Postgres as their JSON storage engine.

Setting up follows the following scenario (Postgres is running on localhost
and only has the current `$USER` to set with appropriate permissions in
Postgres.

~~~
createdb authors
dataset init authods.ds "postgres://$USER@localhost/authors?sslmode=disable"
createdb thesis
dataset init thesis.ds "postgres://$USER@localhost/thesis?sslmode=disable"
createdb data
dataset init data.ds "postgres://$USER@localhost/data?sslmode=disable"
createdb groups
dataset init groups.ds "postgres://$USER@localhost/groups?sslmode=disable"
createdb people
dataset init people.ds "postgres://$USER@localhost/people?sslmode=disable"
~~~

Next you need to create shell environments for authors, thesis and data.
These need to be available from the following files (they get sourced
by `make_dataset.bash` and `make_sites.bash`): 

- authors.env (RDM)
- thesis.env (EPrints)
- data.env (RDM)

For the RDM based repositories you need to set the following environment variables
(NOTE: I am relying on Postgres access for this demo, it's faster than making requests
through the REST API).

REPO_ID
: This is the repository database name, e.g. "caltechauthors", "caltechdata" at
Caltech Library.

C_NAME
: This is the dataset collection name used, e.g. "authors.ds", "data.ds" in our
example

RDM_DB_USER
: This is the Postgres database user for the RDM Postgres database

RDM_DB_HOST
: This is the host used to access the RDM Postgres database

I recommend dumping your production RDM instance's Postgres database
and loading it locally to isolate the demo from your production deployment.
Dumping and reloading the Postgres database is relatively fast. This is actually
the approach I plan to take when we update our production feeds system. Like
with our Postgres based repositories I recommend dumping and loading your
EPrints database locally to avoid burden on your production system.

I use a similar set of environment variables for harvesting our remaining EPrints
repository.

REPO_ID
: This is the repository database, e.g. "caltechthesis" at Caltech Library.

C_NAME
: This is the dataset collection name used, e.g. "thesis.ds" in our example

EPRINT_DB_HOST
: The hostname for the MySQL database holding the EPrints database

EPRINT_DB_USER
: The MySQL user id allowed to access the EPrints database

EPRINT_DB_PASSWORD
: The MySQL password for the user allowed to access the EPrints database

These environment variables are used per repository to harvest content.
They are required for `make_datasets.bash` to function properly. They are
used by ep3tuil, rdm2eprint and rdmutil in the content retrieval process.


## Required software

- irdmtools >= 0.0.57 (use the latest release)
- dataset >= 2.1.6 (use the latest release)
- datatools >= 1.2.5 (use the latest release)
- Bash >= 3 (or equivalent POSIX shell)
- Pandoc >= 3
- Postgres >= 14
- Python >= 3.10

Recommended if you are deploying your site to an S3 bucket.

- s5cmd >= 2.2 (see https://github.com/peak/s5cmd)

## Installation and setup

1. Clone github.com/caltechlibrary/irdmtools
2. Copy the feeds demo directory to where you want to stage things on your system and change into it
3. Install the latest tools
    a. `curl -L https://caltechlibrary.github.io/dataset/installer.sh | sh`
    b. `curl -L https://caltechlibrary.github.io/irdmtools/installer.sh | sh`
    c. `curl -L https://caltechlibrary.github.io/datatools/installer.sh | sh`
4. Make sure Postgres is installed (see https://postgres.org for details)
5. Make sure Pandoc is installed (see https://pandoc.org for details)
    a. Test to confirm it is running and installed
    b. Create an appropriate account if necessary with admin privileges
6. Create the needed environment files as described above, e.g. authors.env, data.env and thesis.env
7. Change into the feeds-demo directory
    
At this point you should be able to run the following scripts to harvest
and build the feeds

1. ./setup_datasets.bash (only need to run this the first time)
2. ./make_datasets.bash (this is run each to to refresh our collections)
3. ./make_site.bash (this is done to stage the website from current state of databases)
4. If you are deploying to an S3 bucket use the s5cmd to copy/sync your site

