%rdmutil-get_record(1) irdmtools user manual | version 0.1.0 6ccc3239
% R. S. Doiel and Tom Morrell
% 2026-10-27

# NAME

rdmutil-get_record - Returns a specific simplified record indicated by RECORD_ID, e.g. bq3se-47g50.

# SYNOPSIS

rdmutil get_record RECORD_ID

# DESCRIPTION

Returns a specific simplified record indicated by RECORD_ID, e.g. bq3se-47g50.
The RECORD_ID is a required parameter.  If Postgres access is defined then it
does this by SQL assembling an object similar to the one available via 
"/api/records/RECORD_ID" path in the JSON API. If Postgres is not configured
then it retrieves the record via the JSON API.

# SEE ALSO

rdmutil(1)
