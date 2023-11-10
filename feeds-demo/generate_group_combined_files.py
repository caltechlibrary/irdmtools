#!/usr/bin/env python3
'''Use the htdocs/groups/group_list.json to build out each of the publication JSON lists'''

import os
import sys
import json
import csv
import operator
from subprocess import Popen, PIPE, TimeoutExpired

from py_dataset import dataset

import progressbar
import yaml


def get_group_combined_list(group_list_json):
    '''Get the group_list.json file and return a useful data structure.'''
    print(f'Reading {group_list_json}', file = sys.stderr)
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
    #print(f'Writing {f_name}', file = sys.stderr)
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
    #print(f'Writing {f_name}')
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


def mk_label(val):
    '''make a label from an id string'''
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

def build_index_object(group):
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
            for resource in flatten_index_resource_types(k, u_map[k], c_map[k], group[k]):
                resources.append(resource)
        else:
            obj[k] = group[k]
    obj['content'] = resources
    return obj

def write_markdown_index_file(f_name, group):
    '''coorodiante the write of a index markdown file'''
    obj = build_index_object(group)
    err = pandoc_write_file(f_name, obj,
        "templates/groups-group-index.md", 
        { 'from_fmt': 'markdown', 'to_fmt': 'markdown' })
    if err is not None:
        print(f'pandoc error: {err}', file = sys.stderr)

def build_combined_object(repo, group, objects):
    '''build an object form the decoded array'''
    #print(f'DEBUG repo -> {repo}', file = sys.stderr)
    # Merge the two objects into a Pandoc friendly structure
    obj = {}
    for k in group:
        if k not in [ 'CaltechAUTHORS', 'CaltechDATA', 'CaltechTHESIS' ]:
            obj[k] = group[k]
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

def write_markdown_combined_file(f_name, repo, group, objects):
    '''coorodiante the write of a combined markdown file'''
    obj = build_combined_object(repo, group, objects)
    err = pandoc_write_file(f_name, obj,
        "templates/groups-group-combined.md", 
        { 'from_fmt': 'markdown', 'to_fmt': 'markdown' })
    if err is not None:
        print(f'pandoc error: {err}', file = sys.stderr)

def render_combined_files(repo, d_name, group_id, group):
    '''render a combined json file'''
    c_name = f'{repo}.ds'
    csv_name = os.path.join('htdocs', 'groups', f'group_{repo}_combined.csv')
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
    # sort the list of objects, should
    #objects.sort(key=operator.itemgetter('title'))
    #objects.sort(key=operator.itemgetter('date'), reverse = True)
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
    write_markdown_combined_file(f_name, repo, group, objects)
    return None


def merge_resource_info(resource_info, obj, resource_type):
    '''update the resource info based on obj attributes, resource_type'''
    merge_fields = [
        "name", "description", "activity", "updated",
        "start", "end", "date", "updated", "alternative",
        "approx_start", "approx_end", "pi" ]
    for val in merge_fields:
        if val in obj:
            resource_info[val] = obj[val]
        elif val in resource_info:
            del resource_info[val]
    resource_info["resource_type"] = resource_type
    resource_info["resource_label"] = mk_label(resource_type)
    return resource_info

def build_repo_resource_objects(c_name, resource_type, repo_resources):
    '''generates a list of objects from resource type and repository resources'''
    objects = []
    for key in repo_resources[resource_type]:
        obj, err = dataset.read(f'{c_name}.ds', key)
        if err is not None and err != '':
            print(f'error access {key} in {c_name}.ds, {err}', file = sys.stderr)
        else:
            objects.append(enhance_object(obj))
    return objects

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
        for resource_type in repo_resources:
            resource_info = merge_resource_info(resource_info, obj, resource_type)
            objects = build_repo_resource_objects(c_name, resource_type, repo_resources)
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


def render_data_files(d_name, obj, group_id = None, people_id = None):
    '''render the resource JSON files for group_id'''
    # build out the resource type JSON file
    c_name = 'data'
    repo_id = f'Caltech{c_name.upper()}'
    if repo_id in obj:
        repo_resources = obj[repo_id]
        for resource_type in repo_resources:
            objects = []
            for key in repo_resources[resource_type]:
                obj, err = dataset.read(f'{c_name}.ds', key)
                if err is not None and err != '':
                    print(f'error access {key} in {c_name}.ds, {err}', file = sys.stderr)
                else:
                    objects.append(enhance_object(obj))
            if len(objects) > 0:
                # Write JSON resource file
                f_name = os.path.join(d_name, f'{resource_type}.json')
                write_json_file(f_name, objects)
                # Handle the recent sub folder
                write_json_file(f_name, objects[0:25])
                resource_info =  {
                        "repository": "CaltechDATA", 
                        "href":"https://data.caltech.edu",
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

def group_has_content(group):
    '''check to make sure it makes sense to render the group. There should
    be some type of records available to populate feeds'''
    if 'CaltechAUTHORS' in group:
        return True
    if 'CaltechTHESIS' in group:
        return True
    if 'CaltechDATA' in group:
        return True
    if ('description' in group) and (group['description'] != ''):
        return True
    return False


def render_a_group(group_list, group_id):
    '''render a specific group's content if valid'''
    if (group_id == '') and (' ' in group_id):
        print(f'error: "{group_id}" is not valid', file = sys.stderr)
        return
    if group_id not in group_list:
        print(f'error: "could not find {group_id}" in group list', file = sys.stderr)
        return
    obj = group_list[group_id]
    if group_has_content(obj):
        src = json.dumps(obj, indent=4)
        d_name = os.path.join('htdocs', 'groups', group_id)
        if not os.path.exists(d_name):
            os.makedirs(d_name, mode=0o777, exist_ok=True)
        # render the combined*.md files
        for repo in [ "authors", "thesis", "data" ]:
            #print(f'DEBUG rending combined files: {repo}', file = sys.stderr)
            err = render_combined_files(repo, d_name, group_id, obj)
            if err is not None:
                print(
                f'error: render_combined_files({repo}' +
                f', {d_name}, {group_id}) -> {err}', file = sys.stderr)


def render_groups(app_name, group_list, group_id = None):
    '''take our agents_csv and agent_pubs_csv filenames and aggregate them'''
    if group_id is not None:
        render_a_group(group_list, group_id)
    else:
        tot = len(group_list)
        widgets=[
            f'{app_name}'
            ' ', progressbar.Counter(), f'/{tot}',
            ' ', progressbar.Percentage(),
            ' ', progressbar.AdaptiveETA(),
        ]
        bar = progressbar.ProgressBar(max_value = tot, widgets=widgets)
        for i, grp_id in enumerate(group_list):
            render_a_group(group_list, grp_id)
            bar.update(i)
        bar.finish()

def main():
    '''main processing method'''
    app_name = os.path.basename(sys.argv[0])
    argc = len(sys.argv)
    if (argc < 2) or (argc > 3):
        print(f'{app_name} expected path to group_combined_list.json file', file = sys.stderr)
        sys.exit(1)
    group_list = get_group_combined_list(sys.argv[1])
    if group_list is None:
        print(f'could not populate group_list from {sys.argv[1]}')
        sys.exit(1)
    group_id = None
    if argc == 3:
        group_id = sys.argv[2]
    render_groups(app_name, group_list, group_id)

if __name__ == '__main__':
    main()
