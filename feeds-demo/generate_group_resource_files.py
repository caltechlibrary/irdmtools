#!/usr/bin/env python3
'''Use the htdocs/groups/group_list.json to build out each of the publication JSON lists'''

import os
import sys
import json
import csv
import operator
from subprocess import Popen, PIPE, TimeoutExpired

# Caltech Library package
from py_dataset import dataset
# 3rd Party packages
import progressbar
import yaml
from pybtex.database import BibliographyData, Entry
from feedgen.feed import FeedGenerator


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

def format_editor(editors):
    '''format the authors to be friendly to Pandoc template'''
    if len(editors) > 0:
        editors = []
        for i, editor in enumerate(editors):
            if 'name' in editors:
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
        if len(editors) == 1:
            return editors[0]
        if len(editors) == 2:
            return ' and '.join(editors[0:2])
        if len(editors) > 2:
            return '; '.join(editors[0:2]) + '; et el.'
    return None

def enhance_object(obj):
    '''given an eprint like record, enhance the record to make it Pandoc template friendly'''
    if 'date' in obj:
        obj['pub_year'] = obj['date'][0:4]
    if ('creators' in obj) and ('items' in obj['creators']):
        _l = format_authors(obj['creators']['items'])
        if _l is not None:
            obj['author_list'] = _l
    if ('editor' in obj) and ('items' in obj['editor']):
        _l = format_editor(obj['editor']['items'])
        if _l is not None:
            obj['editor_list'] = _l
    return obj

def write_json_file(f_name, objects):
    '''render the JSON file from the objects.'''
    src = json.dumps(objects, indent = 4)
    #print(f'Writing {f_name}', file = sys.stderr)
    with open(f_name, 'w', encoding = 'utf-8') as _w:
        _w.write(src)

def object_person_to_name_bib_list(field_name, obj):
    people = []
    agents = obj.get(field_name, None)
    if agents is not None:
        items = agents.get('items', [])
        for item in items:
            name = item.get('name', None)
            if name is not None:
                family = name.get('family', None)
                given = name.get('given', None)
                if (family is not None) and (given is not None):
                    people.append(f'{family}, {given}')
    return people

def object_to_bibtex(obj):
    resource_type = obj.get('resource_type', 'other')
    title = obj.get('title', None)
    journal = obj.get('journal', None)
    pub_year = obj.get('pub_year', None)
    official_url = obj.get('official_url', None)
    record_uri = obj.get('id', None)
    isbn = obj.get('isbn', None)
    issn = obj.get('issn', None)
    doi = obj.get('doi', None)
    vol = obj.get('volume', None)
    vol_no = obj.get('number', None)
    pmcid = obj.get('pmcid', None)
    pages = obj.get('pagerange', None)
    authors = object_person_to_name_bib_list('creators', obj)
    _l = []
    if len(authors) > 0:
        _l.append(('author', ' and '.join(authors)))
    if title is not None:
        _l.append(('title', title))
    if journal is not None:
        _l.append(('journal', journal))
    if pub_year is not None:
        _l.append(('year', pub_year))
    if official_url is not None:
        _l.append(('url', official_url))
    if record_uri is not None:
        _l.append(('id', record_uri))
    if isbn is not None:
        _l.append(('isbn', isbn))
    if issn is not None:
        _l.append(('issn', issn))
    if doi is not None:
        _l.append(('doi', doi))
    if vol is not None:
        _l.append(('volume', vol))
    if vol_no is not None:
        _l.append(('number', vol_no))
    if pmcid is not None:
        _l.append(('pmcid', pmcid))
    if pages is not None:
        _l.append(('pages', pages))
    if resource_type == 'article':
        entry_type = 'article'
    elif resource_type.startswith('book'):
        entry_type = 'book'
    else:
        entry_type = 'other'
    entry = Entry(entry_type, _l)
    return BibliographyData({
        record_uri: entry,
    }).to_string('bibtex') + "\n\n"


def write_bibtex_file(f_name, objects):
    '''takes a list record objects and write BibTeX using pybtex to a file'''
    with open(f_name, 'w') as _f:
        _f.write('\n')
        for obj in objects:
            _f.write(object_to_bibtex(obj))

def remove_prefix(text, prefix):
    if text.startswith(prefix):
        return text[len(prefix):]
    return text

def write_rss_file(f_name, feed_title, objects):
    feed_url = 'https://feeds.library.caltech.edu/' + remove_prefix(f_name, 'htdocs/')
    fg = FeedGenerator()
    fg.id(feed_url)
    fg.title(feed_title)
    fg.description('A Caltech Library Repository Feed')
    fg.link(href = feed_url, rel = 'self')
    fg.language('en')
    fg.author({'name': 'Caltech Library', 'email': 'library@caltech.edu'})
    for obj in objects:
        record_uri = obj.get('id', None)
        title = obj.get('title', None)
        pub_year = obj.get('pub_year', None)
        official_url = obj.get('official_url', None)
        authors = object_person_to_name_bib_list('creators', obj)
        abstract = obj.get('abstract', None)
        doi = obj.get('doi', None)
        pmcid = obj.get('pmcid', None)
        if record_uri is not None:
            description = []
            if len(authors) > 0:
                description.append(f'''Authors: {'; '.join(authors)}''')
            if pub_year is not None:
                description.append(f'Year: {pub_year}')
            if doi is not None:
                description.append(f'DOI: {doi}')
            if pmcid is not None:
                description.append(f'PMCID: {pmcid}')
            if abstract is not None:
                description.append(f'${abstract}')
            entry = fg.add_entry()
            entry.id(record_uri)
            entry.guid(record_uri)
            if official_url is not None:
                entry.link( href = official_url)
            else:
                entry.link(href = record_uri)
            if title is not None:
                entry.title(title)
            entry.description('\n\n'.join(description))
    src = fg.rss_str(pretty=True)
    with open(f_name, 'w') as _f:
        if isinstance(src, bytes):
            src = src.decode('utf-8')
        _f.write(src)


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

def flatten_index_resource_types(repository, href, combined, resource_object):
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
        obj, err = dataset.read(c_name, key)
        if err is not None and err != '':
            print(f'error access {key} in {c_name}, {err}', file = sys.stderr)
        else:
            objects.append(enhance_object(obj))
    return objects

def render_authors_files(d_name, obj, group_id = None, group_name = None, people_id = None):
    '''render the resource JSON files for group_id'''
    # build out the resource type JSON file
    c_name = 'authors.ds'
    repo_id = 'CaltechAUTHORS'
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
                    resource_info["group_label"] = group_name
                if people_id is not None:
                    resource_info["people_id"] = people_id
                    resource_info["people_label"] = mk_label(people_id)
                # Write out Markdown files via Pandoc
                f_name = os.path.join(d_name, f'{resource_type}.md')
                write_markdown_resource_file(f_name, resource_info, objects)
                # Write out BibTeX file
                f_name = os.path.join(d_name, f'{resource_type}.bib')
                write_bibtex_file(f_name, objects)
                # Write out RSS file
                f_name = os.path.join(d_name, f'{resource_type}.rss')
                label = mk_label(resource_type)
                write_rss_file(f_name, f'{label} feed', objects)

def render_thesis_files(d_name, obj, group_id = None, group_name = None, people_id = None):
    '''render the resource JSON files for group_id'''
    # build out the resource type JSON file
    c_name = 'thesis.ds'
    repo_id = 'CaltechTHESIS'
    if repo_id in obj:
        repo_resources = obj[repo_id]
        for resource_type in repo_resources:
            f_name = os.path.join(d_name, f'{resource_type}.json')
            objects = []
            for key in repo_resources[resource_type]:
                obj, err = dataset.read(c_name, key)
                if err is not None and err != '':
                    print(f'error access {key} in {c_name}, {err}', file = sys.stderr)
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
                    resource_info["group_label"] = group_name
                if people_id is not None:
                    resource_info["people_id"] = people_id
                    resource_info["people_label"] = mk_label(people_id)
                # Write out Markdown file
                f_name = os.path.join(d_name, f'{resource_type}.md')
                write_markdown_resource_file(f_name, resource_info, objects)
                # Write out BibTeX file
                f_name = os.path.join(d_name, f'{resource_type}.bib')
                write_bibtex_file(f_name, objects)
                # Write out RSS file
                f_name = os.path.join(d_name, f'{resource_type}.rss')
                label = mk_label(resource_type)
                write_rss_file(f_name, f'{label} feed', objects)


def render_data_files(d_name, obj, group_id = None, group_name = None, people_id = None):
    '''render the resource JSON files for group_id'''
    # build out the resource type JSON file
    c_name = 'data.ds'
    repo_id = f'CaltechDATA'
    if repo_id in obj:
        repo_resources = obj[repo_id]
        for resource_type in repo_resources:
            objects = []
            for key in repo_resources[resource_type]:
                obj, err = dataset.read(c_name, key)
                if err is not None and err != '':
                    print(f'error access {key} in {c_name}, {err}', file = sys.stderr)
                else:
                    objects.append(enhance_object(obj))
            if len(objects) > 0:
                # Write JSON resource file
                f_name = os.path.join(d_name, f'{resource_type}.json')
                write_json_file(f_name, objects)
                # Handle the recent sub folder
                resource_info =  {
                    "repository": "CaltechDATA", 
                    "href":"https://data.caltech.edu",
                    "resource_type": resource_type,
                    "resource_label": mk_label(resource_type)
                }
                if group_id is not None:
                    resource_info["group_id"] = group_id
                    resource_info["group_label"] = group_name
                if people_id is not None:
                    resource_info["people_id"] = people_id
                    resource_info["people_label"] = mk_label(people_id)
                # Write out Markdown file
                f_name = os.path.join(d_name, f'{resource_type}.md')
                write_markdown_resource_file(f_name, resource_info, objects)
                # Write out BibTeX file
                f_name = os.path.join(d_name, f'{resource_type}.bib')
                write_bibtex_file(f_name, objects)
                # Write out RSS file
                f_name = os.path.join(d_name, f'{resource_type}.rss')
                label = mk_label(resource_type)
                write_rss_file(f_name, f'{label} feed', objects)

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
    group_name = obj.get('name', None)
    if group_has_content(obj):
        src = json.dumps(obj, indent=4)
        d_name = os.path.join('htdocs', 'groups', group_id)
        if not os.path.exists(d_name):
            os.makedirs(d_name, mode=0o777, exist_ok=True)

        # Write out the legacy group.json file
        f_name = os.path.join(d_name, 'group.json')
        write_json_file(f_name, obj)

        # Now render the repo resource files.
        render_authors_files(d_name, obj, group_id = group_id, group_name = group_name)
        render_thesis_files(d_name, obj, group_id = group_id, group_name = group_name)
        render_data_files(d_name, obj, group_id = group_id, group_name = group_name)
        # FIXME: render the group index.json file
        f_name = os.path.join(d_name, 'index.json')
        write_json_file(f_name, obj)
        # FIXME: render the group index.md file
        f_name = os.path.join(d_name, 'index.md')
        write_markdown_index_file(f_name, obj)


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
            bar.update(i)
            render_a_group(group_list, grp_id)
        bar.finish()

def main():
    '''main processing method'''
    app_name = os.path.basename(sys.argv[0])
    argc = len(sys.argv)
    if (argc < 2) or (argc > 3):
        print(f'{app_name} expected path to group_resources.json file', file = sys.stderr)
        sys.exit(1)
    group_list = get_group_list(sys.argv[1])
    if group_list is None:
        print(f'could not populate group_list from {sys.argv[1]}')
        sys.exit(1)
    group_id = None
    if argc == 3:
        group_id = sys.argv[2]
    render_groups(app_name, group_list, group_id)

if __name__ == '__main__':
    main()
