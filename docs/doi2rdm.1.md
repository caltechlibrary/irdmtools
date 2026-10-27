%doi2rdm(1) irdmtools user manual | version 0.0.99 a66c9ddf
% R. S. Doiel and Tom Morrell
% 2026-10-27

# NAME

doi2rdm

# SYNOPSIS

doi2rdm [OPTIONS] [OPTIONS_YAML] [crossref|datacite] DOI

# DESCRIPTION

doi2rdm is a Caltech Library oriented command line application
that takes a DOI, queries the CrossRef or DataCite API then returns a
JSON document suitable for import into Invenio RDM. The DOI can be
in either their canonical form or URL form (e.g. "10.1021/acsami.7b15651" or
"https://doi.org/10.1021/acsami.7b15651").

If a DOI is retrieved the exit code is zero and the JSON document is
the only thing written to standard output. If neither CrossRef nor DataCite
has the DOI the exit code is 1. Other problems have their own exit codes, see
EXIT STATUS.

# OPTIONS_YAML

doi2rdm can use an YAML options file to set the behavior of the
crosswalk from CrossRef to RDM. This replaces many of the options
previously required in prior implementations of this tool. See all the
default options setting use the `-show-yaml` command line
options. You can save this to disk, modify it, then use them for
migrating content from CrossRef to RDM.

# TEXT NORMALIZATION

CrossRef returns titles and abstracts as publishers wrote them: with markup,
character entities, MathML formulas, layout whitespace and the subtitle in a
separate field. doi2rdm converts three fields of a CrossRef record before
it writes the JSON, using the text normalization of the crossrefapi package.

title
: the CrossRef title as plain text (tags removed, entities decoded, formulas
  converted to Unicode, layout whitespace tidied) with the subtitle appended
  after a colon (after a space only, if the title already ends in a period,
  question mark, exclamation mark or colon), unless the title already holds
  the subtitle.

additional title
: the other CrossRef titles that add something to the first one, as plain
  text, each with the type "other". A second title that only differs by
  capitalization is not added. The first CrossRef title is the title.

description
: the abstract as HTML that Invenio RDM keeps: paragraphs, headings as bold
  paragraphs, and inline markup such as italics, with formulas written as TeX
  between \( and \). The leading "Abstract" heading and the publisher's
  copyright line are dropped. A record with no abstract has no description.
  The page that shows the description needs MathJax, or another renderer that
  reads the \( \) delimiters, for the formulas to appear as formulas.

DataCite records are not changed by this.

To change the table of symbols the conversion uses, name a crossrefapi
override file in the options YAML (a relative path is read relative to the
options file):

~~~
symbols: symbols.json
~~~

Without it the table built into crossrefapi is used. A missing file exits with
66 and an invalid one with 78, before any record is fetched.

# WARNINGS

When part of a title or abstract cannot be converted cleanly, the record is
still written with the best text available and doi2rdm writes one line
for each such case to standard error (stderr):

~~~
doi2rdm: DOI: field: code: fragment
~~~

field is "title", "additional title" or "description". code is one of
"parse-fallback" (the markup could not be parsed and was kept as text),
"script-unmapped" (a subscript or superscript has no Unicode form),
"math-linear" (a formula was written on one line, for example Z_N),
"math-unconverted" (a formula element was left as text) and
"multiple-titles" (CrossRef gave more than one title; the first is used and the
fragment is the others). fragment is the part of the source the warning is
about. Control characters in a fragment, such as terminal escape sequences,
are replaced by spaces, so a warning is always one line of printable text.
Warnings never change the JSON written to standard output or the exit
status. With -diff only the warnings for the current record are written.

# EXIT STATUS

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

# OPTIONS

-help
: display help

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

# EXAMPLES

Save the default YAML options to a file. You can customize this to match your
vocabulary requirements in your RDM deployment.

~~~
	doi2rdm -show-yaml >options.yaml
~~~

Example generating a JSON document for a single DOI. The resulting
text file is called "article.json". In this example "options.yaml"
is the configuration file for setup for your RDM instance. It'll first
check CrossRef then DataCite.

~~~
	doi2rdm options.yaml "10.1021/acsami.7b15651" >article.json
~~~

See what CrossRef returned for a DOI, and what doi2rdm made of it. The raw
CrossRef JSON and any warnings are in "debug.txt", the RDM record is in
"article.json".

~~~
	doi2rdm -debug options.yaml "10.1021/acsami.7b15651" >article.json 2>debug.txt
~~~

Check to see the difference from the saved "article.json" and
the current metadata retrieved from CrossRef or DataCite.

~~~
	doi2rdm -diff article.json options.yaml "10.1021/acsami.7b15651"
~~~

Example getting metadata for an arXiv record from DataCite

~~~
	doi2rdm options.yaml "arXiv:2312.07215"
~~~


