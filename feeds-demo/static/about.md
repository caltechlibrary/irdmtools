---
title : About feeds.library.caltech.edu
version: 1.5
---

# About feeds.library.caltech.edu

[feeds.library.caltech.edu](/) is a metadata service provided
by Caltech Library.  It contains metadata harvested from our
institutional repositories, public directory and archival systems.
It is part of an initiative to allow data science techniques to
be used on library and archive curated metadata for both internal
and external needs.

## Implementation

- Data harvested using irdmtools[1] and dataset[2]
- Site generation irdmtools, dataset, datatools[3], python and Pandoc[4]
- Search implemented using Pagefind[5]

The content of feeds can easily be hosted in a bucket
oriented object store (e.g. S3, Minio buckets). Thus
all content is calculated and pre-rendered. This doesn't mean
the content can not be made interactive. Modern browser
applications can easily be built from the metadata if you
know what you're looking for and the path to retrieve an appropriate
JSON document. The feed search implementation is a good examples
as well as our "widget builder" used to make it easy to integrated
feed content into your favorite CMS.

The implementation first harvests contents from our curated
data sources (i.e. CaltechAUTHORS, CaltechDATA, CaltechTHESIS,
CaltechGROUPS and CaltechPEOPLE)
as [dataset](https://caltechlibrary.github.io/dataset) collections.
These are used to aggregated and generate the JSON documents in the
feeds web tree. After which various other formatted content is generated.

More details can be found in the changes document for
[1.5](v1.5-changes.md "Changes to feeds from v1.0.x to v1.5")  

[1]: https://caltechlibrary.github.io/irdmtools
[2]: https://caltechlibrary.github.io/dataset
[3]: https://caltechlibrary.github.io/datatools
[4]: https://pandoc.org
[5]: https://pagefind.app


