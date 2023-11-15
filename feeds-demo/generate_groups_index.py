#!/usr/bin/env python3
'''Use the htdocs/groups/group_list.json to build htdocs/groups/index.json 
   and htdocs/groups/index.md'''

import os
import sys
import json
#import csv
#import operator
from subprocess import Popen, PIPE, TimeoutExpired

import yaml

def get_group_list(group_list_json):
    '''Get the group_list.json file and return a useful data structure.'''
    print(f'Reading {group_list_json}', file = sys.stderr)
    with open(group_list_json, encoding = 'utf-8') as _f:
        src = _f.read()
        if isinstance(src, bytes):
            src = src.decode('utf-8')
        group_list = json.loads(src)
        return group_list
    return None

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
            print(f'error: {errs}', file = sys.stderr)
            sys.exit(20)
    return None

def build_a_to_z_object(index, group_list):
    '''build an object form the decoded array'''
    objects = []
    a_to_z = []
    last_letter = ''
    for key in index:
        group = group_list.get(key, None)
        if group is None:
            print(f'WARNING failed to retrieve group {key} in build_a_to_z_object(index, group_list)')
            name = key.replace('-', ' ', -1)
            letter = name[0].upper()
        else:
            name = group.get('name', None)
            letter = name[0].upper()
        if letter != last_letter:
            a_to_z.append({'href': f'#{letter}', 'label': f' {letter} '})
            last_letter = letter
            objects.append({"id": key, "name": name, "letter": f' {letter} '})
        else:
            objects.append({"id": key, "name": name})
    return {
        "a_to_z": a_to_z,
        "content": objects,
    }

def render_a_to_z_list(f_name, index, group_list):
    '''generate an A to Z list for htdocs/groups/index.md from htdocs/groups/index.json'''
    pg_obj = build_a_to_z_object(index, group_list)
    pandoc_write_file(f_name, pg_obj, 'templates/groups-index.md', {
        "from_fmt": "markdown",
        "to_fmt": "markdown"
    })

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


def build_index_list(group_list):
    '''build an index of course based on if they have content or not'''
    index = []
    for k in group_list:
        if (k == '') or (' ' in k):
            print(f'group id {k} is not valid', file = sys.stderr)
        else:
            group = group_list[k]
            if group_has_content(group):
                index.append(k)
    return index

def main():
    '''main processing method'''
    app_name = os.path.basename(sys.argv[0])
    argc = len(sys.argv)
    if argc != 2:
        print(f'{app_name} expected path to group_list.json file', file = sys.stderr)
        sys.exit(1)
    group_list = get_group_list(sys.argv[1])
    index = build_index_list(group_list)
    f_name = os.path.join('htdocs', 'groups', 'index.json')
    # Write out htdocs/groups/index.json
    write_json_file(f_name, index)
    # Write out the A to Z list as htdocs/groups/index.md"
    f_name = os.path.join('htdocs', 'groups', 'index.md')
    render_a_to_z_list(f_name, index, group_list)

if __name__ == '__main__':
    main()
