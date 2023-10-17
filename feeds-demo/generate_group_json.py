#!/usr/bin/env python3
'''Use the htdocs/group_list.json to build out each of the publication JSON lists'''

import os
import sys
import json
import csv

from py_dataset import dataset
def get_group_list(group_list_json):
    '''Get the group_list.json file and return a useful data structure.'''
    with open(group_list_json, encoding = 'utf-8') as f:
        src = f.read()
        if isinstance(src, bytes):
            src = src.decode('utf-8')
        group_list = json.loads(src)
        return group_list
    return None

def render_combined_json_file(d_name, group_id, obj):

def render_authors_json_files(d_name, group_id, obj):
    '''render the resource JSON files for group_id'''
    # build out the resource type JSON file
    c_name = 'authors'
    repo_id = f'Caltech{c_name.upper()}'
    if repo_id in obj:
        repo_resources = obj[repo_id]
        for resource_type in repo_resources:
            f_name = os.path.join(d_name, f'{resource_type}.json')
            objects = []
            for key in repo_resources[resource_type]:
                obj, err = dataset.read(f'{c_name}.ds', key)
                if err is not None and err != '':
                    print(f'error access {key} in {c_name}.ds, {err}', file = sys.stderr)
                else:
                    objects.append(obj)
            if len(objects) > 0:
                src = json.dumps(objects, indent = 4)
                with open(f_name, 'w', encoding = 'utf-8') as w:
                    w.write(src)
                # Handle the recent sub folder
                recent_d_name = os.path.join(d_name, 'recent')
                if not os.path.exists(recent_d_name):
                    os.makedirs(recent_d_name, mode=0o777, exist_ok =True)
                src = json.dumps(objects[0:25], indent = 4)
                f_name = os.path.join(d_name, 'recent', f'{resource_type}.json')
                with open(f_name, 'w', encoding = 'utf-8') as w:
                    w.write(src)

def render_thesis_json_files(d_name, group_id, obj):
    '''render the resource JSON files for group_id'''
    # build out the resource type JSON file
    c_name = 'thesis'
    repo_id = f'Caltech{c_name.upper()}'
    if repo_id in obj:
        repo_resources = obj[repo_id]
        for resource_type in repo_resources:
            f_name = os.path.join(d_name, f'{resource_type}.json')
            objects = []
            for key in repo_resources[resource_type]:
                obj, err = dataset.read(f'{c_name}.ds', key)
                if err is not None and err != '':
                    print(f'error access {key} in {c_name}.ds, {err}', file = sys.stderr)
                else:
                    objects.append(obj)
            if len(objects) > 0:
                src = json.dumps(objects, indent = 4)
                with open(f_name, 'w', encoding = 'utf-8') as w:
                    w.write(src)
                # Handle the recent sub folder
                recent_d_name = os.path.join(d_name, 'recent')
                if not os.path.exists(recent_d_name):
                    os.makedirs(recent_d_name, mode=0o777, exist_ok =True)
                src = json.dumps(objects[0:25], indent = 4)
                f_name = os.path.join(d_name, 'recent', f'{resource_type}.json')
                with open(f_name, 'w', encoding = 'utf-8') as w:
                    w.write(src)

def render_data_json_files(d_name, group_id, obj):
    '''render the resource JSON files for group_id'''
    # build out the resource type JSON file
    c_name = 'data'
    repo_id = f'Caltech{c_name.upper()}'
    if repo_id in obj:
        repo_resources = obj[repo_id]
        for resource_type in repo_resources:
            f_name = os.path.join(d_name, f'{resource_type}.json')
            objects = []
            for key in repo_resources[resource_type]:
                obj, err = dataset.read(f'{c_name}.ds', key)
                if err is not None and err != '':
                    print(f'error access {key} in {c_name}.ds, {err}', file = sys.stderr)
                else:
                    objects.append(obj)
            if len(objects) > 0:
                src = json.dumps(objects, indent = 4)
                with open(f_name, 'w', encoding = 'utf-8') as w:
                    w.write(src)
                # Handle the recent sub folder
                recent_d_name = os.path.join(d_name, 'recent')
                if not os.path.exists(recent_d_name):
                    os.makedirs(recent_d_name, mode=0o777, exist_ok =True)
                src = json.dumps(objects[0:25], indent = 4)
                f_name = os.path.join(d_name, 'recent', f'{resource_type}.json')
                with open(f_name, 'w', encoding = 'utf-8') as w:
                    w.write(src)


def render_json_files(group_list):
    '''take our agents_csv and agent_pubs_csv filenames and aggregate them'''
    for group_id in group_list:
        if (group_id != '') and (not ' ' in group_id):
            obj = group_list[group_id]
            src = json.dumps(obj, indent=4)
            d_name = os.path.join('htdocs', 'groups', group_id)
            f_name = os.path.join(d_name, 'group.json')
            if not os.path.exists(d_name):
                os.makedirs(d_name, mode=0o777, exist_ok=True)
            with open(f_name, 'w', encoding = 'utf-8') as w:
                w.write(src)
            render_authors_json_files(d_name, group_id, obj)
            render_thesis_json_files(d_name, group_id, obj)
            render_data_json_files(d_name, group_id, obj)
        else:
            print(f'error: "{group_id}" should not have a space', file = sys.stderr)
    return None


def main():
    '''main processing method'''
    app_name = os.path.basename(sys.argv[0])
    if len(sys.argv) != 2:
        print(f'{app_name} expected path to group_list.json file', file = sys.stderr)
        sys.exit(1)
    group_list = get_group_list(sys.argv[1])
    if group_list is None:
        print(f'could not populate group_list from {sys.argv[1]}')
        sys.exit(1)
    err = render_json_files(group_list)
    if err is not None:
        print(f'error: {err}', file = sys.stderr)
        sys.exit(10)

if __name__ == '__main__':
    main()
