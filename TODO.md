
Action Items
============

- [x] Drop the 32-bit Raspberry Pi build (`Linux-armv7l`, `GOARM=7`) from the Makefile and any release check the next time this module is worked on (workspace DR-0018, proposed 2026-10-26). 64-bit Raspberry Pi OS stays covered by `Linux-aarch64`. Already-published releases are left as they are.

Project TODOS: <https://github.com/orgs/caltechlibrary/projects/3/views/5?filterQuery=status%3A%22Next%22%2C%22In%20Progress%22%20assignee%3Arsdoiel>

Bugs
----

- [ ] doi2rdm cannot crosswalk a CrossRef journal issue (e.g. 10.1111/jdv.v40.6, `type` "journal-issue"): `resource type "journal-issue" not mapped`. It has no title either. Decide whether to map it in `resource_types` or to skip such records on purpose (found in doi2rdm Part B Phase 2).
- [ ] `CrosswalkCrossRefWork` calls api.ror.org (`lookupROR` in `ror.go`) for each funder whose DOI the publisher asserted, with no timeout and no cache across records, so crosswalking is a network operation and an offline test must drop the funders (found in doi2rdm Part B Phase 2).
- [ ] ep3ds2citations, citations from CaltechDATA seem to be missing orcid even when present in the data.ds record.
- [ ] eprint2rdm missing metadata attributes
	- [x] journal related fields (e.g. journal:journal) in custom fields
	- [ ] thesis related fields
		- [x] `thesis:thesis` wrote the institution under the misspelled key `unversity`, so it was never loaded; now `university` (fixed 2026-11-19 in `eprint2rdm.go`, test `eprint2rdm_thesis_test.go`)
		- [ ] any collection `eprint2rdm` already produced holds the misspelled key `unversity` and lacks the university; re-run it, or rename the key, wherever such a collection is still used (not checked whether the COLD harvest does)
	- [ ] meeting:meeting in custom fields
	- [x] migrate resolver id (eprint.IDNumber) to metadata.identifiers
	- [x] Map CaltechTHESIS custom fields, issue #44
	- [x] Group transfer problem, see issue #42
- [ ] eprint2rdm gets a thesis's degree date and publication date wrong (found 2026-11-19; caltechthesis design brief rulings 12 and 15). For a `date_type` of `degree` (13,248 of 13,250 CaltechTHESIS eprints) the main `date` is the degree date, mostly year-only. `eprint2rdm.go:933-953` defaults `publication_date` to the eprint `datestamp` (when the metadata was last corrected, not when it was first published), overrides it from `date` only when `date_type` is `published` or empty, and writes date type ids `pub_date` and the raw `date_type` (`degree`), which are not RDM stock date types and do not match the local `degree_date`. Fix the logic so a degree-type eprint's `publication_date` comes from `date`, never from `datestamp`, and the degree date is written as `degree_date` with its precision kept. The author's pointer is the Python code in `feeds_v1.6`. A read of it on 2026-11-19 found no degree-date derivation there, only truncation of `publication_date` (`make_people.py:567-569`) and degree-type display labels; `citation.go:544-547` is what sets a thesis's publication date from `date` for `degree`. Read the feeds code the author had in mind before changing anything, and write the test first. Related: irdmtools DR-0010 (a SQL mode that does this for CaltechTHESIS), caltechthesis DR-0002 and DR-0005.
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
- [x] rdm2eprint should always populate official URL, in some cases is showing up `/records/{rdmid}` but in others it is populated, when I checked authors record efd3g-p2669 it isn't populated in the JSON output
- [x] ep3ds2citations, authors array isn't including orcid and clpid as found in authors.ds, thesis.ds and data.ds
- [x] citation record 'id' is winding up with keys like 'authors:authors:XXXXX-XXXXX' for CaltechAUTHORS, CaltechDATA and then 'caltechthesis:thesis:XXXX' for CaltechTHESIS.
- [x] ep3ds2citations, publication_date isn't showing up in JSON stored
- [x] Need an ability to apply an explicit prefix to keys ingested by ep3ds2citation, this avoids the problem where some record pickup an EPrint collection name versus the dataset collection name fallback


Next
----

- [ ] irdmtools Go based commands should not use the RDM JSON API, they should always go direct to the Postgres database
- [x] ep3ds2citation needs to be able to work from a key list or JSON list of keys. When working from key list it should read the entire keylist in then start processing them and display progress
- [x] Integrate a YAML options file into doi2rdm so that we can easily map our customized mapings via configuration instead of hard coding them.
- [x] rdmutil get_all_ids needs a get_all_stale_ids counterpart, see issue #68 (implemented get_record_versions"`
- [x] add put_record to rdmutil, actually done as many steps, new_record, new_draft, update_draft, ...
- [x] Implement a CrossRef to Invenio RDM record
- [x] Figure out a faster way to retrieve RDM ids without using the API or OAI-PMH. Possibly options would be to create an rdmapid service, or direct query via PostgreSQL. 
	- PostgREST can provide a RESTful JSON API to our Invenio RDM data stored in Postgres

Someday, maybe
--------------

- [ ] figure out a faster way to backup stats in RDM other than `elasticdump` which takes a very long time (single three, single CPU)
