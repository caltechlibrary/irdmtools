%doi2rdm-exit_status(1) irdmtools user manual | version 0.1.0 38413f08
% R. S. Doiel and Tom Morrell
% 2026-10-27

# NAME

doi2rdm-exit_status - Standard output is the JSON document and nothing else.

# SYNOPSIS

doi2rdm -help exit_status

# DESCRIPTION

Standard output is the JSON document and nothing else. The exit codes are the
Caltech Library workspace's standard ones (sysexits), so a script can tell what
went wrong from the number alone.

0
: the record was written (any warnings went to standard error)

1
: the DOI was not found: CrossRef and DataCite both answered that they have no
  record for it

2
: the command line is wrong: no DOI was given, or an option is not known

65
: something doi2rdm read is wrong: the -diff file is not valid JSON, or the
  record CrossRef or DataCite returned could not be converted (for example it
  has a resource type the options do not map; the message names it)

66
: a file named on the command line or in the options is missing: the options
  file, the -diff file or the symbols file

69
: CrossRef or DataCite could not be reached or failed (a network error, a
  timeout or an HTTP error such as 500 or 503). The DOI may exist; try again
  later

70
: an error nothing classified, which is a bug in doi2rdm

74
: writing the JSON failed

75
: CrossRef or DataCite asked doi2rdm to try again later (HTTP 429, rate
  limited)

78
: a configuration file is present but wrong: the options file is not valid
  YAML, or the symbols file named in the options is not valid

If neither service has the DOI the exit code is 1, but only when both answered
that they have no record. If either was unavailable and neither returned a
record, the exit code is that service's (CrossRef's first), because the DOI may
still exist. A problem with a file or with the record itself stops the run
at once and is not hidden by trying the other service.

Earlier versions used 2 for "not found" and 8 for an invalid file. A script that
read those must be updated: "not found" is now 1 and 2 is a command line error.

# SEE ALSO

doi2rdm(1)
