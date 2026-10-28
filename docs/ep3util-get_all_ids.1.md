%ep3util-get_all_ids(1) irdmtools user manual | version 0.1.0 6ccc3239
% R. S. Doiel and Tom Morrell
% 2026-10-27

# NAME

ep3util-get_all_ids - Returns a list of all repository record ids.

# SYNOPSIS

ep3util get_all_ids

# DESCRIPTION

Returns a list of all repository record ids. The method uses OAI-PMH for id retrieval. It is rate limited and will take come time to return all record ids. A test instance took 11 minutes to retrieve 24000 record ids.

# SEE ALSO

ep3util(1)
