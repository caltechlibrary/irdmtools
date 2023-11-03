#!/usr/bin/env python3
'''Use the htdocs/people/people_list.json to build out each of the publication JSON lists'''

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

def enhance_object(obj):
    '''given an eprint like record, enhance the record to make it Pandoc template friendly'''
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
    print(f'DEBUG Popen {cmd}', file = sys.stderr)
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
        print(f'error: {out}', file = sys.stderr)
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

def write_markdown_combined_file(f_name, repo, people, objects):
    '''coorodiante the write of a combined markdown file'''
    obj = build_combined_object(repo, people, objects)
    err = pandoc_write_file(f_name, obj,
        "templates/peoples-people-combined.md", 
        { 'from_fmt': 'markdown', 'to_fmt': 'markdown' })
    if err is not None:
        print(f'pandoc error: {err}', file = sys.stderr)

def render_combined_files(repo, d_name, people_id, people):
    '''render a combined json file'''
    c_name = f'{repo}.ds'
    csv_name = os.path.join('htdocs', 'people', f'people_{repo}.csv')
    keys = _retrieve_keys(csv_name, people_id)
    objects = []
    for key in keys:
        obj, err = dataset.read(c_name, key)
        if err is not None and err != '':
            return f'error access {key} in {c_name}.ds, {err}'
        objects.append(enhance_object(obj))
    if len(objects) == 0:
        return None
    o_name = 'combined'
    if repo == 'thesis':
        o_name = 'combined_thesis'
    elif repo == 'data':
        o_name = 'combined_data'
    f_name = os.path.join(d_name, o_name + '.json')
    # Write the combined JSON file for the repository
    write_json_file(f_name, objects)

    # Write  the combined Markdown filefile
    f_name = os.path.join(d_name, o_name + '.md')
    write_markdown_combined_file(f_name, repo, people, objects)
    return None


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
    if (people_id == '') and (' ' in people_id):
        print(f'error: "{people_id}" is not valid', file = sys.stderr)
        return
    src = json.dumps(obj, indent=4)
    # We make the directory since we have a Caltech Person
    d_name = os.path.join('htdocs', 'people', people_id)
    if not os.path.exists(d_name):
        os.makedirs(d_name, mode=0o777, exist_ok=True)
    
    # render the combined*.md files
    for repo in [ "authors", "thesis", "data" ]:
        #print(f'DEBUG rending combined files: {repo}', file = sys.stderr)
        err = render_combined_files(repo, d_name, people_id, obj)
        if err is not None:
            print(
            f'error: render_combined_files({repo}' +
            f', {d_name}, {people_id}) -> {err}', file = sys.stderr)



def map_resources(cl_people_id, person, authors_objects, thesis_objects, data_objects):
    '''map resources from repositories into the people_list'''
    if len(authors_objects) > 0:
        person['CaltechAUTHORS'] = authors_objects
    if len(thesis_objects) > 0:
        person['CaltechTHESIS'] = thesis_objects
    if len(data_objects) > 0:
        person['CaltechDATA'] = data_objects
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
    # Map authors, thesis and data objects into people_list
    people_list = map_people_list(people_list, author_objects, thesis_objects, data_objects)
    if people_list is None:
        print('mapping of authors, thesis and data objects tailed', file = sys.stderr)
        sys.exit(10)
    # Write out the mapping of people combined json with authors, thesis and data resources
    #f_name = os.path.join('htdocs', 'people', 'people_combined.json')
    #print(f'Writing {f_name}', file = sys.stderr)
    #write_json_file(f_name, people_list)

    tot = len(people_list)
    widgets=[
         f'render_a_people ' 
         ' ', progressbar.Counter(), f'/{tot}',
         ' ', progressbar.Percentage(),
         ' ', progressbar.AdaptiveETA(),
    ]
    bar = progressbar.ProgressBar(max_value = tot, widgets=widgets)
    print(f'DEBUG looping through ({tot}) people_list', file = sys.stderr)
    for i, cl_people_id in enumerate(people_list):
        print(f'DEBUG i {i}, cl_people_id {cl_people_id}', file = sys.stderr)
        if (people_id is None) or (cl_people_id == people_id):
            render_a_person(cl_people_id, people_list[cl_people_id])
        bar.update(i)
    bar.finish()

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