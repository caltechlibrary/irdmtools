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

# OPTIONS_YAML

{app_name} can use an YAML options file to set the behavior of the
crosswalk from CrossRef to RDM. This replaces many of the options
previously required in prior implementations of this tool. See all the
default options setting use the `+"`"+`-show-yaml`+"`"+` command line
options. You can save this to disk, modify it, then use them for
migrating content from CrossRef to RDM.

# OPTIONS

-help
: display help

-license
: display license

-version
: display version

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
		fmt.Fprintln(eout, "expected a least a single DOI on the command line")
		os.Exit(1)
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
			if err := app.RunCrossRefToRdm(in, out, eout, optionsFName, doi, diffFName); err != nil {
				fmt.Fprintf(eout, "%s\n", err)
				os.Exit(1)
			}
		case "datacite":
			if err := app.RunDataCiteToRdm(in, out, eout, optionsFName, doi, diffFName); err != nil {
				fmt.Fprintf(eout, "%s\n", err)
				os.Exit(1)
			}
		default:
			if err := app.RunDoiToRdmCombined(in, out, eout, optionsFName, doi, diffFName); err != nil {
				fmt.Fprintf(eout, "%s\n", err)
				os.Exit(1)
			}
	}
}
