
# Next

## bugs

- [ ] Both thesis and data are not setting up write in group_list.json (this is blocking building json file in individual groups)
- [ ] The combined_data.json is be written by an authors routine, this is wrong
- [ ] The combined_data.json written by data combined via local_groups leaves an empty array and his is wrong. If there are no CaltechDATA items then there should probably be no combined_data.json at all, if there are then it should not be an empty array.

## make_datasets.bash

- [ ] groups.ds and people.ds work towards building group_list.json and people_list.json
    - By combing groups.csv and people.csv along with a CSV files with record group/person from authors, thesis and data we can derive all the other JSON files we need to create

## make_site.bash

- [ ] Create groups/groups_list.json from the full, two pass harvested groups.ds 
- [ ] Create groups/people_list.json from the full, two pass harvested people.ds 

