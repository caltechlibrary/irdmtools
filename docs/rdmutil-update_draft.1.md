%rdmutil-update_draft(1) irdmtools user manual | version 0.1.0 38413f08
% R. S. Doiel and Tom Morrell
% 2026-10-27

# NAME

rdmutil-update_draft - Update a draft record.

# SYNOPSIS

rdmutil update_draft RECORD_ID [FILENAME]

# DESCRIPTION

Update a draft record. RECORD_ID is required. FILENAME is optional, if
one is provided the JSON document is used to update RDM, otherwise standard
input is used to get the JSON required to do the update.

# SEE ALSO

rdmutil(1)
