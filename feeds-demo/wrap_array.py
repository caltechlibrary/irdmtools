#!/usr/bin/env python3
'''wrap_array.py takes a JSON document which contains an array and wraps it
into a JSON object with an single attribute, "content", pointing at the array. 
It then converts this to YAML front matter suitable to pipe to Pandoc'''

import sys
import os
import json

import yaml

def build_json_object(src, obj_expr = None):
    '''build an object form the decoded array'''
    obj = {}
    if obj_expr is not None:
        if isinstance(obj_expr, bytes):
            obj_expr = obj_expr.decode('utf-8')
        try:
            obj = json.loads(obj_expr)
        except Exception as err:
            print(f'obj_expr decode error: {err}', file = sys.stderr)
            obj = {}
    if isinstance(src, bytes):
        src = src.decode('utf-8')
    obj['content'] = json.loads(src)
    return obj


def main():
    '''main processing procedure'''
    app_name = os.path.basename(sys.argv[0])
    if len(sys.argv) == 1:
        print(f'usage: {app_name} JSON_FILE [OBJ_EXPR]')
        sys.exit(1)
    if len(sys.argv) > 3:
        print('expected JSON_FILE and optional object expression')
        sys.exit(1)
    f_name = sys.argv[1]
    obj_expr = ''
    if len(sys.argv) == 3:
        obj_expr = sys.argv[2]
    with open(f_name, 'r', encoding = 'utf-8') as _f:
        src = _f.read()
        obj = build_json_object(src, obj_expr)
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
