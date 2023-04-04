// rdmutil is a command line program for working with Invenio RDM.
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
% 2023-04-04

# NAME

{app_name}

# SYNOPSIS

{app_name} [OPTIONS] ACTION [ACTION_PARAMETERS ...]

# DESCRIPTION

__{app_name}__ provides a quick wrapper around Invenio-RDM's OAI-PMH
and REST API. By default {app_name} looks for three environment variables.

RDM_INVENIO_API
: the URL of the Invenio RDM API and OAI-PMH services

RDM_INVENIO_TOKEN
: the token needed to access the Invenio RDM API and OAI-PMH services


RDM_C_NAME
: A dataset collection name. Collection must exist. See `+"`"+`dataset help init`+"`"+`

The environment provides the default values for configuration. They
maybe overwritten by using a JSON configuration file. The corresponding
attributes are "invenio_api", "invenio_token" and "c_name".

{app_name} uses the OAI-PMH service to retrieve record ids. This maybe
slow due to rate limits. Also provided is a query service and record
retrieval using Invenio RDM's REST API. These are faster but the query
services limited the total number of results to 10K records.

# OPTIONS

help
: display help

license
: display license

version
: display version

config
: provide a path to an alternate configuration file (e.g. "rdmtools.json")

# ACTION

__{app_name}__ supports the following actions.

setup
: Display an example JSON setup configuration file, if it already exists then it will display the current configuration file. No optional or required parameters. When displaying the JSON configuration a placeholder will be used for the token value.

get_modified_ids START [END]
: Returns a list of modified record ids (created, updated, deleted) in the time range listed.  This method uses OAI-PMH for id retrieval. It is rate limited. Start and end dates are inclusive and should be specific in YYYY-MM-DD format.

get_all_ids
: Returns a list of all repository record ids. The method uses OAI-PMH for id retrieval. It is rate limited and will take come time to return all record ids. A test instance took 11 minutes to retrieve 24000 record ids.

query QUERY_STRING [size | size sort]
: Returns a result using RDM's search engine. It is limited to about 10K total results. You can use the see RDM's documentation for query construction.  See <https://inveniordm.docs.cern.ch/customize/search/>, <https://inveniordm.docs.cern.ch/reference/rest_api_requests/> and https://www.elastic.co/guide/en/elasticsearch/reference/current/query-dsl-query-string-query.html#query-string-syntax. Query takes one required parameter and two option.

get_record RECORD_ID
: Returns a specific record indicated by RECORD_ID, e.g. bq3se-47g50. The REORCID_ID is a required parameter.

harvest KEY_JSON
: harvest takes a JSON file containing a list of keys and harvests each record into the dataset collection.

# ACTION_PARAMETERS

Action parameters are the specific optional or required parameters need to complete an aciton.


# EXAMPLES

Setup for __{app_name}__ by writing an example JSON configuration file.
"nano" is an example text editor program, you need to edit the sample
configuration appropriately.

~~~
{app_name} setup >rdmtools.json
nano rdmtools.json
~~~

Get a list of Invenio-RDM record ids modified from
Jan 1, 2023 to Jan 31, 2023.

~~~
{app_name} get_modified_ids 2023-01-01 2023-01-31
~~~

Get a list of all Invenio-RDM record ids.

~~~
{app_name} get_all_ids
~~~

Get a specific Invenio-RDM record.

~~~
{app_name} get_record bq3se-47g50
~~~


`
)


func fmtTxt(txt string, appName string) string {
	return strings.ReplaceAll(strings.ReplaceAll(txt, "{app_name}", appName), "{version}", irdmtools.Version)
}

func main() {
	appName := path.Base(os.Args[0])
	showHelp, showVersion, showLicense := false, false, false
	configFName, debug := "", false
	flag.BoolVar(&showHelp, "help", false, "display help")
	flag.BoolVar(&showVersion, "version", false, "display version")
	flag.BoolVar(&showLicense, "license", false, "display license")
	flag.StringVar(&configFName, "config", configFName, "use a config file")
	flag.BoolVar(&debug, "debug", debug, "display additional info to stderr")
	flag.Parse()
	args := flag.Args()

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
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "%s\n", fmtTxt(helpText, appName))
		fmt.Fprintf(os.Stderr, "%s %s\n", appName, irdmtools.Version)
		os.Exit(1)
	}
	// Create a RdmUtil object
	app := new(irdmtools.RdmUtil)
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "missing action, don't know what to do\n")
		os.Exit(1)
	}
	// To start we assume the first parameter is an action
	action, params := args[0], args[1:]
	// double check to see if -setup was used, and adjust
	if action == "setup" {
		if len(params) == 0 {
			params = append(params, configFName)
		}
	} else {
		if err := app.Configure(configFName, "RDM_", debug); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
	}
	if err := app.Run(os.Stdin, os.Stdout, os.Stderr, action, params); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
