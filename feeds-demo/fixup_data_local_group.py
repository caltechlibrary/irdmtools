#!/usr/bin/env python3
'''fixup_data_local_group.py re-maps the local_group data from RDM's contributors to eprint
style local_group for CaltechDATA'''
import os
import sys
import csv

from py_dataset import dataset
import progressbar

def get_groups_csv(groups_csv):
    '''process the groups_csv file returning a list of objects with the preferred group id 
    and a place to holds record ids'''
    object_map = {}
    with open(groups_csv, encoding = 'utf-8', newline = '') as csvfile:
        rows = csv.DictReader(csvfile)
        for obj in rows:
            if 'key' in obj:
                group_id = obj['key']
                # we need to handle looking alternate ids and names too
                for k in [ 'key', 'alternative', 'name' ]:
                    if k in obj:
                        _id = obj[k]
                        object_map[_id] = { 'group_id': group_id, 'record_ids': [] }
        return object_map
    return None

def get_fixup_csv(fixup_csv, object_map):
    '''reading the fixup csv and enhance our object map appropriately'''
    with open(fixup_csv, encoding = 'utf-8', newline = '') as csvfile:
        rows = csv.DictReader(csvfile)
        for row in rows:
            record_id = row['id']
            group_id = row['local_group']
            if group_id in object_map:
                object_map[group_id]['record_ids'].append(record_id)
    return object_map

def merge_local_group_data(app_name, pid, c_name, object_map):
    '''now that we've build up our object map we can update our 
    dataset collection's objects with local group information'''
    tot = len(object_map)
    widgets=[
         f'{app_name} {c_name} (pid:{pid})',
         ' ', progressbar.Counter(), f'/{tot}',
         ' ', progressbar.Percentage(),
         ' ', progressbar.AdaptiveETA(),
    ]
    bar = progressbar.ProgressBar(max_value = tot, widgets=widgets)
    for i, group_id in enumerate(object_map):
        grp_obj = object_map[group_id]
        normalized_id = grp_obj['group_id']
        if len(grp_obj['record_ids']) > 0:
            for record_id in grp_obj['record_ids']:
                obj, err = dataset.read(c_name, record_id)
                if err != '':
                    return f'error: could not get {record_id} from {c_name}, "{err}"'
                if 'local_group' not in obj:
                    obj['local_group'] = {}
                if 'items' not in obj['local_group']:
                    obj['local_group']['items'] = []
                if normalized_id not in obj['local_group']['items']:
                    obj['local_group']['items'].append({'id': normalized_id})
                err = dataset.update(c_name, record_id, obj)
                if err != '':
                    err = f'could not update {record_id} in {c_name}, {err}'
                    return err
        bar.update(i)
    bar.finish()
    return None

def fixup_local_groups(app_name, pid, c_name, groups_csv, fixup_csv):
    '''this is the routine that does the lifting, first builds the list of ids and group
    names using the two CSV files then it runs through the dataset collection updating the
    local_groups attribute.'''
    # Step one read in groups.csv so we know what groups we want to map
    object_map = get_groups_csv(groups_csv)
    if object_map is None:
        return f'something went wrong reading {groups_csv}'

    # merge in the record ids with the groups we're tracking.
    object_map = get_fixup_csv(fixup_csv, object_map)
    if object_map is None:
        return f'something went wrong reading {fixup_csv}'

    # Now we're ready to update our dataset collection records.
    return merge_local_group_data(app_name, pid, c_name, object_map)

# Main process routine
def main():
    '''main processing routine'''
    app_name = os.path.basename(sys.argv[0])
    pid = os.getpid()
    # data.ds, groups.csv, fixup_data_local_groups.csv
    if len(sys.argv) != 4:
        print(f'usage: {app_name} data.ds groups.csv fixup_data_local_groups.csv',
              file = sys.stderr)
        sys.exit(1)
    c_name = sys.argv[1]
    groups_csv = sys.argv[2]
    fixup_csv = sys.argv[3]
    if c_name == "" or os.path.exists(c_name) is False:
        print(f'A valid dataset collection name is required, got "{c_name}"', file = sys.stderr)
        sys.exit(1)
    if groups_csv == "" or os.path.exists(groups_csv) is False:
        print(f'A valid groups CSV file is required, got "{groups_csv}"', file = sys.stderr)
        sys.exit(1)
    if fixup_csv == "" or os.path.exists(fixup_csv) is False:
        print(f'A valid fixup for local groups CSV file is required, got "{fixup_csv}"',
              file = sys.stderr)
        sys.exit(1)
    err = fixup_local_groups(app_name, pid, c_name, groups_csv, fixup_csv)
    if err is not None:
        print(f'error: {err}', file = sys.stderr)
        sys.exit(1)
    print('Success!')

if __name__ == '__main__':
    main()
