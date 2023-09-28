#!/usr/bin/env python3
'''csv_to_dataset.py converts CSV files for groups and people to dataset collections'''

import sys
import os
import csv
from py_dataset import dataset

def csv_to_dataset(c_name, csv_name):
    '''Converts a CSV file to a dataset collection'''
    has_errors = False
    with open(csv_name, newline = '', encoding = 'utf-8') as csvfile:
        reader = csv.DictReader(csvfile)
        for i, row in enumerate(reader):
            key = None
            if 'key' in row:
                key = row['key']
                del row['key']
            elif 'cl_people_id' in row:
                key = row['cl_people_id']
                row['_Key'] = key
            else:
                print(f'row {i}, failed to find key for row', file = sys.stderr)
                has_errors = True
            if key is not None:
                err = None
                if dataset.has_key(c_name, key):
                    err = dataset.update(c_name, key, row)
                else:
                    err = dataset.create(c_name, key, row)
                if err is not None and err != '':
                    has_errors = True
                    print(f'row {i}, failed to add record {key}, {err}', file = sys.stderr)
    if has_errors:
        return 'some records failed to import'
    return None

#
# Main procesing
#

def main():
    '''main processing functions to avoid global variable collisions'''
    app_name = os.path.basename(sys.argv[0])
    args = sys.argv[1:]
    if len(args) != 2:
        print(f'usage: {app_name} C_NAME CSV_NAME', file = sys.stderr)
        sys.exit(10)

    err = csv_to_dataset(args[0], args[1])
    if err is not None:
        print(f'error: {err}')
        sys.exit(10)

if __name__ == '__main__':
    main()
