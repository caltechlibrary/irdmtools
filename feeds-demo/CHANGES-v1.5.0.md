
# v1.5

This is a major rewrite of our feeds system predicated by the migration of CaltechAUTHORS
from EPrints 3.3 to Invenio-RDM.

It is based on dataset v2.1 leveraging Postgres 15 as well as recent releases of datatools
and the new irdmtools that replaces the old eprinttools cli collection.

General process orchestration is done with Bash scripts delegating low level data flow
to the irdmtools collection and data shaping in the website rendering phase to python and
Pandoc.

## Explicit changes

Our feeds system has been updated as a result of the CaltechAUTHORS repository migrating from EPrints to Invenio RDM. As a result some things on feeds.library.caltech.edu had to be changed. This is a list of intended changes.

- In the recent directory there is no recent 25 for CaltechTHESIS content, just doesn't make sense we graduate people in "class of" groupings, which 25 people should be listed? That has gone away as a result
- With the migration from one system to another many Caltech GROUPs have revised names and identifiers, e.g. `/groups/TCCON` became `/groups/Total-Carbon-Column-Observing-Network`
- Record ids changed between systems for CaltechAUTHORS so if you look at those links they look different
- Minor HTML markup changes to make the feeds site more accessible (e.g. A to Z list in groups now uses a "menu" element instead of a paragraph with pipe delimiters)
- [Pagefind](https://pagefind.app) provides searching of feed's HTML pages
- The software to generate the feeds website has been completely rewritten so invariably there are changes that I have not mentioned.
- Every thing in the "htdocs" tree is generated, this means that directory can be safely removed and recreated as needed
- Static files that need to be included in the "htdocs" tree can be found in the "static" directory (they are just copied into when needed)
- Pandoc is used exclusively to render Markdown, HTML, HTML Includes, BibTeX and RSS files from JSON files rendered from the repositories and collections
- Pandoc templates can be found in the "templates" directory. Their file extensions correspond to the format they are intended to render
    - The generated Markdown is used to render both HTML, HTML Includes
    - HTML Include is generated directly by Pandoc without a template
    - BibTeX and RSS required their own templates
- Processing order
    - Static content is copied into place
    - JSON files are created
    - Markdown files are created along side some JSON
    - BibTeX and RSS are generated where approapriate
    - All Markdown files are rendered as HTML and HTML Includes
    - Pagefind indexes the site and generates our static search support
- Dataset collections are no longer be published in feeds
- The `recent` directories under people and groups are no longer being generated in favor of using JavaScript or our Widget to present shorter lists
- Groups inclusion is based on being identified in the groups list of objects and having something in one or more records in CaltechAUTHORS, CaltechTHESIS or CaltechDATA
- People inclusion is based on being identified as a Caltech person and being listed as an author in a CaltechAUTHORS's metadata record


## System requirements

Feeds v1.5 requires the following software to be built

- irdmtools >= 0.0.57 (use the latest release)
- dataset >= 2.1.6 (use the latest release)
- datatools >= 1.2.5 (use the latest release)
- Bash >= 3 (or equivalent POSIX shell)
- Pandoc >= 3
- Postgres >= 14
- Python >= 3.10
- PageFind >= v1.0.3

Bash scripts orchestrate most of the processing. Python is used to transformed the legacy data shapes into needed forms and
to generate Markdown content via Pandoc.





