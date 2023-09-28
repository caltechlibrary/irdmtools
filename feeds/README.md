
# feeds demo

This is a demonstration of generating a feeds like static website using current irdmtools and dataset version v2.1.4 or better.
The primise is to recreate feeds.library.caltech.edu (version 1) by first harvesting an RDM implementation of our EPrints 
repositories into a dataset collection(s) using the EPrints datastruct crosswalked from RDM. This is done with rdmutil to
get a list of keys in the RDM repository, then uses rdm2eprints to harvest the contents into a dataset collection. Note the
dataset collections are defined to use Postgres as the JSON store running on localhost.  The next step is to generate the 
directory structure and populate all the JSON files used to build the site.  This is done by ysing dsquery taking advantage
of Postgres's SQL dialect to create lists of JSON objects that make up feeds. 
