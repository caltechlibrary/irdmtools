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
% 2022-10-27

# NAME

{app_name}

# SYNOPSIS

{app_name} [OPTIONS] ACTION [ACTION_PARAMETERS ...]

# DESCRIPTION

__{app_name}__ provides a quick wrapper around access Invenio-RDM
JSON API. By default {app_name} looks in the current working directory
for a JSON configuration file that holds "base_url" to the Invenio-RDM
instance and any authentication information need to access the API.

# OPTIONS

help
: display help

license
: display license

version
: display version

config
: provide a path to an alternate configuration file

dataset
: store ivenio-rdm JSON record in a dataset collection (dataset v2)

# EXAMPLES

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

var (
	configFName = "irdmtools.json"
)

func fmtTxt(txt string, appName string) string {
	return strings.ReplaceAll(txt, "{app_name}", appName)
}

func main() {
	appName := path.Base(os.Args[0])
	showHelp, showVersion, showLicense := false, false, false
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
	if _, err := os.Stat(configFName); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "cound not find %s, %s\n", configFName, err)
		os.Exit(1)
	}
	util, err := irdmtools.MakeUtil(configFName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
	err = util.Run(os.Stdin, os.Stdout, os.Stderr, args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
