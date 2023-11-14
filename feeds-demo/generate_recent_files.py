#!/usr/bin/env python3
'''Use the htdocs/recent/*.json to build out each of the publication type markdown files'''

import os
import sys
import json
import csv
import operator
from subprocess import Popen, PIPE, TimeoutExpired

# Caltech Library Packages
from py_dataset import dataset
# 3rd Party Packages
import progressbar
import yaml
from pybtex.database import BibliographyData, Entry
from feedgen.feed import FeedGenerator


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
    print(f'Writing {f_name}', file = sys.stderr)
    with open(f_name, 'w') as _f:
        _f.write('\n')
        for obj in objects:
            record_id = obj.get('id', None)
            title = obj.get('title', None)
            abstract = obj.get('abstract', None)
            if (record_id is not None) and (title is not None) and (abstract is not None):
                _f.write(object_to_bibtex(obj))


def remove_prefix(text, prefix):
    if text.startswith(prefix):
        return text[len(prefix):]
    return text


def write_rss_file(f_name, feed_title, objects):
    print(f'Writing {f_name}', file = sys.stderr)
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
        if (record_uri is not None) and (title is not None) and (abstract is not None):
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
            if len(description) > 0:
                entry.description('\n\n'.join(description))
    src = fg.rss_str(pretty=True)
    with open(f_name, 'w') as _f:
        if isinstance(src, bytes):
            src = src.decode('utf-8')
        _f.write(src)


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
    #print(f'DEBUG enhance object -> {obj}', file = sys.stderr)
    if 'date' in obj:
        obj['pub_year'] = obj['date'][0:4]
    if ('creators' in obj) and ('items' in obj['creators']):
        _l = format_authors(obj['creators']['items'])
        if _l is not None:
            obj['author_list'] = _l
    return obj

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
    resource = enhance_object(resource)
    #print(f'DEBUG enhance resource for Pandoc -> {resource}', file = sys.stderr)
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
    '''write a recent resource page by transform our list of objects'''
    p_objects = pandoc_build_resource(base_object, resource)
    err = pandoc_write_file(f_name, p_objects, 'templates/recent-resource.md')
    if err is not None:
        print(f'pandoc error: {err}', file = sys.stderr)

def write_markdown_combined_file(f_name, base_object, resource_list):
    '''coorodiante the write of a combined markdown file'''
    repository = base_object.get('repository', None)
    href = base_object.get('href', None)
    objects = base_object
    objects['content'] = []
    resource_type = None
    for item in resource_list:
        r_type = item.get('resource_type', None)
        if resource_type != r_type:
            resource_type = r_type
            objects['content'].append(pandoc_enhance_item(
                repository, href, resource_type, item))
        else:
            objects['content'].append(pandoc_enhance_item(
                None, None, None, item))

    err = pandoc_write_file(f_name, objects, 'templates/recent-combined.md',
        { 'from_fmt': 'markdown', 'to_fmt': 'markdown' })
    if err is not None:
        print(f'pandoc error: {err}', file = sys.stderr)

def render_authors():
    # render individual resource files
    f_name = os.path.join('htdocs', 'recent', 'object_types.json')
    resource_types = read_json_file(f_name)  
    obj = { 
        'repository': 'CaltechAUTHORS',
        'href': 'https://authors.library.caltech.edu'
    }
    for resource in resource_types:
        resource_type = resource.get('name', None)
        obj['resource_type'] = resource_type
        f_name = os.path.join('htdocs', 'recent', resource_type + '.json')
        resource_list = read_json_file(f_name)
        # Write out Markdown file
        f_name = os.path.join('htdocs', 'recent', resource_type + '.md')
        write_markdown_resource_file(f_name, obj, resource_list)
        # Write out BibTeX file
        f_name = os.path.join('htdocs', 'recent', resource_type + '.bib')
        write_bibtex_file(f_name, resource_list)
        # Write out RSS file
        f_name = os.path.join('htdocs', 'recent', resource_type + '.rss')
        label = mk_label(resource_type) + ' feed'
        write_rss_file(f_name, label, resource_list)
    # render combined resource file
    del obj['resource_type']
    f_name = os.path.join('htdocs', 'recent', 'combined.json')
    resource_list = read_json_file(f_name)
    # Write out Markdown file
    f_name = os.path.join('htdocs', 'recent', 'combined.md')
    write_markdown_combined_file(f_name, obj, resource_list)
    # Write out BibTeX file
    f_name = os.path.join('htdocs', 'recent', 'combined.bib')
    write_bibtex_file(f_name, resource_list)
    # Write out RSS file
    f_name = os.path.join('htdocs', 'recent', 'combined.rss')
    label = 'Combined feed'
    write_rss_file(f_name, label, resource_list)


def render_data():
    # render individual resource files
    f_name = os.path.join('htdocs', 'recent', 'data_object_types.json')
    resource_types = read_json_file(f_name)  
    obj = { 
        'repository': 'CaltechDATA',
        'href': 'https://data.caltech.edu'
    }
    for resource in resource_types:
        resource_type = resource.get('name', None)
        obj['resource_type'] = resource_type
        f_name = os.path.join('htdocs', 'recent', resource_type + '.json')
        resource_list = read_json_file(f_name)
        # Write out Markdown file
        f_name = os.path.join('htdocs', 'recent', resource_type + '.md')
        write_markdown_resource_file(f_name, obj, resource_list)
        # Write out BibTeX file
        f_name = os.path.join('htdocs', 'recent', resource_type + '.bib')
        write_bibtex_file(f_name, resource_list)
        # Write out RSS file
        f_name = os.path.join('htdocs', 'recent', resource_type + '.rss')
        label = mk_label(resource_type) + ' feed'
        write_rss_file(f_name, label, resource_list)
    # render combined resource file
    del obj['resource_type']
    f_name = os.path.join('htdocs', 'recent', 'combined_data.json')
    resource_list = read_json_file(f_name)
    # Write out Markdown file
    f_name = os.path.join('htdocs', 'recent', 'combined_data.md')
    write_markdown_combined_file(f_name, obj, resource_list)
    # Write out BibTeX file
    f_name = os.path.join('htdocs', 'recent', 'combined_data.bib')
    write_bibtex_file(f_name, resource_list)
    # Write out RSS file
    f_name = os.path.join('htdocs', 'recent', 'combined_data.rss')
    label = 'Combined Data feed'
    write_rss_file(f_name, label, resource_list)


def render_recent():
    # Render authors' recent
    render_authors()
    # Recent data's recent
    render_data()
    
def main():
    '''main processing method'''
    app_name = os.path.basename(sys.argv[0])
    argc = len(sys.argv)
    if argc != 1:
        print(f'{app_name} expected only the app name to render htdocs/recent/*', file = sys.stderr)
        sys.exit(1)
    render_recent()

if __name__ == '__main__':
    main()
