#!/usr/bin/env python3
'''jsonlist_to_csv.py converts a JSON list generated with dsquery to a CSV formatted output'''

import os
import sys
import csv
import json

def usage(app_name):
    '''describes out to use this program'''
    print(f'{app_name} JSON_ARRAY_FILENAME')

def main():
    '''this is the main processing function that does the work'''
    app_name = os.path.basename(sys.argv[0])
    args = sys.argv[1:]
    if len(args) < 1:
        usage(app_name)
        sys.exit(1)
    # Get list of columns from objects
    with open(args[0], encoding = 'utf-8') as _f:
        src = _f.read()
        if isinstance(src, bytes):
            src = src.decode('utf-8')
        _l = json.loads(src)
        columns = []
        for obj in _l:
            for name in obj:
                if not name in columns:
                    columns.append(name)
        # Now we can writeout our list as a CSV file.
        writer = csv.DictWriter(sys.stdout, fieldnames = columns)
        writer.writeheader()
        for obj in _l:
            writer.writerow(obj)

if __name__ == '__main__':
    main()
