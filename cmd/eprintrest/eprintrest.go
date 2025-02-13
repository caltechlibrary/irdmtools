// eprintrest is a command line program that re-creates a EPrints 3.x REST
// API running on localhost. It requires access to the repository's 
// "archives" directory as well as the MySQL database.
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

{app_name} [OPTIONS]

# DESCRIPTION

{app_name} is a Caltech Library oriented localhost web service
that creates a functionally similar replica of the EPrints REST API 
for EPrints 3.3.x based repositories. It uses the path to the 
"archives" directory and a MySQL Database for the repository. 
It only supports "archive" eprint.eprint_status records and
only the complete XML. Start up time is slow because it builds 
the data structures representing the content in memory. This
makes the response times to request VERY fast compared to
the EPrints REST API.

NOTE: the rest API does not enforce user permissions, restrictions
or roles. It is a minimal READ ONLY re-implementation of the EPrints 3.3
REST API!

# ENVIRONMENT

The application is configured from the environment. The following
environment variables need to be set. The environment variables can
be set at the shell level or in a ".env" file.

REPO_ID
: The repository id string (e.g. caltechauthors). Also the name of the database for the repository.

EPRINT_ARCHIVES_PATH
: A path to the "archives" directory holding your repository content 
(e.g. /usr/local/eprints/archives)

DB_USER
: The user name needed to access the MySQL database[^1]

DB_PASSWORD
: The password needed to access the MySQL database[^1]

REST_PORT
: The localhost port to use for the read only REST API.

[^1]: MySQL, like this REST service assumes to be running on localhost.


# OPTIONS

-help
: display help

-license
: display license

-version
: display version


# EXAMPLE

This is an example environment

~~~
REPO_ID="caltechauthors"
EPRINT_ARCHIVES_PATH="/code/eprints3.3/archives"
REST_PORT=80
DB_USER="eprints"
DB_PASSWORD="something_secret_here"
~~~

Running the localhost REST API clone

~~~
{app_name}
~~~

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
	flag.BoolVar(&showHelp, "help", false, "display help")
	flag.BoolVar(&showVersion, "version", false, "display version")
	flag.BoolVar(&showLicense, "license", false, "display license")

	flag.Parse()

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
	app := new(irdmtools.EPrintRest)
	if err := app.Run(os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
