#!/usr/bin/env python3
'''Use the htdocs/recent/*.json to build out each of the publication type markdown files'''

import os
import sys
import json
import csv
import operator
from subprocess import Popen, PIPE, TimeoutExpired

from py_dataset import dataset
import progressbar
import yaml


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
    print(f'DEBUG enhance object -> {obj}', file = sys.stderr)
    if 'date' in obj:
        obj['pub_year'] = obj['date'][0:4]
    if ('creators' in obj) and ('items' in obj['creators']):
        _l = format_authors(obj['creators']['items'])
        if _l is not None:
            obj['author_list'] = _l
    return obj

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
    print(f'Writing {f_name}')
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
    resource = enhance_object(resource)
    print(f'DEBUG enhance resource for Pandoc -> {resource}', file = sys.stderr)
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

def mk_label(val):
    '''make a label from an id string'''
    if '_' in val:
        val = val.replace('_', ' ', -1)
    return val.title()

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
                repository, href, mk_label(resource_type), item))
        else:
            objects['content'].append(pandoc_enhance_item(None, None, None, item))
    return objects

def write_markdown_resource_file(f_name, base_object, resource):
    '''write a recent resource page by transform our list of objects'''
    p_objects = pandoc_build_resource(base_object, resource)
    err = pandoc_write_file(f_name, p_objects, 'templates/recent-resource.md')
    if err is not None:
        print(f'pandoc error: {err}', file = sys.stderr)

def write_markdown_combined_file(f_name, base_object, resource_list):
    '''coorodiante the write of a combined markdown file'''
    repository = base_object.get('repository', None)
    href = base_object.get('href', None)
    objects = base_object
    objects['content'] = []
    resource_type = None
    for item in resource_list:
        r_type = item.get('resource_type', None)
        if resource_type != r_type:
            resource_type = r_type
            objects['content'].append(pandoc_enhance_item(
                repository, href, resource_type, item))
        else:
            objects['content'].append(pandoc_enhance_item(
                None, None, None, item))

    err = pandoc_write_file(f_name, objects, 'templates/recent-combined.md',
        { 'from_fmt': 'markdown', 'to_fmt': 'markdown' })
    if err is not None:
        print(f'pandoc error: {err}', file = sys.stderr)

def render_authors():
    # render individual resource files
    f_name = os.path.join('htdocs', 'recent', 'object_types.json')
    resource_types = read_json_file(f_name)  
    obj = { 
        'repository': 'CaltechAUTHORS',
        'href': 'https://authors.library.caltech.edu'
    }
    for resource in resource_types:
        obj['resource_type'] = resource.get('name', None)
        f_name = os.path.join('htdocs', 'recent', resource['name'] + '.json')
        resource_list = read_json_file(f_name)
        f_name = os.path.join('htdocs', 'recent', resource['name'] + '.md')
        write_markdown_resource_file(f_name, obj, resource_list)
    # render combined resource file
    del obj['resource_type']
    f_name = os.path.join('htdocs', 'recent', 'combined.json')
    resource_list = read_json_file(f_name)
    f_name = os.path.join('htdocs', 'recent', 'combined.md')
    write_markdown_combined_file(f_name, obj, resource_list)



def render_data():
    # render individual resource files
    f_name = os.path.join('htdocs', 'recent', 'data_object_types.json')
    resource_types = read_json_file(f_name)  
    obj = { 
        'repository': 'CaltechDATA',
        'href': 'https://data.caltech.edu'
    }
    for resource in resource_types:
        obj['resource_type'] = resource.get('name', None)
        f_name = os.path.join('htdocs', 'recent', resource['name'] + '.json')
        resource_list = read_json_file(f_name)
        f_name = os.path.join('htdocs', 'recent', resource['name'] + '.md')
        write_markdown_resource_file(f_name, obj, resource_list)
    # render combined resource file
    del obj['resource_type']
    f_name = os.path.join('htdocs', 'recent', 'combined_data.json')
    resource_list = read_json_file(f_name)
    f_name = os.path.join('htdocs', 'recent', 'combined_data.md')
    write_markdown_combined_file(f_name, obj, resource_list)


def render_recent():
    # Render authors' recent
    render_authors()
    # Recent data's recent
    render_data()
    
def main():
    '''main processing method'''
    app_name = os.path.basename(sys.argv[0])
    argc = len(sys.argv)
    if argc != 1:
        print(f'{app_name} expected only the app name to render htdocs/recent/*', file = sys.stderr)
        sys.exit(1)
    render_recent()

if __name__ == '__main__':
    main()
