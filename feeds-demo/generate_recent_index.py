#!/usr/bin/env python3
'''Use the htdocs/recent/*.json files to htdocs/recent/index.md files.'''

import os
import sys
import json
#import csv
import operator
from subprocess import Popen, PIPE, TimeoutExpired

import yaml

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

def read_json_file(f_name):
    with open(f_name) as _f:
        src = _f.read()
        if isinstance(src, bytes):
            src = src.decode('utf-8')
        return json.loads(src)
    return None

def generate_recent_list(repositories_info):
    obj = {}
    obj['content'] = []
    for repo_name in [ 'CaltechAUTHORS', 'CaltechDATA' ]:
        info = repositories_info.get(repo_name, None)
        # load the combined and object types file for authors
        combined_name = info.get('combined', None)
        if combined_name is not None:
            obj['content'].append({'repository': repo_name, 'label': 'Combined', 'name': combined_name})
        b_name = info.get('object_types', None)
        if b_name is None:
            print('failed to find object types JSON filename for {repo_name}', file = sys.stderr)
        else:
            f_name = os.path.join('htdocs', 'recent', b_name + '.json')
            object_types = read_json_file(f_name)
            object_types.sort(key = operator.itemgetter('label'))
            for obj_type in object_types:
                obj['content'].append(obj_type)
    return obj

def main():
    '''main processing method'''
    app_name = os.path.basename(sys.argv[0])
    argc = len(sys.argv)
    if argc != 1:
        print(f'{app_name} process combined and object type JSON files in htdocs/recent', file = sys.stderr)
        sys.exit(1)
    # generate the an object structure for sending to Pandoc
    obj = generate_recent_list({
        "CaltechAUTHORS": {
            "combined": "combined",
            "object_types": "object_types"
        },
        "CaltechDATA": {
            "combined": "combined_Data",
            "object_types": "data_object_types"
        }
    })
    # Write out the htdocs/recent/index.md file
    f_name = os.path.join('htdocs', 'recent', 'index.md')
    pandoc_write_file(f_name, obj, 'templates/recent-index.md', {
        "from_fmt": "markdown",
        "to_fmt": "markdown"
    })

if __name__ == '__main__':
    main()
