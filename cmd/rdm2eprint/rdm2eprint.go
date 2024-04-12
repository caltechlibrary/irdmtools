// rdm2eprint is a command line program for harvesting an RDM record and rendering it as a EPrint 3.3 record.
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
	"io"
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

{app_name} [OPTIONS] RDM_ID

# DESCRIPTION

{app_name} is a Caltech Library oriented command line application
that takes an RDM record ID and returns a EPrint record JSON document.
It was created to allow us migrate our EPrints repositories minimal change
to our feeds system which works with EPrint structured data.
It uses RDM_URL, RDMTOK, RDM_COMMUNITY_ID environment variables for
configuration.  It can read data from a previously harvest RDM record
or directly from RDM via the API url. The tool is intended to run
in a pipe line so have minimal options.

# OPTIONS

-help
: display help

-license
: display license

-version
: display version

-config
: provide a path to an alternate configuration file (e.g. "irdmtools.json")

-harvest C_NAME
: harvest JSON formatted eprint records into the dataset collection 
specified by C_NAME.

-ids JSON_ID_FILE
: read ids from a file.

-xml
: output as EPrint XML rather than JSON, does not work with -harvest.

-pipeline
: read from standard input and write crosswalk to standard out.

-latest
: only convert record(s) if latest version.

# EXAMPLE

Example generating a EPRINT JSON document from RDM would use the following
variables.

the EPrints repository
hosted as "eprints.example.edu" for EPrint ID 118621.  Access to
the EPrint REST API is configured in the environment.  The result
is saved in "article.json". EPRINT_USER, EPRINT_PASSWORD and
EPRINT_HOST (e.g. eprints.example.edu) via the shell environment.

~~~
RDM_URL="__URL_TO_RDM_INSTANCE_HERE__"
RDMTOK="__RDM_ACCESS_TOKEN_HERE__"
RDM_COMMUNITY_ID="rdm.example.edu"
{app_name} k3tpc-ga970 >article.json
~~~

`
)

// getRdmIds will read in a JSON list of RDM ids from either standard
// input or a JSON file.
func getRdmIds(idsFName string) ([]string, error) {
	var err error
 	in := os.Stdin
  	if idsFName != "-" {
   		in, err = os.Open(idsFName)
   		if err != nil {
   			return nil, err
   		}
   		defer in.Close()
   	}
   	src, err := io.ReadAll(in)
   	if err != nil {
		return nil, err
   	}
   	ids := []string{}
   	if err := irdmtools.JSONUnmarshal(src, &ids); err != nil {
		return nil, err
   	}
	return ids, nil
}

func main() {
	appName := path.Base(os.Args[0])
	// NOTE: The following are set when version.go is generated
	version := irdmtools.Version
	releaseDate := irdmtools.ReleaseDate
	releaseHash := irdmtools.ReleaseHash
	fmtHelp := irdmtools.FmtHelp
	latestVersions := false

	showHelp, showVersion, showLicense := false, false, false
	configFName, debug, asXML := "", false, false
	idsFName, cName, pipeline := "", "", false
	flag.BoolVar(&showHelp, "help", false, "display help")
	flag.BoolVar(&showVersion, "version", false, "display version")
	flag.BoolVar(&showLicense, "license", false, "display license")
	flag.StringVar(&configFName, "config", configFName, "use a config file")
	flag.BoolVar(&asXML, "xml", asXML, "output as EPrint XML, does not work with -harvest")
	flag.BoolVar(&debug, "debug", debug, "display additional info to stderr")
	flag.StringVar(&idsFName, "ids", idsFName, "read ids from a file")
	flag.StringVar(&cName, "harvest", cName, "harvest JSON eprint records into the dataset collection.")
	flag.BoolVar(&pipeline, "pipeline", pipeline, "read from standard input, crosswalk and write to standard out")
	flag.BoolVar(&latestVersions, "latest", latestVersions, "only convert record if the latest version")

	flag.Parse()
	rdmids := flag.Args()


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
	if idsFName != "" {
		ids, err := getRdmIds(idsFName)
		if err != nil {
   			fmt.Fprintf(os.Stderr, "%s\n", err)
   			os.Exit(1)
		}
		rdmids = append(rdmids, ids...)
	}

	if len(rdmids) == 0 && ! pipeline {
		fmt.Fprintf(os.Stderr, "%s, requires ids unless running as a pipeline\n", appName)
		os.Exit(1)
	}

	app := new(irdmtools.Rdm2EPrint)
	if err := app.Configure(configFName, "", debug); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
	if cName != "" {
		if err := app.RunHarvest(os.Stdin, os.Stdout, os.Stderr, cName, rdmids, latestVersions); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	if pipeline {
		if err := app.RunPipeline(os.Stdin, os.Stdout, os.Stderr, asXML, latestVersions); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	if err := app.Run(os.Stdin, os.Stdout, os.Stderr, rdmids, asXML, latestVersions); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
