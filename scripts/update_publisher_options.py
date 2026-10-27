#!/usr/bin/env python3

import sys
import os
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

def merge_publisher_data(yaml_name, csv_name):
    '''read in the yaml and CSV File and merge the CSV content into the YAML
    data structure. Dump to standard out'''
    with open(csv_name, newline = "") as csvfile:
        m = load(yaml_name)
        issn_publishers = {}
        issn_journals = {}
        #doi_prefix_to_publisher = {}
        header = [ "Preferred Publisher Name","journal","issn","e-issn","DOI pattern","Note" ]
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
    
#
# Main
#
if __name__ == "__main__":
    app_name = os.path.basename(sys.argv[0])
    if len(sys.argv) != 3:
        print(f'''USAGE: {app_name} YAML_NAME CSV_FILE >OUTPUT_YAML
 
 Updates the YAML_FILE based on the contents of the CSV_FILE. The
 CSV file iexpected to have th following columns.

     Preferred Publisher Name, journal, issn, e-issn, DOI pattern, Note

 If the columns are labeled differently you will not get the results you
 expect. The columns can be in a different order.

''')
        sys.exit(1)
    merge_publisher_data(sys.argv[1], sys.argv[2])
