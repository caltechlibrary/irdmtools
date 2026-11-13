%rdmutil-new_version(1) irdmtools user manual | version 0.2.0 c655545a
% R. S. Doiel and Tom Morrell
% 2026-11-12

# NAME

rdmutil-new_version - This will create a new version of the record.

# SYNOPSIS

rdmutil new_version RECORD_ID

# DESCRIPTION

This will create a new version of the record. RECORD_ID is required.
NOTE: When you create a new version .metadata.publication_date and 
.metadata.version are removed.  These need to be you recreated in the draft
of the new version as well as to handle any files, access restrictions before
publishing the new version draft.

# SEE ALSO

rdmutil(1)
