// doi2rdm is a command line program for harvesting DOI metadata from CrossRef and DataCite returning a JSON documentument sutiable for import into Invenio RDM.
//
// @author R. S. Doiel, <rsdoiel@caltech.edu>
// @author Tom Morrell, <tmorrell@caltech.edu>
//
// Copyright (c) 2023, Caltech
// All rights not granted herein are expressly reserved by Caltech.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice,
// this list of conditions and the following disclaimer.
//
// 2. Redistributions in binary form must reproduce the above copyright notice,
// this list of conditions and the following disclaimer in the documentation
// and/or other materials provided with the distribution.
//
// 3. Neither the name of the copyright holder nor the names of its contributors
// may be used to endorse or promote products derived from this software without
// specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.
package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	// Caltech Library packages
	"github.com/caltechlibrary/irdmtools"
)

var (
	helpText = `%{app_name}(1) irdmtools user manual | version {version} {release_hash}
% R. S. Doiel and Tom Morrell
% {release_date}

# NAME

{app_name}

# SYNOPSIS

{app_name} [OPTIONS] [OPTIONS_YAML] [crossref|datacite] DOI

# DESCRIPTION

{app_name} is a Caltech Library oriented command line application
that takes a DOI, queries the CrossRef or DataCite API then returns a
JSON document suitable for import into Invenio RDM. The DOI can be
in either their canonical form or URL form (e.g. "10.1021/acsami.7b15651" or
"https://doi.org/10.1021/acsami.7b15651").

If a DOI is retrieved the exit code is zero and the JSON document is
the only thing written to standard output. If neither CrossRef nor DataCite
has the DOI the exit code is 1. Other problems have their own exit codes, see
EXIT STATUS.

# OPTIONS_YAML

{app_name} can use an YAML options file to set the behavior of the
crosswalk from CrossRef to RDM. This replaces many of the options
previously required in prior implementations of this tool. See all the
default options setting use the `+"`"+`-show-yaml`+"`"+` command line
options. You can save this to disk, modify it, then use them for
migrating content from CrossRef to RDM.

# TEXT NORMALIZATION

CrossRef returns titles and abstracts as publishers wrote them: with markup,
character entities, MathML formulas, layout whitespace and the subtitle in a
separate field. {app_name} converts three fields of a CrossRef record before
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
still written with the best text available and {app_name} writes one line
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
: something {app_name} read is wrong: the -diff file is not valid JSON, or the
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
: an error nothing classified, which is a bug in {app_name}

74
: writing the JSON failed

75
: CrossRef or DataCite asked {app_name} to try again later (HTTP 429, rate
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
  debugging {app_name} or a CrossRef or DataCite record. It is written once
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
	{app_name} -show-yaml >options.yaml
~~~

Example generating a JSON document for a single DOI. The resulting
text file is called "article.json". In this example "options.yaml"
is the configuration file for setup for your RDM instance. It'll first
check CrossRef then DataCite.

~~~
	{app_name} options.yaml "10.1021/acsami.7b15651" >article.json
~~~

See what CrossRef returned for a DOI, and what {app_name} made of it. The raw
CrossRef JSON and any warnings are in "debug.txt", the RDM record is in
"article.json".

~~~
	{app_name} -debug options.yaml "10.1021/acsami.7b15651" >article.json 2>debug.txt
~~~

Check to see the difference from the saved "article.json" and
the current metadata retrieved from CrossRef or DataCite.

~~~
	{app_name} -diff article.json options.yaml "10.1021/acsami.7b15651"
~~~

Example getting metadata for an arXiv record from DataCite

~~~
	{app_name} options.yaml "arXiv:2312.07215"
~~~

`
)


func main() {
	appName := path.Base(os.Args[0])
	// NOTE: the following are set when version.go is generated
	version := irdmtools.Version
	releaseDate := irdmtools.ReleaseDate
	releaseHash := irdmtools.ReleaseHash
	fmtHelp := irdmtools.FmtHelp

	showHelp, showVersion, showLicense := false, false, false
	debug, showYAML := false, false
	diffFName := ""
	flag.BoolVar(&showHelp, "help", false, "display help")
	flag.BoolVar(&showVersion, "version", false, "display version")
	flag.BoolVar(&showLicense, "license", false, "display license")
	flag.BoolVar(&showYAML, "show-yaml", false, "display the YAML configuration")
	flag.StringVar(&diffFName, "diff", diffFName, "compare the JSON file with the current record generated from CrossRef")
	flag.BoolVar(&debug, "debug", debug, "display additional info to stderr")
	flag.Parse()
	args := flag.Args()

	in := os.Stdin
	out := os.Stdout
	eout := os.Stderr

	if showHelp {
		fmt.Fprintf(out, "%s\n", fmtHelp(helpText, appName, version, releaseDate, releaseHash))
		os.Exit(0)
	}
	if showVersion {
		fmt.Fprintf(out, "%s %s %s\n", appName, version, releaseHash)
		os.Exit(0)
	}
	if showLicense {
		fmt.Fprintf(out, "%s %s\n", appName, version)
		fmt.Fprintf(out, "%s\n", irdmtools.LicenseText)
		os.Exit(0)
	}
	if showYAML {
		fmt.Fprintf(out, "%s\n", irdmtools.DefaultDoi2RdmOptionsYAML)
		os.Exit(0)
	}
	// Create a appity object
	app := new(irdmtools.Doi2Rdm)
	app.Cfg = new(irdmtools.Config)
	if debug {
		app.Cfg.Debug = true
	} else {
		app.Cfg.Debug = false
	}

	optionsFName, dataSource, doi := "", "", ""
	if len(args) < 1 {
		fmt.Fprintln(eout, "expected at least a single DOI on the command line")
		os.Exit(int(irdmtools.ExitUsage))
	} else if len(args) == 1 {
		optionsFName, dataSource, doi = "", "", args[0]
	} else if len(args) == 2 {
		optionsFName, dataSource, doi = args[0], "", args[1]
	} else if len(args) > 2 {
		optionsFName, dataSource, doi = args[0], args[1], args[2]
	} else {
		dataSource, doi = args[0], args[1]
	}
	switch dataSource {
		case "crossref":
			if exitCode, err := app.RunCrossRefToRdm(in, out, eout, optionsFName, doi, diffFName); err != nil {
				fmt.Fprintf(eout, "%s\n", err)
				os.Exit(exitCode)
			}
		case "datacite":
			if exitCode, err := app.RunDataCiteToRdm(in, out, eout, optionsFName, doi, diffFName); err != nil {
				fmt.Fprintf(eout, "%s\n", err)
				os.Exit(exitCode)
			}
		default:
			if exitCode, err := app.RunDoiToRdmCombined(in, out, eout, optionsFName, doi, diffFName); err != nil {
				fmt.Fprintf(eout, "%s\n", err)
				os.Exit(exitCode)
			}
	}
}
