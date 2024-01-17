
Action Items
============

Bugs
----

- [ ] rdm2eprint should always populate official URL, in some cases is showing up `/records/{rdmid}` but in others it is populated, when I checked authors record efd3g-p2669 it isn't populated in the JSON output
- [x] Local group items are including an empty "{}" as an entry when retrieved with ep3util (e.g. thesis repository, ep3util get_record 7320)
- [x] progress.go is producing non-sensical estamates of the time remaining, need to review my math (this was a really bad SQL query on my part)
- [ ] eprint2rdm missing metadata attributes
	- [x] journal related fields (e.g. journal:journal) in custom fields
	- [ ] thesis related fields
	- [ ] meeting:meeting in custom fields
	- [x] migrate resolver id (eprint.IDNumber) to metadata.identifiers
	- [ ] Map CaltechTHESIS custom fields, issue #44
	- [ ] Group transfer problem, see issue #42
- [ ] rdmuil
	- [ ] Send to Cummunity not working
	- [ ] Submit Draft not working
    - [ ] Review Drafft not wokring
	- [ ] Submit Publish not working
	- [x] Upload files to Draft
	- [x] Delete files from Draft
	- [ ] Import Files to Draft
	- [ ] Put eprints 'suggests' private notes into review comments, see issue #16
	- [ ] Do file mapping, see issue #3 
- [ ] migrate_record.py (running Python Fixup code)
	- [x] resolver id migrated
	- [ ] migrate DOI to metadata.identifiers if already in repository, issue #15
- [ ] doi2rdm
	- [ ] Migrate monographs from CrossRef more effectively, see issue #40
	- [ ] figure out how to transform mml markup, issue #36


Next
----

- [x] rdmutil get_all_ids needs a get_all_stale_ids counterpart, see issue #68 (implemented get_record_versions"`
- [x] add put_record to rdmutil, actually done as many steps, new_record, new_draft, update_draft, ...
- [x] Implement a CrossRef to Invenio RDM record
- [x] Figure out a faster way to retrieve RDM ids without using the API or OAI-PMH. Possibly options would be to create an rdmapid service, or direct query via PostgreSQL. 
	- PostgREST can provide a RESTful JSON API to our Invenio RDM data stored in Postgres

Someday, maybe
--------------

