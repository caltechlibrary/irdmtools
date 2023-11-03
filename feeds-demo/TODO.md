# Next

- [ ] convert generate_people_resource_files.py to work on people/resource
- [ ] create generate_people_combined_files.py from generate_people_resource_files.py
- [ ] create aggregate_people.py to work like aggregate_groups.py

## bugs

- [ ] On individual people pages CaltechTHESIS and CaltechDATA items aren't being listed at all, e.g. Wennberg-P-O, could be corsswalk from ORCID to clpid is broken
- [ ] publish.bash still needs tuning for numworkers using s5cmd so it doesn't fail with reset
- [ ] People resource types to labels needs improvements for handling plurals given types, also some underscores aren't being mapped to spaces
- [x] Groups, GALCIT, combined thesis sort order needs to be reversed (newest first)
- [ ] The resource types in the recent feeds need vetting, "Software" shows up under "CaltechAUTHORS" but appears to be pointing at CaltechDATA content, I need to double check the "Thesis" listed under CaltechAUTHORS and make sure they areally are non-CaltechAUTHOR thesis (the citation shows Caltech as publisher, but that might be bad crosswalk data)
- [x] Recent pages the links to the current page because the `official_url` attribute is not popluated. Also authors_list, pub_year, doi, record id (rdmid) is not being populated either. I need to look at how I am enhancing the items in the content array, results can be viewed in `/recent/*.md` files for resources (e.g. article, monograph, etc). 
- [x] The title in Markdown generated via Pandoc seem to wrap at a specific column, might be happening as a result of PyYAML wrapping strings, need to sort out as it is causing a problem in linking as implememented in citation.md
- [x] `htdocs/groups/<GROUP_ID>/<COMBINED>.md` is not getting rendered from `htdocs/groups/<GROUP_ID>/<COMBINED>.json`
- [x] `htdocs/recent/<RESOURCE_TYPE>.md` is not getting rendered from `htdocs/recent/<RESOURCE_TYPE>.json`
- [x] group resource pages isn't including resource type in the H2 heading above the UL list of citations
- [ ] ciations in resource pages and recent pages aren't complete
    - [x] Move citations to a partial template
    - [ ] Improve the partial template citation.md
- [x] Both thesis and data are not setting up correctly in group_list.json (this is blocking building json file in individual groups)
- [x] The combined_data.json is be written by an authors function, this is wrong
- [x] The combined_data.json written by processing data via local_groups leaves an empty array and his is wrong. If there are no CaltechDATA items then there should probably be no combined_data.json at all, if there are then it should not be an empty array.
- [x] `combined*.json` should move into generate_local_group.py since we are generating the individual resource type JSON files for authors, thesis and data
- [ ] page.tmpl needs to be enhance so the page title is meaningful and will improve the useful for pagefind search results for the site.
    - [x] Pages need to include group, people or resource in title
    - [ ] Group pages should have title plus "resource type" for sub pages
    - [x] People pages should have title plus "resource type" for sub pages
    - [ ] Recent pages should have title plus "resource type" for sub pages

## make_datasets.bash

- [ ] Refactor, code to generate "combined" lists need a different sort than group_list.json and people_resources.json provide, implement group_combined.json, people_combined.json via dsquery or python program, tease this code out from `generate_*_files.py`. 
    - [x] Rename `generate_group_files.py` to `generate_group_resource_files.py`
    - [x] Create `generate_group_combined_files.py`
    - [ ] Rename `generate_*_files.py` to `generate_*_resource_files.py`
    - [ ] Create `generate_*_combined_files.py`
- [x] figure out how to update people.ds counts, this is probable faster to be done via an export of counts and cl_people_id from authors.ds, data.ds and thesis.ds then merge the the result into people.ds
    - [x] authors_count
    - [x] editor_count
    - [x] thesis_count
    - [x] advisor_count
    - [x] data_count
    - [x] committee_count
- [x] groups.ds and people.ds work towards building group_list.json and people_list.json
    - By combing groups.csv and people.csv along with a CSV files with record group/person from authors, thesis and data we can derive all the other JSON files we need to create
    - cleanup fixup_data_local_groups.py, make sure the structure is `.local_groups.items[*].id` and that I can get the strings out without quoting

## make_site.bash

- [ ] Missing combined for people resource pages
- [ ] Missing RSS rendering for all feeds
- [ ] Missing BibTeX rendering for all feeds
    - [x] Write a simple BibTeX template that iterates over "content"
    - [ ] Need to add this to generate_people_files.py and generate_groups_files.py and the one that generates the recent lists
- [x] Look at performance issues in make_group_pages and see if pushing this processing into generate_group_json.py might improve things (at least elimate some loops)
- [x] Create groups/groups_list.json from the CSV files for group's repos and groups.csv
- [x] Create groups/people_list.json from the CSV files for people's repos and people.csv

## CL.js improvements, fixes

- [x] Remove lunrs JS files, site uses PageFind and that is my recommendation going forward os there is no "Search" widget possible
- [ ] Test CL.js scripts using test buckets, may need to make the base url for content configurable when rendering static to htdocs tree
    - [ ] consider a CL-config.js file that everything uses to pickup appropiate configurtion (e.g. production or testing)


