
# Next

## bugs

- [ ] `htdocs/groups/<GROUP_ID>/<COMBINED>.md` is not getting rendered from `htdocs/groups/<GROUP_ID>/<COMBINED>.json`
- [ ] `htdocs/recent/<RESOURCE_TYPE>.md` is not getting rendered from `htdocs/recent/<RESOURCE_TYPE>.json`
- [ ] group resource pages isn't including resource type in the H2 heading above the UL list of citations
- [ ] ciations in resource pages and recent pages aren't complate, look at using a partial template for this to make this consistent across renderings
- [x] Both thesis and data are not setting up correctly in group_list.json (this is blocking building json file in individual groups)
- [x] The combined_data.json is be written by an authors function, this is wrong
- [x] The combined_data.json written by processing data via local_groups leaves an empty array and his is wrong. If there are no CaltechDATA items then there should probably be no combined_data.json at all, if there are then it should not be an empty array.
- [x] `combined*.json` should move into generate_local_group.py since we are generating the individual resource type JSON files for authors, thesis and data
- [ ] page.tmpl needs to be enhance so the page title is meaningful and will improve the useful for pagefind search results for the site.

## make_datasets.bash

- [x] groups.ds and people.ds work towards building group_list.json and people_list.json
    - By combing groups.csv and people.csv along with a CSV files with record group/person from authors, thesis and data we can derive all the other JSON files we need to create
    - cleanup fixup_data_local_groups.py, make sure the structure is `.local_groups.items[*].id` and that I can get the strings out without quoting

## make_site.bash

- [ ] Look at performance issues in make_group_pages and see if pushing this processing into generate_group_json.py might improve things (at least elimate some loops)
- [x] Create groups/groups_list.json from the CSV files for group's repos and groups.csv
- [ ] Create groups/people_list.json from the CSV files for people's repos and people.csv

