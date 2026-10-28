%rdmutil-get_modified_ids(1) irdmtools user manual | version 0.1.0 38413f08
% R. S. Doiel and Tom Morrell
% 2026-10-27

# NAME

rdmutil-get_modified_ids - Returns a list of modified record ids (created, updated, deleted) in the time range listed.

# SYNOPSIS

rdmutil get_modified_ids START [END]

# DESCRIPTION

Returns a list of modified record ids (created, updated, deleted) in the
time range listed.  If Postgres is not configured this method uses OAI-PMH
for id retrieval. It is rate limited.  Start and end dates are inclusive
and should be specific in YYYY-MM-DD format.

# SEE ALSO

rdmutil(1)
