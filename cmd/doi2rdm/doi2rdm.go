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

{app_name} [OPTIONS] DOI

# DESCRIPTION

{app_name} is a Caltech Library centric command line application
that takes a DOI, queries the CrossRef API then returns a JSON document
suitable for import into Invenio RDM. The DOI can be in either their
canonical form or URL form (e.g. "10.1021/acsami.7b15651" or
"https://doi.org/10.1021/acsami.7b15651").

# OPTIONS

-help
: display help

-license
: display license

-version
: display version

-diff JSON_FILENAME
: compare the JSON_FILENAME contents with record generated from CrossRef works record

-dot-initials
: Add period to initials in given name

-download
: attempt to download the digital object if object URL provided

-mailto
: (string) set the mailto value for CrossRef API access (default "helpdesk@library.caltech.edu")

-resource-map
: Use this two column CSV file (no header row) to map resource types in CrossRef to RDM

-contributor-map
: Use this two column CSV file (no header row) to map contributor types from CrossRef (e.g.
"author", "translator", "editor", "chair") to RDM roles.

# EXAMPLES

Example generating a JSON document for a single DOI. The resulting
text file is called "article.json".

~~~
	{app_name} "10.1021/acsami.7b15651" >article.json
~~~

Check to see the difference from the saved "article.json" and
the current metadata retrieved from CrossRef.

~~~
	{app_name} -diff article.json "10.1021/acsami.7b15651
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
	debug, downloadDocument := false, false
	dotInitials := false
	mailTo, diffFName := "", ""
	resourceTypeFName, contributorTypeFName := "", ""
	flag.BoolVar(&showHelp, "help", false, "display help")
	flag.BoolVar(&showVersion, "version", false, "display version")
	flag.BoolVar(&showLicense, "license", false, "display license")
	flag.StringVar(&diffFName, "diff", diffFName, "compare the JSON file with the current record generated from CrossRef")
	flag.BoolVar(&dotInitials, "dot-initials", dotInitials, "Add period to initials in given name")
	flag.BoolVar(&downloadDocument, "download", downloadDocument, "attempt to download the digital object if object URL provided")
	flag.StringVar(&mailTo, "mailto", mailTo, "set the mail to value for CrossRef API access")
	flag.StringVar(&resourceTypeFName, "resource-map", resourceTypeFName, "Use this CSV to map resource types from CrossREF to RDM")
	flag.StringVar(&contributorTypeFName, "contributor-map", contributorTypeFName, "Use this CSV to map contributor types from CrossREF to RDM")
	flag.BoolVar(&debug, "debug", debug, "display additional info to stderr")
	flag.Parse()
	args := flag.Args()

	if showHelp {
		fmt.Fprintf(os.Stdout, "%s\n", fmtHelp(helpText, appName, version, releaseDate, releaseHash))
		os.Exit(0)
	}
	if showVersion {
		fmt.Fprintf(os.Stdout, "%s %s %s\n", appName, version, releaseHash)
		os.Exit(0)
	}
	if showLicense {
		fmt.Fprintf(os.Stdout, "%s %s\n", appName, version)
		fmt.Fprintf(os.Stdout, "%s\n", irdmtools.LicenseText)
		os.Exit(0)
	}
	// Create a appity object
	app := new(irdmtools.Doi2Rdm)
	app.Cfg = new(irdmtools.Config)
	options := map[string]string{}
	if resourceTypeFName != "" {
		options["resource-map"] = resourceTypeFName
	}
	if contributorTypeFName != "" {
		options["contributor-map"] = contributorTypeFName
	}
	if diffFName != "" {
		options["diff"] = diffFName
	}
	if dotInitials {
		options["dot_initials"] = "true"
	}
	if downloadDocument {
		options["download_document"] = "true"
	}
	if mailTo != "" {
		options["mailto"] = mailTo
	}
	if debug {
		app.Cfg.Debug = true
	} else {
		app.Cfg.Debug = false
	}

	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "expected a single DOI on the command line")
		os.Exit(1)
	}
	if err := app.Run(os.Stdin, os.Stdout, os.Stderr, options, args[0]); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
