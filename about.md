---
title: irdmtools
abstract: |-
  Tools for working with institutional repositories and data management systems. Current implementation targets Invenio-RDM.
authors:
  - family_name: Doiel
    given_name: R. S.
    id: https://orcid.org/0000-0003-0900-6903
  - family_name: Morrell
    given_name: Tom
    id: https://orcid.org/0000-0001-9266-5146

contributor:
  - family_name: Johnson
    given_name: Kathy
    id: https://orcid.org/0000-0002-0026-2516
  - family_name: Porter
    given_name: George
    id: https://orcid.org/0000-0002-6539-638X
  - family_name: Kiser
    given_name: Cynthia
    id: https://orcid.org/0000-0002-0102-996X


repository_code: https://github.com/caltechlibrary/irdmtools
version: 0.0.98
license_url: https://caltechlibrary.github.io/irdmtools/LICENSE
operating_system:
  - Linux
  - macOS
  - Windows

programming_language:
  - Go
  - Python

keywords:
  - institutional repository
  - data management
  - Invenio
  - Invenio-RDM

date_released: 2026-08-13
---

About this software
===================

## irdmtools 0.0.98

- Fix issue #87, set title type to "other" on additional_titles crosswalked from CrossRef records with multiple titles
- Fix typo mapping teachingresource-lecturenotes resource type back to EPrints ("teching_resource" -> "teaching_resource")
- Updated for Go v1.26.1
- Changed get_modified_id to pick a default end date that is 24 hours in advanced before today to capture today's records
- Proof of concept relatedItem mappings for DataCite records per issue #77
- Fix issue #85, upgrade to latest crossrefapi package, then added support for ROR as funder identifier.
- Fix issue #86 for mapping journals to journal:journal and correctly map editors to contributors

## Authors

- [R. S. Doiel](https://orcid.org/0000-0003-0900-6903)
- [Tom Morrell](https://orcid.org/0000-0001-9266-5146)


## Contributors

- [Kathy Johnson](https://orcid.org/0000-0002-0026-2516)
- [George Porter](https://orcid.org/0000-0002-6539-638X)
- [Cynthia Kiser](https://orcid.org/0000-0002-0102-996X)




Tools for working with institutional repositories and data management systems. Current implementation targets Invenio-RDM.

- [License](https://caltechlibrary.github.io/irdmtools/LICENSE)
- [Code Repository](https://github.com/caltechlibrary/irdmtools)
  - [Issue Tracker](https://github.com/caltechlibrary/irdmtools/issues)

## Programming languages

- Go
- Python


## Operating Systems

- Linux
- macOS
- Windows


## Software Requirements

- Go >= 1.26.1
- CMTools >= 0.0.40
- PostgreSQL >= 16
- PostgREST >= 12
- Pandoc >= 3.1
- MySQL >= 8
- SQLite >= 3.49


## Software Suggestions

- GNU Make > 3.8
- PostgreSQL >= 16
- PostgREST >= 12
- Pandoc >= 3.1
- MySQL >= 8
- SQLite >= 3.49


