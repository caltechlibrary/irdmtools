package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
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

{app_name} [OPTIONS] < INPUT_JSON_FILE > OUTPUT_VOC_YAML_FILE

# DESCRIPTION

{app_name} converts a JSON array of people objects to a YAML
file suitable for import into Invenio-RDM.

# OPTIONS

-help
: display help

-license
: display license

-version
: display version

-i
: Read input from file

-o
: Write output to file

-csv
: Input is in csv format


# EXAMPLES

~~~shell
    {app_name} < htdocs/people/people.json \
	     >people-vocabulary.yaml

	{app_name} -csv < htdocs/people/people.csv \
	     >people-vocabulary.yaml
~~~

`
)

func fmtTxt(txt string, appName string, version string) string {
	return strings.ReplaceAll(strings.ReplaceAll(txt, "{app_name}", appName), "{version}", version)
}

func mapField(person *simplified.Person, key string, val string) error {
	if val == "" {
		// NOTE: An empty value isn't an error, we just don't map it.
		return nil
	}
	//fmt.Printf("DEBUG key: %q val: %T %q\n", key, val, val)
	switch key {
	case "family_name":
		person.Family = val
	case "given_name":
		person.Given = val
	case "clpid":
		if val != "" {
			identifier := new(simplified.Identifier)
			identifier.Scheme = "clpid"
			identifier.Identifier = val
			person.Identifiers = append(person.Identifiers, identifier)
		}
	case "cl_people_id":
		identifier := new(simplified.Identifier)
		identifier.Scheme = "clpid"
		identifier.Identifier = val
		person.Identifiers = append(person.Identifiers, identifier)
	case "thesis_id":
	case "advisor_id":
	case "authors_id":
	case "archivesspace_id":
	case "directory_id":
	case "viaf_id":
	case "lcnaf":
	case "isni":
		identifier := new(simplified.Identifier)
		identifier.Scheme = key
		identifier.Identifier = val
		person.Identifiers = append(person.Identifiers, identifier)
	case "wikidata":
	case "snac":
	case "orcid":
		identifier := new(simplified.Identifier)
		identifier.Scheme = key
		identifier.Identifier = val
		person.Identifiers = append(person.Identifiers, identifier)
	case "image":
	case "educated_at":
	case "caltech":
		affiliation := new(simplified.Affiliation)
		affiliation.ID = "05dxps055"
		affiliation.Name = "Caltech"
		person.Affiliations = append(person.Affiliations, affiliation)
	case "jpl":
		affiliation := new(simplified.Affiliation)
		affiliation.ID = "027k65916"
		affiliation.Name = "JPL"
		person.Affiliations = append(person.Affiliations, affiliation)
	case "faculty":
	case "alumn":
	case "status":
	case "directory_person_type":
	case "title":
	case "bio":
	case "division":
	case "authors_count":
	case "thesis_count":
	case "data_count":
	case "advisor_count":
	case "editor_count":
	case "updated":
	default:
		return fmt.Errorf("not know how to map %q <- %q", key, val)
	}
	if person.Name == "" && person.Given != "" && person.Family != "" {
		person.Name = fmt.Sprintf("%s, %s", person.Family, person.Given)
	}
	return nil
}

func main() {
	var (
		err    error
		input  string
		output string

		showHelp    bool
		showVersion bool
		showLicense bool

		inputIsCSV bool
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
	flag.BoolVar(&inputIsCSV, "csv", false, "input is CSV format")
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
	if inputIsCSV {
		//FIXME: read in spreadsheet and write out vocabulary file
		r := csv.NewReader(bytes.NewBuffer(src))
		fields := []string{}
		rowNo := 0
		e := 0
		for {
			cells, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Printf("row %d error, %s", rowNo, err)
				e += 1
				break
			}
			if rowNo == 0 {
				fields = cells[:]
			} else {
				person := new(simplified.Person)
				if len(cells) > len(fields) {
					log.Printf("row %d error, too many columns", rowNo)
					e += 1
					break
				}
				for colNo, val := range cells {
					key := fields[colNo]
					if err := mapField(person, key, val); err != nil {
						log.Printf("row %d error, %s", rowNo, err)
						e += 1
					}
				}
				peopleList = append(peopleList, person)
			}
			rowNo++
		}
		if e > 0 {
			os.Exit(1)
		}
	} else {
		if err := json.Unmarshal(src, &peopleList); err != nil {
			fmt.Fprintf(eout, "%s\n", err)
			os.Exit(1)
		}
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
