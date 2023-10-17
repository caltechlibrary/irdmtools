#!/usr/bin/env python3
'''fixup_data_local_group.py re-maps the local_group data from RDM's contributors to eprint
style local_group for CaltechDATA'''
import os
import sys
import csv

from py_dataset import dataset
import progressbar

def reset_local_group_data(app_name, pid, c_name):
    '''reset the local_group attribute for all records'''
    keys = dataset.keys(c_name)
    tot = len(keys)
    widgets=[
         f'{app_name} {c_name} (pid:{pid})',
         ' ', progressbar.Counter(), f'/{tot}',
         ' ', progressbar.Percentage(),
         ' ', progressbar.AdaptiveETA(),
    ]
    bar = progressbar.ProgressBar(max_value = tot, widgets=widgets)
    for i, record_id in enumerate(keys):
        obj, err = dataset.read(c_name, record_id)
        if err != '':
            return f'error: could not get {record_id} from {c_name}, "{err}"'
        if 'local_group' in obj:
            del obj['local_group']
        err = dataset.update(c_name, record_id, obj)
        if err != '':
            err = f'could not update {record_id} in {c_name}, {err}'
            return err
        bar.update(i)
    bar.finish()
    return None

# Main process routine
def main():
    '''main processing routine'''
    app_name = os.path.basename(sys.argv[0])
    pid = os.getpid()
    # data.ds, groups.csv, fixup_data_local_groups.csv
    if len(sys.argv) != 2:
        print(f'usage: {app_name} data.ds',
              file = sys.stderr)
        sys.exit(1)
    c_name = sys.argv[1]
    if c_name == "" or os.path.exists(c_name) is False:
        print(f'A valid dataset collection name is required, got "{c_name}"', file = sys.stderr)
        sys.exit(1)
    err = reset_local_group_data(app_name, pid, c_name)
    if err is not None:
        print(f'error: {err}', file = sys.stderr)
        sys.exit(1)
    print('Success!')

if __name__ == '__main__':
    main()
