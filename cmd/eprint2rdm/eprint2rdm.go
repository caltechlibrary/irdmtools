// eprint2rdm is a command line program for harvesting an EPrint metadata record and return a Invenio RDM style record.
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
	"strings"

	// Caltech Library packages
	"github.com/caltechlibrary/irdmtools"
)

var (
	helpText = `% {app_name}(1) {app_name} user manual | Version {version}
% R. S. Doiel and Tom Morrell
% 2023-03-30

# NAME

{app_name}

# SYNOPSIS

{app_name} [OPTIONS] EPRINT_HOSTNANE EPRINT_ID

# DESCRIPTION

{app_name} is a Caltech Library centric command line application
that takes an EPrint hostname and EPrint ID and returns a JSON
document suitable to import into Invenio RDM. It relies on
access to EPrint's REST API. It uses EPRINT_USER and EPRINT_PASSWORD
environment variables to access the API. Using the "-keys" options
you can get a list of keys available from the EPrints REST API.

# OPTIONS

-help
: display help

-license
: display license

-version
: display version

-all-ids
: return a list of EPrint ids, one per line.

-resource-map FILENAME
: use this comma delimited resource map from EPrints to RDM resource types.
The resource map file is a comma delimited file without a header row.
first column is the EPrint resource type string, the second is the
RDM resource type string.


# EXAMPLE


Example generating a JSON document for from the EPrints repository
hosted as "eprints.example.edu" for EPrint ID 118621.  Access to
the EPrint REST API is configured in the environment.  The result
is saved in "article.json".

~~~
EPRINT_USER="__USERNAME_GOES_HERE__"
EPRINT_PASSWORD="__PASSWORD_GOES_HERE__"
{app_name} eprints.example.edu 118621 \
	>article.json
~~~

Generate a list of EPrint ids from a repository (e.g. eprints.example.edu).

~~~
{app_name} -all-ids eprints.example.edu >eprintids.txt
~~~

Generate a JSON document from the EPrints repository
hosted as "eprints.example.edu" for EPrint ID 118621 using a
resource map file to map the EPrints resource type to an
Invenio RDM resource type.

~~~
{app_name} --resource-map resource-types.csv \
      eprints.example.edu 118621 \
	  >article.json
~~~

Putting it together in the to harvest an EPrints repository
saving the results in a dataset collection for analysis or
migration.

1. create a dataset collection
2. get the EPrint ids to harvest applying a resource type map, "resource-types.csv"
3. Harvest the eprint records and save in our dataset collection

~~~
dataset init example_edu.ds
{app_name} -all-ids eprints.example.edu >eprintids.txt
while read EPRINTID; do
    {app_name} -resource-map resource-types.csv \
       eprints.example.edu "${EPRINTID}" |\
	   dataset create -i - example_edu.ds "${EPRINTID}"
done <eprintids.txt
~~~

At this point you would be ready to improve the records in
example_edu.ds before migrating them into Invenio RDM.
`
)

func fmtTxt(txt string, appName string) string {
	return strings.ReplaceAll(strings.ReplaceAll(txt, "{app_name}", appName), "{version}", irdmtools.Version)
}

func main() {
	appName := path.Base(os.Args[0])
	showHelp, showVersion, showLicense := false, false, false
	allIds, debug := false, false
	resourceTypesFName := ""
	flag.BoolVar(&showHelp, "help", false, "display help")
	flag.BoolVar(&showVersion, "version", false, "display version")
	flag.BoolVar(&showLicense, "license", false, "display license")
	flag.BoolVar(&debug, "debug", debug, "display additional info to stderr")
	flag.BoolVar(&allIds, "all-ids", false, "retrieve all the eprintids from an EPrints repository via REST API, one per line")
	flag.StringVar(&resourceTypesFName, "resource-map", resourceTypesFName, "use this file to map resource types from EPrints to Invenio RDM")
	flag.Parse()
	args := flag.Args()

	eprintUser := os.Getenv("EPRINT_USER")
	eprintPassword := os.Getenv("EPRINT_PASSWORD")

	if showHelp {
		fmt.Fprintf(os.Stdout, "%s\n", fmtTxt(helpText, appName))
		os.Exit(0)
	}
	if showVersion {
		fmt.Fprintf(os.Stdout, "%s %s\n", appName, irdmtools.Version)
		os.Exit(0)
	}
	if showLicense {
		fmt.Fprintf(os.Stdout, "%s %s\n", appName, irdmtools.Version)
		fmt.Fprintf(os.Stdout, "%s\n", irdmtools.LicenseText)
		os.Exit(0)
	}

	// Create a appity object
	host, eprintid := "", ""
	app := new(irdmtools.EPrint2Rdm)
	if allIds {
		if len(args) != 1 {
			fmt.Fprintf(os.Stderr, "expected an EPrint hostname and -keys option")
			os.Exit(1)
		} else {
			host = args[0]
		}
	} else {
		if len(args) != 2 {
			fmt.Fprintf(os.Stderr, "expected an EPrint hostname and EPrint ID")
			os.Exit(1)
		} else {
			host, eprintid = args[0], args[1]
		}
	}
	if err := app.Run(os.Stdin, os.Stdout, os.Stderr, eprintUser, eprintPassword, host, eprintid, resourceTypesFName, allIds, debug); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
