// ep3util is a command line program for working with a read only EPrint API.
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
	helpText = `%{app_name}(1) irdmtools user manual | version {version} {release_hash}
% R. S. Doiel and Tom Morrell
% {release_date}

# NAME

{app_name}

# SYNOPSIS

{app_name} [OPTIONS] ACTION [ACTION_PARAMETERS ...]

# DESCRIPTION

__{app_name}__ provides a quick wrapper around EPrints 3.3 REST API.
By default {app_name} looks for five environment variables.

REPO_ID
: the EPrints repository id (name of database and archive subdirectory).

EPRINT_HOST
: the hostname for EPrint's.

EPRINT_USER
: the username having permissions to access the EPrint REST API.

EPRINT_PASSWORD
: the password for the username with access to the EPrint REST API.

C_NAME
: If harvesting the dataset collection name to harvest the records to.


The environment provides the default values for configuration. They
maybe overwritten by using a JSON configuration file. The corresponding
attributes are "repo_id", "eprint_host" and "c_name".


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

get_all_ids
: Returns a list of all repository record ids. The method uses OAI-PMH for id retrieval. It is rate limited and will take come time to return all record ids. A test instance took 11 minutes to retrieve 24000 record ids.

get_record RECORD_ID
: Returns a specific simplified record indicated by RECORD_ID, e.g. 23808. The REORCID_ID is a required parameter.

harvest [HARVEST_OPTIONS] [KEY_LIST_JSON]
: harvest takes a JSON file containing a list of keys and harvests each record into a dataset collection. If combined
with one of the options, e.g. `+"`"+`-all`+"`"+`, you can skip provideing the KEY_LIST_JSON file.

# HARVEST_OPTIONS

-all
: Harvest all records

-modified START [END]
: Harvest records modified between start and end dates.

# ACTION_PARAMETERS

Action parameters are the specific optional or required parameters need to complete an aciton.


# EXAMPLES

Setup for __{app_name}__ by writing an example JSON configuration file.
"nano" is an example text editor program, you need to edit the sample
configuration appropriately.

~~~
{app_name} setup >eprinttools.json
nano eprinttools.json
~~~

Get a list of all EPrint record ids.

~~~
{app_name} get_all_ids
~~~

Get a specific EPrint record. Record is validated
against irdmtool EPrints data model.

~~~
{app_name} get_record 23808
~~~

Harvest all records

~~~
{app_name} harvest -all
~~~

Harvest records created or modified in the month of September, 2023.

~~~
{app_name} harvest -modified 2023-09-01 2023-09-30
~~~
`
)

func fmtTxt(txt string, appName string) string {
	return strings.ReplaceAll(strings.ReplaceAll(txt, "{app_name}", appName), "{version}", irdmtools.Version)
}

func main() {
	appName := path.Base(os.Args[0])
	// NOTE: The following will be set when version.go is generated
	version := irdmtools.Version
	releaseDate := irdmtools.ReleaseDate
	releaseHash := irdmtools.ReleaseHash
	fmtHelp := irdmtools.FmtHelp

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
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "%s %s\n", appName, irdmtools.Version)
		os.Exit(1)
	}
	// Create a Ep3Util object
	app := new(irdmtools.Ep3Util)
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "missing action, don't know what to do\n")
		os.Exit(1)
	}
	app.Debug = debug
	// To start we assume the first parameter is an action
	action, params := args[0], args[1:]
	// double check to see if -setup was used, and adjust
	if action == "setup" {
		if len(params) == 0 {
			params = append(params, configFName)
		}
	} else {
		if err := app.Configure(configFName, "", debug); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
	}
	if err := app.Run(os.Stdin, os.Stdout, os.Stderr, action, params); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
