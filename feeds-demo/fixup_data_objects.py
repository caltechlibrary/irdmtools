#!/usr/bin/env python3
'''fixup_data_people.py maps the CaltechDATA creator's orcid to a cl_people_id using people.csv.'''
import os
import sys
import csv
import json

from py_dataset import dataset
import progressbar

def read_json_file(src_name):
    '''read a JSON file and return a JSON object'''
    print(f'Reading {src_name}', file = sys.stderr)
    src = ''
    with open(src_name, 'r', encoding = 'utf-8') as _f:
        src = _f.read()
        if isinstance(src, bytes):
            src= src.decode('utf-8')
    if src == '':
        print(f'failed to read {src_name}', file = sys.stderr)
        sys.exit(1)
    return json.loads(src)

def write_json_file(f_name, objects):
    '''render the JSON file from the objects.'''
    src = json.dumps(objects, indent = 4)
    print(f'Writing {f_name}', file = sys.stderr)
    with open(f_name, 'w', encoding = 'utf-8') as _w:
        _w.write(src)

def make_orcid_to_people_id_map(people_csv):
    '''process the people_csv file returning a list of objects with the preferred group id 
    and a place to holds record ids'''
    object_map = {}
    with open(people_csv, encoding = 'utf-8', newline = '') as csvfile:
        rows = csv.DictReader(csvfile)
        for obj in rows:
            cl_people_id = obj.get('cl_people_id', None)
            orcid = obj.get('orcid', None)
            if (cl_people_id is not None) and (orcid is not None):
                object_map[orcid] = cl_people_id
        return object_map
    return None


def fixup_data_objects(app_name, pid, people_csv, data_objects_json):
    '''read in people_csv and building an oricid to cl_poeple_id map, then
    reading our data_objects_json, then apply to each object in
    data_objects_json and write them out'''
    orcid_to_people_id = make_orcid_to_people_id_map(people_csv)
    data_objects = read_json_file(data_objects_json)
    tot = len(data_objects)
    widgets=[
         f'update {data_objects_json} with cl_people_id' 
         ' ', progressbar.Counter(), f'/{tot}',
         ' ', progressbar.Percentage(),
         ' ', progressbar.AdaptiveETA(),
    ]
    bar = progressbar.ProgressBar(max_value = tot, widgets=widgets)
    for i, obj in enumerate(data_objects):
        orcid = obj.get('orcid', None)
        if (orcid is not None) and (orcid in orcid_to_people_id):
            cl_people_id = orcid_to_people_id[orcid]
            obj['cl_people_id'] = cl_people_id
        resource_id = obj.get('resource_id', None)
        if resource_id is not None:
            obj['href'] = f'https://data.caltech.edu/records/{resource_id}'
        bar.update(i)
    bar.finish()
    write_json_file(data_objects_json, data_objects)

# Main process routine
def main():
    '''main processing routine'''
    app_name = os.path.basename(sys.argv[0])
    pid = os.getpid()
    # data.ds, groups.csv, fixup_data_local_groups.csv
    if len(sys.argv) != 3:
        print(f'usage: {app_name} people.csv htdocs/people/data_objects.json',
              file = sys.stderr)
        sys.exit(1)
    people_csv = sys.argv[1]
    data_objects_json = sys.argv[2]
    if people_csv == "" or os.path.exists(people_csv) is False:
        print(f'A valid people CSV file is required, got "{people_csv}"', file = sys.stderr)
        sys.exit(1)
    if data_objects_json == "" or os.path.exists(data_objects_json) is False:
        print(f'A valid data_objects.json file is required, got "{data_objects_json}"',
              file = sys.stderr)
        sys.exit(1)
    err = fixup_data_objects(app_name, pid, people_csv, data_objects_json)
    if err is not None:
        print(f'error: {err}', file = sys.stderr)
        sys.exit(1)
    print('Success!')

if __name__ == '__main__':
    main()
