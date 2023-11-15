#!/usr/bin/env python3
'''Use the htdocs/people/people_list.json to build out each of the publication JSON lists for each resource type'''

import os
import sys
import json
import csv
import operator
from subprocess import Popen, PIPE, TimeoutExpired

from py_dataset import dataset
import progressbar
import yaml


def get_people_list(people_list_json):
    '''Get the people_list.json file and return a useful data structure.'''
    print(f'Reading {people_list_json}', file = sys.stderr)
    with open(people_list_json, encoding = 'utf-8') as _f:
        src = _f.read()
        if isinstance(src, bytes):
            src = src.decode('utf-8')
        people_list = json.loads(src)
        return people_list
    return None


def format_authors(creators):
    '''format the authors to be friendly to Pandoc template'''
    if len(creators) > 0:
        authors = []
        for i, creator in enumerate(creators):
            if 'name' in creator:
                family_name, given_name = '', ''
                if 'family' in creator['name']:
                    family_name = creator['name']['family']
                if 'given' in creator['name']:
                    given_name = creator['name']['given']
                if given_name != '':
                    authors.append(f'{family_name}, {given_name}')
                else:
                    authors.append(f'{family_name}')
                if i > 3:
                    break
        if len(authors) == 1:
            return authors[0]
        if len(authors) == 2:
            return ' and '.join(authors[0:2])
        if len(authors) > 2:
            return '; '.join(authors[0:2]) + '; et el.'
    return None

def enhance_object(obj, repo_url = None):
    '''given an eprint like record, enhance the record to make it Pandoc template friendly'''
    obj_id = obj.get('id', None)
    if (repo_url is not None) and obj_id is None:
       obj.set('id', f'{repo_url}{obj_id}')
    if 'type' in obj and 'resource_type' not in obj:
        obj['resource_type'] = obj['type']
    if 'date' in obj:
        obj['pub_year'] = obj['date'][0:4]
    if ('creators' in obj) and ('items' in obj['creators']):
        _l = format_authors(obj['creators']['items'])
        if _l is not None:
            obj['author_list'] = _l
    return obj

def write_json_file(f_name, objects):
    '''render the JSON file from the objects.'''
    src = json.dumps(objects, indent = 4)
    print(f'Writing {f_name}', file = sys.stderr)
    with open(f_name, 'w', encoding = 'utf-8') as _w:
        _w.write(src)


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


def map_authors(cl_people_id, authors_objects):
    '''for a given cl_people_id map the resources identified from list of objects'''
    m = {}
    for obj in authors_objects:
        author_id = obj.get('cl_people_id', None)
        if author_id == cl_people_id:
            resource_id = obj.get('resource_id', None)
            resource_type = obj.get('resource_type', None)
            if (resource_type is not None) and (resource_id is not None):
                if resource_type not in m:
                    m[resource_type] = []
                m[resource_type].append(resource_id)
    return m


def map_thesis(cl_people_id, thesis_objects):
    '''for a given cl_people_id map the resources identified from list of objects'''
    m = {}
    for obj in thesis_objects:
        author_id = obj.get('cl_people_id', None)
        if author_id == cl_people_id:
            resource_id = obj.get('resource_id', None)
            resource_type = obj.get('thesis_type', None)
            if (resource_type is not None) and (resource_id is not None):
                if resource_type not in m:
                    m[resource_type] = []
                m[resource_type].append(resource_id)
    return m

def map_data(cl_people_id, data_objects):
    '''for a given cl_people_id map the resources identified from list of objects'''
    m = {}
    for obj in data_objects:
        author_id = obj.get('cl_people_id', None)
        if author_id == cl_people_id:
            resource_id = obj.get('resource_id', None)
            resource_type = obj.get('resource_type', None)
            if (resource_type is not None) and (resource_id is not None):
                if resource_type not in m:
                    m[resource_type] = []
                m[resource_type].append(resource_id)
    return m
                    
def map_resources(cl_people_id, person, authors_objects, thesis_objects, data_objects):
    '''map resources from repositories into the people_list'''
    r_map = map_authors(cl_people_id, authors_objects)
    if len(r_map) > 0:
        person['CaltechAUTHORS'] = r_map
    r_map = map_thesis(cl_people_id, thesis_objects)
    if len(r_map) > 0:
        person['CaltechTHESIS'] = r_map
    r_map = map_data(cl_people_id, data_objects)
    if len(r_map) > 0:
        person['CaltechDATA'] = r_map
    return person

def map_people_list(people_list, authors_objects, thesis_objects, data_objects):
    '''map_people_list takes the JSON array and turns it into a dict'''
    m = {}
    print('mapping people list with authors, thesis and data resources (takes a while)', file = sys.stderr)
    tot = len(people_list)
    widgets=[
         f'map people_list to people_resources.json' 
         ' ', progressbar.Counter(), f'/{tot}',
         ' ', progressbar.Percentage(),
         ' ', progressbar.AdaptiveETA(),
    ]
    bar = progressbar.ProgressBar(max_value = tot, widgets=widgets)
    for i, person in enumerate(people_list):
        cl_people_id = person.get('cl_people_id', None)
        if (cl_people_id is None) or (cl_people_id == '') or (' ' in cl_people_id):
            #print(f'problem cl_people_id ({i}) -> {person}, skipping')
            continue
        m[cl_people_id] = map_resources(cl_people_id, person, authors_objects, thesis_objects, data_objects)
        bar.update(i)
    bar.finish()
    return m

def render_peoples(people_list, people_id = None):
    '''take our CSV and JSON files and aggregate them'''
    ### #FIXME: Need to enhance the person objects with record data from each repository
    # Load authors_objects.json
    f_name = os.path.join('htdocs', 'people', 'authors_objects.json')
    author_objects = read_json_file(f_name)
    if author_objects is None:
        print(f'failed to read author objects from {f_name}', file = sys.stderr)
        sys.exit(10)
    # Load thesis_objects.json
    f_name = os.path.join('htdocs', 'people', 'thesis_objects.json')
    thesis_objects = read_json_file(f_name)
    if thesis_objects is None:
        print(f'failed to read thesis objects from {f_name}', file = sys.stderr)
        sys.exit(10)
    # Load data_objects.json
    f_name = os.path.join('htdocs', 'people', 'data_objects.json')
    data_objects = read_json_file(f_name)
    if data_objects is None:
        print(f'failed to read data objects from {f_name}', file = sys.stderr)
        sys.exit(10)
    # Map authors, thesis and data objects into people_list
    people_list = map_people_list(people_list, author_objects, thesis_objects, data_objects)
    if people_list is None:
        print('mapping of authors, thesis and data objects tailed', file = sys.stderr)
        sys.exit(10)
    # Write out the mapping of people_list with authors, thesis and data resources
    f_name = os.path.join('htdocs', 'people', 'people_resources.json')
    write_json_file(f_name, people_list)

def main():
    '''main processing method'''
    app_name = os.path.basename(sys.argv[0])
    argc = len(sys.argv)
    if (argc < 2) or (argc > 3):
        print(f'{app_name} expected path to people_list.json file', file = sys.stderr)
        sys.exit(1)
    people_list = get_people_list(sys.argv[1])
    if people_list is None:
        print(f'could not populate people_list from {sys.argv[1]}')
        sys.exit(1)
    people_id = None
    if argc == 3:
        people_id = sys.argv[2]
    render_peoples(people_list, people_id)

if __name__ == '__main__':
    main()
