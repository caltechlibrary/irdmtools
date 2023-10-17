
# Next

## bugs

- [x] Both thesis and data are not setting up correctly in group_list.json (this is blocking building json file in individual groups)
- [x] The combined_data.json is be written by an authors function, this is wrong
- [ ] The combined_data.json written by processing data via local_groups leaves an empty array and his is wrong. If there are no CaltechDATA items then there should probably be no combined_data.json at all, if there are then it should not be an empty array.
- [ ] `combined*.json` should move into generate_local_group.py since we are generating the individual resource type JSON files for authors, thesis and data

## make_datasets.bash

- [ ] groups.ds and people.ds work towards building group_list.json and people_list.json
    - By combing groups.csv and people.csv along with a CSV files with record group/person from authors, thesis and data we can derive all the other JSON files we need to create
    - cleanup fixup_data_local_groups.py, make sure the structure is `.local_groups.items[*].id` and that I can get the strings out without quoting

## make_site.bash

- [ ] Create groups/groups_list.json from the full, two pass harvested groups.ds 
- [ ] Create groups/people_list.json from the full, two pass harvested people.ds 

