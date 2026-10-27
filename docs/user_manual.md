
irdmtools user manual
=====================

Tools
-----

- [rdmutil](rdmutil.1.md) work with Invenio RDM metadata from the command line
- [eprint2rdm](eprint2rdm.1.md) retrieve an EPrint record via the EPrint REST API and return a simplified JSON record almost ready for important into Invenio RDM
- [doi2rdm](doi2rdm.1.md) retrieve a DOI from CrossRef and render as a simplified JSON record almost ready for import into Invenio RDM
- [rdm2eprint](rdm2eprint.1.md) take an RDM record id and return an EPrint record as JSON, so the feeds system can keep working with EPrint structured data
- [eprintrest](eprintrest.1.md) a localhost web service that replicates the EPrints 3.3 REST API from the "archives" directory and the MySQL database, with very fast responses
- [people2vocabulary](people2vocabulary.1.md) transfor a JSON array of Person objects into a YAML vocabularly file suitable for RDM.
- [ep3ds2citations](ep3ds2citations.1.md) convert an EPrint dataset collection to a citations dataset collection.
- [rdmds2citations](rdmds2citations.1.md) convert an RDM dataset collection to a citations dataset collection.
- [Extending RDM with PostgREST](extending-rdm-with-postgrest.md)

Guides
------

- [Migration workflow](Migration-Workflow.md) moving a repository from EPrints to Invenio RDM
- [Setting up RDM metadata replication](setting-up-rdm-replication.md)
- For examples of using irdmtools from Python see the [irdm_harvester](https://github.com/caltechlibrary/irdm_harvester) repository.

Project Status
--------------

- [Issues](https://github.com/caltechlibrary/irdmtools/issues)
- [Releases](https://github.com/caltechlibrary/irdmtools/releases)
- [TODO list](https://github.com/caltechlibrary/irdmtools/blob/main/TODO.md)
