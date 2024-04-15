%people2vocabulary(1) irdmtools user manual | version 0.0.81 05d8f535
% R. S. Doiel
% 2024-04-15

# NAME

people2vocabulary

# SYNOPSIS

people2vocabulary [OPTIONS] < INPUT_JSON_FILE > OUTPUT_VOC_YAML_FILE

# DESCRIPTION

people2vocabulary converts a JSON array of people objects to a YAML
file suitable for import into Invenio-RDM.

# OPTIONS

-help
: display help

-license
: display license

-version
: display version

-i
: Read input from file

-o
: Write output to file

-csv
: (default: true) Input is in csv format

-clrules
: (default: true) use Caltech Library rules

# EXAMPLES

~~~shell
    people2vocabulary < htdocs/people/people.json \
	     >people-vocabulary.yaml

	people2vocabulary -csv < htdocs/people/people.csv \
	     >people-vocabulary.yaml
~~~


