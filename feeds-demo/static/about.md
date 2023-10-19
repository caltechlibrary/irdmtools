---
title : About feeds.library.caltech.edu
---

# About feeds.library.caltech.edu

[feeds.library.caltech.edu](/) is a metadata service provided
by the Caltech Library. It contains metadata harvested from
our institutional repositories, public directory and archival
systems. It is part of an initiative to allow data science
techniques to be used on library and archive curated metadata
for both internal and external needs.

## Implementation

+ [Harvest](./#harvest)
+ [Feed Generation](./#genfeeds)
+ [Website UI](./#website)

The content of feeds can easily be hosted in a bucket
oriented object store (e.g. S3, Minio buckets). Thus
all content is calculated and pre-rendered. This doesn't mean
the content can not be made interactive. Modern browser
applications can easily be built from the metadata if you
know what you're looking for and the path to retrieve an appropriate
JSON document.

The implementation first harvests contents from our curated
data sources (e.g. CaltechAUTHORS, CaltechDATA, CaltechTHESIS),
as [dataset](https://caltechlibrary.github.io/dataset) collections.
These are used to create some aggregated collections such as
those needed to build people data.

After all the dataset collections are created/updated a process
is run to create JSON documents used as specific fields. The
JSON documents are created to be web accessible. These include
documents representing a people, publication types and
counts, publication lists (e.g. combined, articles, books),
data repository items (by resource type) and thesis feeds. This is
done for both individual people as well as group listings (e.g.
GALCIT, LIGO).

Once all the JSON feed documents are rendered we use those to
generate Markdown documents for a webview of the metadata. These
are then rendered as their final HTML and HTML include versions
of documents.
