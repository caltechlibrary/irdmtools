package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	// Caltech Library
	"github.com/caltechlibrary/irdmtools"
	"github.com/caltechlibrary/simplified"

	// 3rd Party Libraries
	"gopkg.in/yaml.v3"
)

const (
	helpText = `%{app_name}(1) {app_name} user manual
% R. S. Doiel
% May 17, 2023

# NAME

{app_name}

# SYSNOPSIS

{app_name} < INPUT_JSON_FILE > OUTPUT_VOC_YAML_FILE

# DESCRIPTION

{app_name} converts a JSON array of people objects to a YAML
file suitable for import into Invenio-RDM.

# EXAMPLES

~~~shell
    {app_name} < htdocs/people/people.json \
	     >htdocs/people/people-vocabulary.yaml
~~~

`
)

func fmtTxt(txt string, appName string, version string) string {
	return strings.ReplaceAll(strings.ReplaceAll(txt, "{app_name}", appName), "{version}", version)
}

func main() {
	var (
		err    error
		input  string
		output string

		showHelp bool
		showVersion bool
		showLicense bool
	)
	appName := path.Base(os.Args[0])
	version := irdmtools.Version
	in := os.Stdin
	out := os.Stdout
	eout := os.Stderr

	flag.BoolVar(&showHelp, "help", false, "display help text")
	flag.BoolVar(&showVersion, "version", false, "display version")
	flag.BoolVar(&showLicense, "license", false, "display license")
	flag.StringVar(&input, "i", "", "input filename")
	flag.StringVar(&output, "o", "", "output filename")
	flag.Parse()
	args := flag.Args()
	if showHelp {
		fmt.Fprintf(out, "%s\n", fmtTxt(helpText, appName, version))
		os.Exit(0)
	}
	if showLicense {
		fmt.Fprintf(out, "%s\n", irdmtools.LicenseText)
		os.Exit(0)
	}
	if showVersion {
		fmt.Fprintf(out, "%s %s\n", appName, version)
		os.Exit(0)
	}
	if (len(args) > 0) && (input == "") {
		input = args[0]
	}
	if (len(args) > 1) && (output == "") {
		output = args[1]
	}
	if (input != "") && (input != "-") {
		in, err = os.Open(input)
		if err != nil {
			fmt.Fprintf(eout, "%s\n", err)
			os.Exit(1)
		}
		defer in.Close()
	}
	if (output != "") && (output != "-") {
		out, err = os.Create(output)
		if err != nil {
			fmt.Fprintf(eout, "%s\n", err)
			os.Exit(1)
		}
		defer out.Close()
	}
	src, err := ioutil.ReadAll(in)
	if err != nil {
		fmt.Fprintf(eout, "%s\n", err)
		os.Exit(1)
	}
	peopleList := []*simplified.Person{}
	if err := json.Unmarshal(src, &peopleList); err != nil {
		fmt.Fprintf(eout, "%s\n", err)
		os.Exit(1)
	}

	// NOTE: Invenio-RDM can only import people if they have an ORCID into the people
	// controlled vocabulary file used for auto-complete.

	// Process list
	orcidPeople := []*simplified.Person{}
	for _, obj := range peopleList {
		// Prune unwanted fields (e.g. .sort_name)
		obj.Sort = ""
		// filter for ORCID in the identifier list
		for j, identifier := range obj.Identifiers {
			switch obj.Identifiers[j].Scheme {
			case "orcid":
				// Reset the identifier list so we can save it in our orcidPeople list
				obj.Identifiers = append([]*simplified.Identifier{}, identifier)
				orcidPeople = append(orcidPeople, obj)
				break
			}
		}
	}

	src, err = yaml.Marshal(orcidPeople)
	if err != nil {
		fmt.Fprintf(eout, "%s\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(out, "%s\n", src)
}
