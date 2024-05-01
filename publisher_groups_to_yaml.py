#!/usr/bin/env python3

import csv
import yaml

def load(f_name):
    '''read in a vocabulary expressed in YAML and return a python object'''
    data = None
    with open(f_name, encoding = 'utf-8') as f:
        src = f.read()
        data = yaml.load(src, Loader= yaml.Loader)
    if data is None:
        print(f'failed to load {f_name}', file = sys.stderr)
        return False
    return data

with open("CaltechAUTHORS_publisher_groups.csv", newline = "") as csvfile:
    m = load("options.yaml")
    issn_publishers = {}
    issn_journals = {}
    #doi_prefix_to_publisher = {}
    header = [ "Preferred Publisher Name","journal","issn","e-issn","DOI pattern","Note" 
 ]
    reader = csv.DictReader(csvfile, header)
    for i, row in enumerate(reader):
        if i > 0:
            issn = row.get('issn', '').strip()
            e_issn = row.get('e-issn', '').strip()
            publisher = row.get('Preferred Publisher Name', '').strip()
            journal = row.get('journal', '').strip()
            if issn != '':
                if publisher != '':
                    issn_publishers[issn] = publisher
                if journal != '':
                    issn_journals[issn] = journal
            if e_issn != '':
                if publisher != '':
                    issn_publishers[e_issn] = publisher
                if journal != '':
                    issn_journals[e_issn] = journal
    m['issn_journals'] = issn_publishers
    m['issn_publishers'] = issn_journals
    print(yaml.dump(m))

