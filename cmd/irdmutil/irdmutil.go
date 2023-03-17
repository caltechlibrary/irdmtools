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
% 2023-03-16

# NAME

{app_name}

# SYNOPSIS

{app_name} [OPTIONS] ACTION [ACTION_PARAMETERS ...]

# DESCRIPTION

__{app_name}__ provides a quick wrapper around access Invenio-RDM
JSON API. By default {app_name} looks in the current working directory
for a JSON configuration file that holds "base_url" to the Invenio-RDM
instance and any authentication information need to access the API.

__{app_name}__ normally expects a configuration file but if one
is not found then it can read its configuration from the environment.
The two environment variables used as INVENIO_API holding the string
that points to the Invenio JSON API and INVENIO_TOKEN needed to 
authenticate and use the API.  If one or both a missing an error will
be returned. If the environment INVENIO_API is set and you run the 
setup action they it will be used to populate the JSON configuration
example displayed by setup. If a configuration file exists it will
display the configuration with the invenio token overwritten with
a placeholder value.

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
: Display an example JSON setup configuration file, if it already exists then it will display the current configuration file. No optional or required parameters.

get_record_ids
: Returns a list of all repository record ids (can take a while). No optional or required parameters.

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

Get a list of Invenio-RDM record ids.

~~~
{app_name} get_record_ids
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
	configFName := "irdmtools.json"
	flag.BoolVar(&showHelp, "help", false, "display help")
	flag.BoolVar(&showVersion, "version", false, "display version")
	flag.BoolVar(&showLicense, "license", false, "display license")
	flag.StringVar(&configFName, "config", configFName, "use a config file")
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
		if err := util.LoadConfig(configFName); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
	}
	if err := util.Run(os.Stdin, os.Stdout, os.Stderr, action, params); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
