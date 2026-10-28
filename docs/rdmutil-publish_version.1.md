%rdmutil-publish_version(1) irdmtools user manual | version 0.1.0 6ccc3239
% R. S. Doiel and Tom Morrell
% 2026-10-27

# NAME

rdmutil-publish_version - This will publish a new draft version.

# SYNOPSIS

rdmutil publish_version RECORD_ID [VERSION_LABEL] [PUBLICATION_DATE]

# DESCRIPTION

This will publish a new draft version. RECORD_ID is required. VERSION_LABEL
is optional. PUBLICATION_DATE is optional. PUBLICATION_DATE is in YYYY,
YYYY-MM or YYYY-MM-DD format.  Will need to set .metadata.publication_date,
doi and version name in the versioned draft before publish_version is called.
RECORD_ID is required.

# SEE ALSO

rdmutil(1)
