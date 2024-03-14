
Action Items
============

Bugs
----

- [ ] rdm2eprint should always populate official URL, in some cases is showing up `/records/{rdmid}` but in others it is populated, when I checked authors record efd3g-p2669 it isn't populated in the JSON output
- [ ] eprint2rdm missing metadata attributes
	- [x] journal related fields (e.g. journal:journal) in custom fields
	- [ ] thesis related fields
	- [ ] meeting:meeting in custom fields
	- [x] migrate resolver id (eprint.IDNumber) to metadata.identifiers
	- [x] Map CaltechTHESIS custom fields, issue #44
	- [x] Group transfer problem, see issue #42
- [x] Local group items are including an empty "{}" as an entry when retrieved with ep3util (e.g. thesis repository, ep3util get_record 7320)
- [x] progress.go is producing non-sensical estamates of the time remaining, need to review my math (this was a really bad SQL query on my part)
- [x] rdmutil
	- [x] Send to Cummunity not working
	- [x] Submit Draft not working
    - [x] Review Drafft not wokring
	- [x] Submit Publish not working
	- [x] Upload files to Draft
	- [x] Delete files from Draft
	- [x] Import Files to Draft
	- [x] Put eprints 'suggests' private notes into review comments, see issue #16
	- [x] Do file mapping, see issue #3 
- [x] migrate_record.py (running Python Fixup code)
	- [x] resolver id migrated
	- [x] migrate DOI to metadata.identifiers if already in repository, issue #15
- [x] doi2rdm
	- [x] Migrate monographs from CrossRef more effectively, see issue #40
	- [x] figure out how to transform mml markup, issue #36


Next
----

- [ ] irdmtools Go based commands should not use the RDM JSON API, they should all be used to directly access the Postgres database
- [x] Integrate a YAML options file into doi2rdm so that we can easily map our customized mapings via configuration instead of hard coding them.
- [x] rdmutil get_all_ids needs a get_all_stale_ids counterpart, see issue #68 (implemented get_record_versions"`
- [x] add put_record to rdmutil, actually done as many steps, new_record, new_draft, update_draft, ...
- [x] Implement a CrossRef to Invenio RDM record
- [x] Figure out a faster way to retrieve RDM ids without using the API or OAI-PMH. Possibly options would be to create an rdmapid service, or direct query via PostgreSQL. 
	- PostgREST can provide a RESTful JSON API to our Invenio RDM data stored in Postgres

Someday, maybe
--------------

- [ ] figure out a faster way to backup stats in RDM other than `elasticdump` which takes a very long time (single three, single CPU)
