#!/usr/bin/env python3
'''Use the htdocs/people/index.md from people_list.json'''

import os
import sys
import json
#import csv
import operator
from subprocess import Popen, PIPE, TimeoutExpired

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
        print(f'error: {out}', file = sys.stderr)
        sys.exit(20)
    return None

def build_a_to_z_object(people_list):
    '''build an object form the decoded array'''
    objects = []
    a_to_z = []
    last_letter = ''
    for i, person in enumerate(people_list):
        key = person.get('cl_people_id', None)
        if key is None:
            print(f'missing people id in people_list[{i}] -> {person}', file = sys.stderr)
            continue
        name = person.get('sort_name', None)
        if name is None:
            print(f'missing sort_name in people_list[{i}] -> {person}', file = sys.stderr)
            continue
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

def render_a_to_z_list(f_name, index):
    '''generate an A to Z list for htdocs/people/index.md from htdocs/people/people_list.json'''
    pg_obj = build_a_to_z_object(index)
    pandoc_write_file(f_name, pg_obj, 'templates/people-index.md', {
        "from_fmt": "markdown",
        "to_fmt": "markdown"
    })

def people_has_content(people):
    '''check to make sure it makes sense to render the people. There should
    be some type of records available to populate feeds'''
    authors_count = int(people.get('authors_count', 0))
    data_count = int(people.get('data_count', 0))

    if (authors_count > 0) or (data_count > 0):
        return True
    return False


def main():
    '''main processing method'''
    app_name = os.path.basename(sys.argv[0])
    argc = len(sys.argv)
    if argc != 2:
        print(f'{app_name} expected path to people.json file', file = sys.stderr)
        sys.exit(1)
    # retrieve the people_list
    people_list = get_people_list(sys.argv[1])
    # Write out the A to Z list as htdocs/people/index.md"
    f_name = os.path.join('htdocs', 'people', 'index.md')
    render_a_to_z_list(f_name, people_list)

if __name__ == '__main__':
    main()
