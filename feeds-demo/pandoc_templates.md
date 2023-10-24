
# Pandoc Templates

The Pandoc templates are found in the templates folder. The purpose is listed below.

Markdown is used to the Markdown pages which in turn will be used to
render HTML and HTML include files. Markdown templates need to be rendered before 
attempting to render HTML and HTML include files.


## Markdown templates

templates/citation.md
: This a partial rendering markdown citation an "enhanced" object from an EPrint JSON structure.
It is a partial template used by other templates for individual repository record citation.

templates/groups-group-index.md
: This template is used to generate `htdocs/groups/<GROUP_ID>/index.md`

templates/groups-group-resource.md
: This  template is used to generate `htdocs/groups/<GROUP_ID>/<RESOURCE_TYPE>.md`

templates/groups-index.md
: This template is used to genrate `htdocs/groups/index.md`, the A to Z list of groups

templates/recent-index.md
: This template is used to generate `htdocs/recent/index.md`

templates/recent-resource.md
: This template is used to generate `htdocs/recent/<RESOURCE_TYPE>.md`

## HTML templates

templates/page.html

