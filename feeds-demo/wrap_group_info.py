#!/usr/bin/env python3
'''wrap_array.py takes a JSON document which contains an array and wraps it
into a JSON object with an single attribute, "content", pointing at the array. 
It then converts this to YAML front matter suitable to pipe to Pandoc'''

import sys
import os
import json

import yaml

def mk_label(val):
    if '_' in val:
        val = val.replace('_', ' ', -1)
    return val.title()

def flatten_resource_types(repository, href, combined, resource_object):
    '''flatten takes a dictionary pointing at an array and flattens it into an
    array suitable for processing with Pandoc templates'''
    objects = []
    for i, resource_type in enumerate(resource_object):
        if i == 0:
            objects.append({'repository': repository, 'href': href, 'combined': combined, 'label': mk_label(resource_type), 'resource_type': resource_type})
        else:
            objects.append({'label': mk_label(resource_type), 'resource_type': resource_type})
    return objects

def build_json_object(src):
    '''build an object form the decoded array'''
    if src == '':
        print('object type of group JSON is empty', file = sys.stderr)
        return None
    group = json.loads(src)
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
            for resource in flatten_resource_types(k, u_map[k], c_map[k], group[k]):
                resources.append(resource)
        else:
            obj[k] = group[k]
    obj['content'] = resources
    return obj


def main():
    '''main processing procedure'''
    app_name = os.path.basename(sys.argv[0])
    if len(sys.argv) == 1:
        print(f'usage: {app_name} GROUPS_JSON_FILE')
        sys.exit(1)
    if len(sys.argv) != 2:
        print('expected GROUPS_JSON_FILE')
        sys.exit(1)
    f_name = sys.argv[1]
    src = ''
    with open(f_name, 'r', encoding = 'utf-8') as _f:
        src = _f.read()
        if isinstance(src, bytes):
            src= src.decode('utf-8')
    if src == '':
        print(f'failed to read {f_name}', file = sys.stderr)
        sys.exit(1)
    obj = build_json_object(src)
    if obj is not None:
        print('---')
        print(yaml.dump(obj))
        print('---')
    else:
        print('problem building object', file = sys.stderr)
        sys.exit(1)


#
# Main
#
if __name__ == '__main__':
    main()
