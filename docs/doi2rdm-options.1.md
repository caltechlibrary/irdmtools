%doi2rdm-options(1) irdmtools user manual | version 0.1.0 38413f08
% R. S. Doiel and Tom Morrell
% 2026-10-27

# NAME

doi2rdm-options - -help, -license, -version, -debug, -diff, -show-yaml

# SYNOPSIS

doi2rdm -help options

# DESCRIPTION

-help
: display help. Use -help topics to list the help topics and -help TOPIC to show the page for one.

-license
: display license

-version
: display version

-debug
: write the JSON that CrossRef ("works JSON:") or DataCite ("objects JSON:")
  returned to standard error, before it is converted. Use it to tell a problem
  in the service's own data from a problem in the conversion, whether you are
  debugging doi2rdm or a CrossRef or DataCite record. It is written once
  the service has answered, so a DOI that is not found shows nothing. Standard
  output is not changed. The output can be long, so save it with 2>FILENAME.

-diff JSON_FILENAME
: compare the JSON_FILENAME contents with record generated from CrossRef or DataCite works record

-show-yaml
: This will display the default YAML configuration file. You can save this and customize to suit your needs.

# SEE ALSO

doi2rdm(1)
