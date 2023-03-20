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
	helpText = `% {app_name}(1) user manual
% R. S. Doiel
% 2023-03-20

# NAME

{app_name}

# SYNOPSIS

{app_name} [OPTIONS] ACTION [ACTION_PARAMETERS ...]

# DESCRIPTION

__{app_name}__ provides a quick wrapper around Invenio-RDM's OAI-PMH
and REST API. By default {app_name} looks for two environment variables.

- RDM_INVENIO_API
- RDM_INVENIO_TOKEN

These are use to acces the Invenio RDM REST API and OAI-PMH services.

You may specify a JSON configuration file holding the attributes of 
"invenio_api" and "invenio_token" instead of using environment variables.

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
: provide a path to an alternate configuration file (default is irdmtools.json)

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

# ACTION_PARAMETERS

Action parameters are the specific optional or required parameters need to complete an aciton.


# EXAMPLES

Setup for __{app_name}__ by writing an example JSON configuration file.
"nano" is an example text editor program, you need to edit the sample
configuration appropriately.

~~~
{app_name} setup >irdmtools.json
nano irdmtools.json
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
	return strings.ReplaceAll(txt, "{app_name}", appName)
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
		fmt.Fprintf(os.Stdout, "%s %s\n", appName, irdmtools.Version)
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
	// Create a utility object
	util := new(irdmtools.Util)
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
		if err := util.Configure(configFName, "RDM_", debug); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
	}
	if err := util.Run(os.Stdin, os.Stdout, os.Stderr, action, params); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
