#!/usr/bin/env python3
'''Use the htdocs/group_list.json to build out each of the publication JSON lists'''

import os
import sys
import json
import csv
import operator
from subprocess import Popen, PIPE, TimeoutExpired

from py_dataset import dataset

import yaml


def get_group_list(group_list_json):
    '''Get the group_list.json file and return a useful data structure.'''
    print(f'Writing {group_list_json}', file = sys.stderr)
    with open(group_list_json, encoding = 'utf-8') as _f:
        src = _f.read()
        if isinstance(src, bytes):
            src = src.decode('utf-8')
        group_list = json.loads(src)
        return group_list
    return None

def _retrieve_keys(csv_name, group_id):
    '''used by render_combined_files(), return a list of keys'''
    keys = []
    with open(csv_name, 'r', encoding = 'utf-8', newline = '') as csvfile:
        _r = csv.DictReader(csvfile)
        for row in _r:
            if ('local_group' in row) and (group_id == row['local_group']):
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

def pandoc_write_file(f_name, objects, template, title = None, from_fmt = None, to_fmt = None):
    '''render the objects to a markdown file using template'''
    if len(objects) == 0:
        return f'pandoc_write_file({f_name}, objects, {template}): no objects to write'
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
    src = ('\n'.join(['---', yaml.dump(objects), '---'])).encode('utf-8')
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
    if resource is None:
        return None
    if repository is not None:
        resource['repository'] = repository
    if href is not None:
        resource['href'] = href
    if resource_type is not None:
        resource['resource_type'] = resource_type
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
                repository, href, mk_label(resource_type), item))
        else:
            objects['content'].append(pandoc_enhance_item(None, None, None, item))
    return objects


def write_markdown_resource_file(f_name, base_object, resource):
    '''write a group resource page by transform our list of objects'''
    p_objects = pandoc_build_resource(base_object, resource)
    err = pandoc_write_file(f_name, p_objects, 'templates/groups-group-resource.md')
    if err is not None:
        print(f'pandoc error: {err}', file = sys.stderr)

def mk_label(val):
    '''make a label from an id string'''
    if '_' in val:
        val = val.replace('_', ' ', -1)
    return val.title()

def flatten_combined_resource_types(repository, href, combined, resource_object):
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

def build_combined_object(group_id, group):
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
    for k in group:
        if k in [ 'CaltechAUTHORS', 'CaltechTHESIS', 'CaltechDATA' ]:
            # flatten nested dict
            for resource in flatten_combined_resource_types(k, u_map[k], c_map[k], group[k]):
                resources.append(resource)
        else:
            obj[k] = group[k]
    obj['content'] = resources
    return obj

def read_json_file(src_name):
    '''read a JSON file and return a JSON object'''
    print(f'Reading {src_name}', file = sys.stderr)
    src = ''
    with open(src_name, 'r', encoding = 'utf-8') as _f:
        src = _f.read()
        if isinstance(src, bytes):
            src= src.decode('utf-8')
    if src == '':
        print(f'failed to read {f_name}', file = sys.stderr)
        sys.exit(1)
    return json.loads(src)

def write_markdown_index_file(f_name, group_id, group):
    '''coorodiante the write of a combined markdown file'''
    obj = build_combined_object(group_id, group)
    err = pandoc_write_file(f_name, obj,
        "templates/groups-group-index.md", 
        from_fmt = 'markdown',
        to_fmt = 'markdown')
    if err is not None:
        print(f'pandoc error: {err}', file = sys.stderr)

def render_combined_files(repo, d_name, group_id):
    '''render a combined json file'''
    c_name = f'{repo}.ds'
    csv_name = os.path.join('htdocs', 'groups', f'group_{repo}.csv')
    keys = _retrieve_keys(csv_name, group_id)
    objects = []
    for key in keys:
        obj, err = dataset.read(c_name, key)
        if err is not None and err != '':
            return f'error access {key} in {c_name}.ds, {err}'
        objects.append(enhance_object(obj))
    if len(objects) == 0:
        #print(f'DEBUG no objects found for {group_id} in {d_name}, {repo}', file = sys.stderr)
        return None
    # sort the list of objects
    objects.sort(key=operator.itemgetter('date', 'title'))
    o_name = 'combined'
    if repo == 'thesis':
        o_name = 'combined_thesis'
    elif repo == 'data':
        o_name = 'combined_data'
    f_name = os.path.join(d_name, o_name + '.json')
    # Write the combined JSON file for the repository
    write_json_file(f_name, objects)
    # Write the group index Markdown file for the repository
    src_name = os.path.join(d_name, "group.json")
    group = read_json_file(src_name)
    f_name = os.path.join(d_name, 'index.md')
    write_markdown_index_file(f_name, group_id, group)

    # We don't created a combined thesis for recent, doesn't make sense.
    if repo in [ 'authors', 'data' ]:
        # Handle the recent subfolder
        recent_d_name = os.path.join(d_name, 'recent')
        if not os.path.exists(recent_d_name):
            os.makedirs(recent_d_name, mode=0o777, exist_ok =True)
        f_name = os.path.join(recent_d_name, o_name + '.json')
        # Write the recent combined JSON file for the repository
        write_json_file(f_name, objects[0:25])
        # Write the combined Markdown file for the repository
        f_name = os.path.join(recent_d_name, 'index.md')
        write_markdown_index_file(f_name, group_id, group)
    return None


def render_authors_files(d_name, obj, group_id = None, people_id = None):
    '''render the resource JSON files for group_id'''
    # build out the resource type JSON file
    c_name = 'authors'
    repo_id = f'Caltech{c_name.upper()}'
    resource_info =  {
        "repository": repo_id,
        "href":"https://authors.library.caltech.edu"
    }
    if repo_id in obj:
        repo_resources = obj[repo_id]
        merge_fields = [ "name", "description", "activity", "updated", 
                        "start", "end", "date", "updated", "alternative", 
                        "approx_start", "approx_end", "pi" ]
        for val in merge_fields:
            if val in obj:
                resource_info[val] = obj[val]
            elif val in resource_info:
                del resource_info[val]
        for resource_type in repo_resources:
            resource_info["resource_type"] = resource_type
            resource_info["resource_label"] = mk_label(resource_type)
            objects = []
            for key in repo_resources[resource_type]:
                obj, err = dataset.read(f'{c_name}.ds', key)
                if err is not None and err != '':
                    print(f'error access {key} in {c_name}.ds, {err}', file = sys.stderr)
                else:
                    objects.append(enhance_object(obj))
            if len(objects) > 0:
                # Write the group resource files out
                f_name = os.path.join(d_name, f'{resource_type}.json')
                write_json_file(f_name, objects)
                # Setup to write Markdown files
                if group_id is not None:
                    resource_info["group_id"] = group_id
                    resource_info["group_label"] = mk_label(group_id)
                if people_id is not None:
                    resource_info["people_id"] = people_id
                    resource_info["people_label"] = mk_label(people_id)
                # Write out Markdown files via Pandoc
                f_name = os.path.join(d_name, f'{resource_type}.md')
                write_markdown_resource_file(f_name, resource_info, objects)
                # Handle the recent sub folder
                recent_objects = objects[0:25]
                recent_d_name = os.path.join(d_name, 'recent')
                if not os.path.exists(recent_d_name):
                    os.makedirs(recent_d_name, mode=0o777, exist_ok =True)
                f_name = os.path.join(d_name, 'recent', f'{resource_type}.json')
                write_json_file(f_name, recent_objects)
                # Write out Markdown files via Pandoc
                f_name = os.path.join(d_name, 'recent', f'{resource_type}.md')
                write_markdown_resource_file(f_name, resource_info, recent_objects)

def render_thesis_files(d_name, obj, group_id = None, people_id = None):
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
                    objects.append(enhance_object(obj))
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
                if group_id is not None:
                    resource_info["group_id"] = group_id
                    resource_info["group_label"] = mk_label(group_id)
                if people_id is not None:
                    resource_info["people_id"] = people_id
                    resource_info["people_label"] = mk_label(people_id)
                # Write out Markdown files via Pandoc
                f_name = os.path.join(d_name, f'{resource_type}.md')
                write_markdown_resource_file(f_name, resource_info, objects)
# CUT NOTE: recent thesis 25 doesn't make sense since degrees are clusterred by class year
##                  # Handle the recent sub folder
##                  recent_d_name = os.path.join(d_name, 'recent')
##                  if not os.path.exists(recent_d_name):
##                      os.makedirs(recent_d_name, mode=0o777, exist_ok =True)
##                  f_name = os.path.join(d_name, 'recent', f'{resource_type}.json')
##                  write_json_file(f_name, objects[0:25])

def render_data_files(d_name, obj, group_id = None, people_id = None):
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
                    objects.append(enhance_object(obj))
            if len(objects) > 0:
                # Write JSON resource file
                write_json_file(f_name, objects)
                # Handle the recent sub folder
                recent_d_name = os.path.join(d_name, 'recent')
                if not os.path.exists(recent_d_name):
                    os.makedirs(recent_d_name, mode=0o777, exist_ok =True)
                write_json_file(f_name, objects[0:25])
                resource_info =  {
                        "repository": "CaltechDATA", 
                        "href":"https://data.library.caltech.edu",
                        "resource_type": resource_type,
                        "resource_label": mk_label(resource_type)
                    }
                if group_id is not None:
                    resource_info["group_id"] = group_id
                    resource_info["group_label"] = mk_label(group_id)
                if people_id is not None:
                    resource_info["people_id"] = people_id
                    resource_info["people_label"] = mk_label(people_id)
                # Write out Markdown files via Pandoc
                f_name = os.path.join(d_name, 'recent', f'{resource_type}.json')
                write_markdown_resource_file(f_name, resource_info, objects)


def render_files(group_list):
    '''take our agents_csv and agent_pubs_csv filenames and aggregate them'''
    for group_id in group_list:
        if (group_id != '') and (not ' ' in group_id):
            obj = group_list[group_id]
            src = json.dumps(obj, indent=4)
            d_name = os.path.join('htdocs', 'groups', group_id)
            f_name = os.path.join(d_name, 'group.json')
            if not os.path.exists(d_name):
                os.makedirs(d_name, mode=0o777, exist_ok=True)
            print(f'Writing {f_name}', file = sys.stderr)
            with open(f_name, 'w', encoding = 'utf-8') as _w:
                _w.write(src)
            render_authors_files(d_name, obj, group_id = group_id)
            render_thesis_files(d_name, obj, group_id = group_id)
            render_data_files(d_name, obj, group_id = group_id)
            for repo in [ "authors", "thesis", "data" ]:
                err = render_combined_files(repo, d_name, group_id)
                if err is not None:
                    print(
                    f'error: render_combined_files({repo}' +
                    f', {d_name}, {group_id}) -> {err}', file = sys.stderr)
        else:
            print(f'error: "{group_id}" should not have a space', file = sys.stderr)


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
    render_files(group_list)

if __name__ == '__main__':
    main()
