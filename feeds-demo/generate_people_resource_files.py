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

def _retrieve_keys(csv_name, people_id):
    '''used by render_combined_files(), return a list of keys'''
    keys = []
    with open(csv_name, 'r', encoding = 'utf-8', newline = '') as csvfile:
        _r = csv.DictReader(csvfile)
        for row in _r:
            if ('cl_people_id' in row) and (people_id == row['cl_people_id']):
                keys.append(row['id'])
    return keys

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
    #print(f'DEBUG obj_id -> {obj_id}, repo_url {repo_url}', file = sys.stderr)
    if (repo_url is not None) and (obj_id is not None):
        if not obj_id.startswith('https://'):
            obj['id'] = f'{repo_url}{obj_id}'
            #print(f'DEBUG enhanced id -> {obj["id"]}', file = sys.stderr)
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

def pandoc_write_file(f_name, objects, template, params = None):
    '''render the objects to a markdown file using template'''
    if params is None:
        title, from_fmt, to_fmt = None, None, None
    else:
        title = params.get('title', None)
        from_fmt = params.get('from_fmt', None)
        to_fmt = params.get('to_fmt', None)
    if len(objects) == 0:
        return f'pandoc_write_file({f_name}, objects, {template}, {params}): no objects to write'
    # We'll assume YAML to feed to Pandoc, we set default flow style to false to avoid wrapping titles
    src = ('\n'.join(['---', yaml.dump(objects), '---'])).encode('utf-8')
    cmd = [ "pandoc", '--template', template, '-o', f_name ]
    if title is not None:
        cmd.append('--metadata')
        cmd.append(f'title={title}')
    if to_fmt is not None:
        cmd.append('-t')
        cmd.append(to_fmt)
    if from_fmt is not None:
        cmd.append('-f')
        cmd.append(from_fmt)
        if from_fmt == 'json':
            src = json.dumps(objects).encode('utf--8')
    print(f'Writing {f_name}', file = sys.stderr)
    with Popen(cmd, stdin = PIPE, stdout = PIPE, stderr = PIPE) as proc:
        try:
            out, errs = proc.communicate(src, timeout = 60)
        except TimeoutExpired:
            proc.kill()
            out, errs = proc.communicate()
        if out != b'':
            print(f'{out}', file = sys.stderr)
        if errs != b'':
            print(f'error: {errs}', file = sys.stderr)
            sys.exit(20)
    return None

def pandoc_enhance_item(repository = None, href = None, resource_type = None, resource = None):
    '''given a resource, enhance it to make it friendly to tempalte in Pandoc'''
    if resource is None:
        return None
    if repository is not None:
        resource['repository'] = repository
    if href is not None:
        resource['href'] = href
    if resource_type is not None:
        resource['resource_type'] = resource_type
        if resource_type.endswith('s'):
            resource['resource_label'] = mk_label(resource_type)
        else:
            resource['resource_label'] = mk_label(resource_type) + 's'
    return resource


def pandoc_build_resource(base_object, resource_list):
    '''build a object structure suitable for processing with a pandoc template from
    a base object and a list of object resources'''
    objects = base_object
    if 'repository' not in objects:
        print('missing repository value in object expression', file = sys.stderr)
        return None
    repository = objects['repository']
    if 'href' not in objects:
        print('missing href value in object expression', file = sys.stderr)
        return None
    href = objects['href']
    if 'resource_type' in objects:
        resource_type = objects['resource_type']
    else:
        resource_type = None
    objects['content'] = []
    for i, item in enumerate(resource_list):
        if i == 0:
            objects['content'].append(pandoc_enhance_item(
                repository, href, resource_type, item))
        else:
            objects['content'].append(pandoc_enhance_item(None, None, None, item))
    return objects


def write_markdown_resource_file(f_name, base_object, resource):
    '''write a people resource page by transform our list of objects'''
    p_objects = pandoc_build_resource(base_object, resource)
    err = pandoc_write_file(f_name, p_objects, 'templates/peoples-people-resource.md')
    if err is not None:
        print(f'pandoc error: {err}', file = sys.stderr)

def mk_label(val):
    '''make a label from an id string'''
    if val is None:
        raise ValueError('mk_label(val) failed, val cannot be None')  
        return None
    if '_' in val:
        val = val.replace('_', ' ', -1)
    return val.title()

def flatten_index_resource_types(repository, href, combined, resource_object):
    '''flatten takes a dictionary pointing at an array and flattens it into an
    array suitable for processing with Pandoc templates'''
    objects = []
    for i, resource_type in enumerate(resource_object):
        if i == 0:
            objects.append({'repository': repository,
                'href': href, 
                'combined': combined, 
                'label': mk_label(resource_type), 
                'resource_type': resource_type})
        else:
            objects.append({'label': mk_label(resource_type), 'resource_type': resource_type})
    return objects

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

def build_index_object(people):
    '''build an object form the decoded array'''
    # Merge the two objects into a Pandoc friendly structure
    obj = {}
    resources = []
    u_map = {
        "CaltechAUTHORS": "https://authors.library.caltech.edu",
        "CaltechDATA": "https://data.caltech.edu",
        "CaltechTHESIS": "https://thesis.library.caltech.edu"
    }
    c_map = {
        "CaltechAUTHORS": "combined",
        "CaltechDATA": "combined_data",
        "CaltechTHESIS": "combined_thesis"
    }
    for k in people:
        if k in [ 'CaltechAUTHORS', 'CaltechTHESIS', 'CaltechDATA' ]:
            # flatten nested dict
            for resource in flatten_index_resource_types(k, u_map[k], c_map[k], people[k]):
                resources.append(resource)
        else:
            obj[k] = people[k]
    obj['content'] = resources
    return obj

def write_markdown_index_file(f_name, people):
    '''coorodiante the write of a index markdown file'''
    obj = build_index_object(people)
    err = pandoc_write_file(f_name, obj,
        "templates/peoples-people-index.md", 
        { 'from_fmt': 'markdown', 'to_fmt': 'markdown' })
    if err is not None:
        print(f'pandoc error: {err}', file = sys.stderr)

def build_combined_object(repo, people, objects):
    '''build an object form the decoded array'''
    # Merge the two objects into a Pandoc friendly structure
    obj = {}
    for k in people:
        if k not in [ 'CaltechAUTHORS', 'CaltechDATA', 'CaltechTHESIS' ]:
            obj[k] = people[k]
    # flatten nested dict
    u_map = {
        "CaltechAUTHORS": "https://authors.library.caltech.edu",
        "CaltechDATA": "https://data.caltech.edu",
        "CaltechTHESIS": "https://thesis.library.caltech.edu"
    }
    repo_name = 'Caltech' + repo.upper()
    obj['repository'] = repo_name
    repo_url = u_map.get(repo_name, None)
    if repo_url is not None:
        obj['href'] = repo_url
    obj['content'] = objects
    return obj


def build_repo_resource_objects(c_name, repo_url, resource_type, repo_resources):
    '''generates a list of objects from resource type and repository resources'''
    objects = []
    for key in repo_resources[resource_type]:
        obj, err = dataset.read(f'{c_name}.ds', key)
        if err is not None and err != '':
            print(f'error access {key} in {c_name}.ds, {err}', file = sys.stderr)
        else:
            objects.append(enhance_object(obj, repo_url = repo_url))
    return objects

def render_authors_files(d_name, obj, people_id = None):
    '''render the resource JSON files for people_id'''
    # build out the resource type JSON file
    c_name = 'authors'
    repo_id = 'CaltechAUTHORS'
    repo_url = "https://authors.library.caltech.edu"
    resource_info =  {
        "repository": repo_id,
        "href": repo_url
    }
    if repo_id in obj:
        repo_resources = obj[repo_id]
        for resource_type in repo_resources:
            objects = build_repo_resource_objects(c_name, repo_url, resource_type, repo_resources)
            if len(objects) > 0:
                # Write the people resource files out
                f_name = os.path.join(d_name, f'{resource_type}.json')
                write_json_file(f_name, objects)
                # Setup to write Markdown files
                if people_id is not None:
                    resource_info["people_id"] = people_id
                    resource_info["people_label"] = mk_label(people_id)
                if people_id is not None:
                    resource_info["people_id"] = people_id
                    resource_info["people_label"] = mk_label(people_id)
                # Write out Markdown files via Pandoc
                f_name = os.path.join(d_name, f'{resource_type}.md')
                write_markdown_resource_file(f_name, resource_info, objects)

def render_thesis_files(d_name, obj, people_id = None):
    '''render the resource JSON files for people_id'''
    # build out the resource type JSON file
    c_name = 'thesis'
    repo_id = 'CaltechTHESIS'
    repo_url = 'https://thesis.library.caltech.edu'
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
                    objects.append(enhance_object(obj, repo_url = repo_url))
            if len(objects) > 0:
                # Handle writing files
                write_json_file(f_name, objects)
                # setup for Markdown files
                resource_info =  {
                        "repository": "CaltechTHESIS", 
                        "href":"https://thesis.library.caltech.edu",
                        "resource_type": resource_type,
                        "resource_label": mk_label(resource_type)
                    }
                if people_id is not None:
                    resource_info["people_id"] = people_id
                    resource_info["people_label"] = mk_label(people_id)
                if people_id is not None:
                    resource_info["people_id"] = people_id
                    resource_info["people_label"] = mk_label(people_id)
                # Write out Markdown files via Pandoc
                f_name = os.path.join(d_name, f'{resource_type}.md')
                write_markdown_resource_file(f_name, resource_info, objects)


def render_data_files(d_name, obj, people_id = None):
    '''render the resource JSON files for people_id'''
    # build out the resource type JSON file
    c_name = 'data'
    repo_id = 'CaltechDATA'
    repo_url = 'https://data.caltech.edu'
    data_count = obj.get('data_count', 0)
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
                    objects.append(enhance_object(obj, repo_url = repo_url))
            if len(objects) > 0:
                # Write JSON resource file
                write_json_file(f_name, objects)
                # Handle the recent sub folder
                resource_info =  {
                        "repository": "CaltechDATA", 
                        "href":"https://data.caltech.edu",
                        "resource_type": resource_type,
                        "resource_label": mk_label(resource_type)
                    }
                if people_id is not None:
                    resource_info["people_id"] = people_id
                    resource_info["people_label"] = mk_label(people_id)
                if people_id is not None:
                    resource_info["people_id"] = people_id
                    resource_info["people_label"] = mk_label(people_id)
                f_name = os.path.join(d_name, f'{resource_type}.md')
                #print(f'DEBUG writing Markdown for {people_id} of {f_name} -> '+json.dumps(objects, indent = 4), file = sys.stderr)
                write_markdown_resource_file(f_name, resource_info, objects)
    elif data_count != 0:
        print(f'something went wrong for {people_id}, data count {data_count}', file = sys.stderr)

def people_has_content(people):
    '''check to make sure it makes sense to render the people. There should
    be some type of records available to populate feeds'''
    if 'CaltechAUTHORS' in people:
        return True
    if 'CaltechTHESIS' in people:
        return True
    if 'CaltechDATA' in people:
        return True
    if ('description' in people) and (people['description'] != ''):
        return True
    return False


def render_a_person(people_id, obj):
    '''render a specific people's content if valid'''
    authors_count = obj.get('authors_count', 0)
    thesis_count = obj.get('thesis_count', 0)
    data_count = obj.get('data_count', 0)
    if authors_count == 0:
        print(f'"{people_id}" has no CaltechAUTHORS content, skipping', file = sys.stderr)
        return
    if (people_id == '') and (' ' in people_id):
        print(f'error: "{people_id}" is not valid', file = sys.stderr)
        return
    src = json.dumps(obj, indent=4)
    # We make the directory since we have a Caltech Person
    d_name = os.path.join('htdocs', 'people', people_id)
    if not os.path.exists(d_name):
        os.makedirs(d_name, mode=0o777, exist_ok=True)
    # Write out the people json file
    f_name = os.path.join(d_name, 'people.json')
    write_json_file(f_name, obj)
    
    # Now render the repo resource files.
    render_authors_files(d_name, obj, people_id = people_id)
    if thesis_count > 0:
        render_thesis_files(d_name, obj, people_id = people_id)
    if data_count > 0:
        render_data_files(d_name, obj, people_id = people_id)

    # Now render the people index.json file
    f_name = os.path.join(d_name, 'index.json')
    write_json_file(f_name, obj)

    # Now render the people index.md file
    f_name = os.path.join(d_name, 'index.md')
    write_markdown_index_file(f_name, obj)


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
         f'map people_list' 
         ' ', progressbar.Counter(), f'/{tot}',
         ' ', progressbar.Percentage(),
         ' ', progressbar.AdaptiveETA(),
    ]
    bar = progressbar.ProgressBar(max_value = tot, widgets=widgets)
    for i, person in enumerate(people_list):
        cl_people_id = person.get('cl_people_id', None)
        if (cl_people_id is None) or (cl_people_id == '') or (' ' in cl_people_id):
            print(f'problem cl_people_id ({i}) -> {person}, skipping')
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
    # Load the map of authors, thesis and data objects from people_resources.json
    f_name = os.path.join('htdocs', 'people', 'people_resources.json')
    if os.path.exists(f_name):
        people_list = read_json_file(f_name)
    else:
        people_list = map_people_list(people_list, author_objects, thesis_objects, data_objects)
        # Write out the mapping of people_list with authors, thesis and data resources
        write_json_file(f_name, people_list)
    if people_list is None:
        print('mapping of authors, thesis and data objects failed, aborting', file = sys.stderr)
        sys.exit(10)
    for cl_people_id in people_list:
        if (people_id is None) or (cl_people_id == people_id):
            render_a_person(cl_people_id, people_list[cl_people_id])

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
