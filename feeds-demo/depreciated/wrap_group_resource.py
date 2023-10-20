#!/usr/bin/env python3
'''wrap_group_resource.py generates YAML front matter to render a Markdown
document for a group's specific resource (e.g. article, PhD)'''

import sys
import os
import json

import yaml

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
                if len(authors) > 3:
                    break
        if len(authors) == 1:
            return authors[0]
        if len(authors) == 2:
            return ' and '.join(authors[0:2])
        if len(authors) > 2:
            return '; '.join(authors[0:2]) + '; et el.'
    return None

def enhance_item(repository = None, href = None, resource = {}):
    '''given a resource, enhance it to make it friendly to tempalte in Pandoc'''
    citation = {}
    if repository is not None:
        citation['repository'] = repository
    if href is not None:
        citation['href'] = href
    if 'date' in resource:
        citation['pub_year'] = resource['date'][0:4]
    if ('creators' in resource) and ('items' in resource['creators']):
        _l = format_authors(resource['creators']['items'])
        if _l is not None:
            citation['author_list'] = _l
    for k in resource:
        citation[k] = resource[k]
    return citation

def build_json_object(src, obj_expr = None):
    '''build an object form the decoded array'''
    objects = {}
    if obj_expr is not None:
        if isinstance(obj_expr, bytes):
            obj_expr = obj_expr.decode('utf-8')
        objects = json.loads(obj_expr)
    if 'repository' not in objects:
        print(f'missing repository value in object expression', file = sys.stderr)
        return None
    repository = objects['repository']
    if 'href' not in objects:
        print(f'missing href value in object expression', file = sys.stderr)
        return None
    href = objects['href']

    if isinstance(src, bytes):
        src = src.decode('utf-8')
    resource = json.loads(src)

    objects['content'] = []
    for i, obj in enumerate(resource):
        if i == 0:
            objects['content'].append(enhance_item(repository, href, resource[i]))
        else:
            objects['content'].append(enhance_item(None, None, resource[i]))
    return objects


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
