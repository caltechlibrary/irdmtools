%people2vocabulary(1) people2vocabulary user manual
% R. S. Doiel
% May 17, 2023

# NAME

people2vocabulary

# SYSNOPSIS

people2vocabulary < INPUT_JSON_FILE > OUTPUT_VOC_YAML_FILE

# DESCRIPTION

people2vocabulary converts a JSON array of people objects to a YAML
file suitable for import into Invenio-RDM.

# EXAMPLES

~~~shell
    people2vocabulary < htdocs/people/people.json \
	     >htdocs/people/people-vocabulary.yaml
~~~


