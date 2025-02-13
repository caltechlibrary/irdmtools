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

{app_name} [OPTIONS] [EPRINT_HOST] EPRINT_ID

# DESCRIPTION

{app_name} is a Caltech Library oriented command line application
that takes an EPrint hostname and EPrint ID and returns a JSON
document suitable to import into Invenio RDM. It relies on
access to EPrint's REST API. It uses EPRINT_USER, EPRINT_PASSWORD
and EPRINT_HOST environment variables to access the API. Using
the "-all-ids" options you can get a list of keys available from
the EPrints REST API.

{app_name} can harvest a set of eprint ids into a dataset collection
using the "-id-list" and "-harvest" options. You map also provide
customized resource type and person role mapping for the content
you harvest. This will allow you to be substantially closer to the
final record form needed to crosswalk EPrints data into Invenio RDM.

# ENVIRONMENT

Environment variables can be set at the shell level or in a ".env" file.

EPRINT_USER
: The eprint user id to access the REST API

EPRINT_PASSWORD
: The eprint user password to access the REST API

EPRINT_HOST
: The hostname of the EPrints service


# OPTIONS

-help
: display help

-license
: display license

-version
: display version

-all-ids
: return a list of EPrint ids, one per line.

-harvest DATASET_NAME
: Harvest content to a dataset collection rather than standard out

-id-list ID_FILE_LIST
: (used with harvest) Retrieve records based on the ids in a file,
one line per id.

-resource-map FILENAME
: use this comma delimited resource map from EPrints to RDM resource types.
The resource map file is a comma delimited file without a header row.
The First column is the EPrint resource type string, the second is the
RDM resource type string.

-contributor-map FILENAME
: use this comma delimited contributor type map from EPrints to RDM
contributor types.  The contributor map file is a comma delimited file
without a header row. The first column is the value stored in the EPrints
table "eprint_contributor_type" and the second value is the string used
in the RDM instance.

# EXAMPLE


Example generating a JSON document for from the EPrints repository
hosted as "eprints.example.edu" for EPrint ID 118621.  Access to
the EPrint REST API is configured in the environment.  The result
is saved in "article.json". EPRINT_USER, EPRINT_PASSWORD and
EPRINT_HOST (e.g. eprints.example.edu) via the shell environment.

~~~
EPRINT_USER="__USERNAME_GOES_HERE__"
EPRINT_PASSWORD="__PASSWORD_GOES_HERE__"
EPRINT_HOST="eprints.example.edu"
{app_name} 118621 >article.json
~~~

Generate a list of EPrint ids from a repository 

~~~
{app_name} -all-ids >eprintids.txt
~~~

Generate a JSON document from the EPrints repository
hosted as "eprints.example.edu" for EPrint ID 118621 using a
resource map file to map the EPrints resource type to an
Invenio RDM resource type and a contributor type map for
the contributors type between EPrints and RDM.

~~~
{app_name} -resource-map resource_types.csv \
      -contributor-map contributor_types.csv \
      eprints.example.edu 118621 \
	  >article.json
~~~

Putting it together in the to harvest an EPrints repository
saving the results in a dataset collection for analysis or
migration.

1. create a dataset collection
2. get the EPrint ids to harvest applying a resource type map, "resource_types.csv"
   and "contributor_types.csv" for contributor type mapping
3. Harvest the eprint records and save in our dataset collection

~~~
dataset init eprints.ds
{app_name} -all-ids >eprintids.txt
{app_name} -id-list eprintids.txt -harvest eprints.ds
~~~

At this point you would be ready to improve the records in
eprints.ds before migrating them into Invenio RDM.
`
)

func main() {
	appName := path.Base(os.Args[0])
	// NOTE: The following are set when version.go is generated
	version := irdmtools.Version
	releaseDate := irdmtools.ReleaseDate
	releaseHash := irdmtools.ReleaseHash
	fmtHelp := irdmtools.FmtHelp

	showHelp, showVersion, showLicense := false, false, false
	allIds, debug := false, false
	idList, cName, configFName := "", "", ""
	resourceTypesFName, contributorTypesFName := "", ""
	flag.BoolVar(&showHelp, "help", false, "display help")
	flag.BoolVar(&showVersion, "version", false, "display version")
	flag.BoolVar(&showLicense, "license", false, "display license")
	flag.BoolVar(&debug, "debug", debug, "display additional info to stderr")
	flag.BoolVar(&allIds, "all-ids", false, "retrieve all the eprintids from an EPrints repository via REST API, one per line")
	flag.StringVar(&idList, "id-list", idList, "retrieve the record ids in the list")
	flag.StringVar(&cName, "harvest", cName, "harvest the record into a dataset collection")
	flag.StringVar(&resourceTypesFName, "resource-map", resourceTypesFName, "use this file to map resource types from EPrints to Invenio RDM")
	flag.StringVar(&contributorTypesFName, "contributor-map", contributorTypesFName, "use this file to map contributor types from EPrints to Invenio RDM")
	flag.StringVar(&configFName, "config", configFName, "user config file")
	flag.Parse()
	args := flag.Args()

	eprintUser := os.Getenv("EPRINT_USER")
	eprintPassword := os.Getenv("EPRINT_PASSWORD")
	eprintHostname := os.Getenv("EPRINT_HOST")
	eprintid := ""

	if showHelp {
		fmt.Fprintf(os.Stdout, "%s\n", fmtHelp(helpText, appName, version, releaseDate, releaseHash))
		os.Exit(0)
	}
	if showVersion {
		fmt.Fprintf(os.Stdout, "%s %s %s\n", appName, version, releaseHash)
		os.Exit(0)
	}
	if showLicense {
		fmt.Fprintf(os.Stdout, "%s\n", irdmtools.LicenseText)
		os.Exit(0)
	}

	// Create a appity object
	app := new(irdmtools.EPrint2Rdm)
	if err := app.Configure(configFName, "", debug); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
	if (allIds || idList != "") && eprintHostname == "" {
		if len(args) != 1 {
			fmt.Fprintf(os.Stderr, "expected an EPrint hostname with either -all-ids or -ids-list and -harvest options")
			os.Exit(1)
		} else {
			eprintHostname = args[0]
		}
	} else {
		if eprintHostname != "" && len(args) == 1 {
			eprintid = args[0]
		} else if len(args) != 2 {
			fmt.Fprintf(os.Stderr, "expected an EPrint hostname and EPrint ID")
			os.Exit(1)
		} else {
			eprintHostname, eprintid = args[0], args[1]
		}
	}
	if err := app.Run(os.Stdin, os.Stdout, os.Stderr, eprintUser, eprintPassword, eprintHostname, eprintid, resourceTypesFName, contributorTypesFName, allIds, idList, cName, debug); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
