#!/usr/bin/env python3
'''wrap_a_to_z_array.py takes a JSON document which contains an array and wraps it
into a JSON object to produce a JSON object that is easy to implement an A to Z
list for groups.'''

import sys
import os
import json

import yaml

def build_json_object(src):
    '''build an object form the decoded array'''
    keys = json.loads(src)
    keys.sort()
    objects = []
    a_to_z = []
    last_letter = ''
    for key in keys:
        name = key.replace('-', ' ', -1)
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


def main():
    '''main processing procedure'''
    app_name = os.path.basename(sys.argv[0])
    if len(sys.argv) == 1:
        print(f'usage: {app_name} JSON_FILE')
        sys.exit(1)
    f_name = sys.argv[1]
    if len(sys.argv) != 2:
        print('expected a JSON_FILE containing an array of group ids')
        sys.exit(1)
    with open(f_name, 'r', encoding = 'utf-8') as _f:
        src = _f.read()
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
