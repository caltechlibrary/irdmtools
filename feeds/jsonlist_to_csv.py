#!/usr/bin/env python3

import os
import sys
import csv
import json

def usage(app_name):
    print(f'{app_name} JSON_ARRAY_FILENAME')

def main():
    app_name = os.path.basename(sys.argv[0])
    args = sys.argv[1:]
    if len(args) < 1:
        usage()
        sys.exit(1)
    # Get list of columns from objects
    with open(args[0]) as f:
        src = f.read()
        if isinstance(src, bytes):
            src = src.decode('utf-8')
        l = json.loads(src)
        columns = []
        for obj in l:
            for k in obj:
                if not k in columns:
                    columns.append(k)
        # Now we can writeout our list as a CSV file.
        writer = csv.DictWriter(sys.stdout, fieldnames = columns)
        writer.writeheader()
        for obj in l:
            writer.writerow(obj)
        
if __name__ == '__main__':
    main()
