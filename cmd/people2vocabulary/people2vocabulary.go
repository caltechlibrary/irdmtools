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

	// Caltech Library
	"github.com/caltechlibrary/irdmtools"
	"github.com/caltechlibrary/simplified"

	// 3rd Party Libraries
	"gopkg.in/yaml.v3"
)

const (
	helpText = `%{app_name}(1) irdmtools user manual | version {version} {release_hash}
% R. S. Doiel
% {release_date}

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
: (default: true) Input is in csv format

-clrules
: (default: true) use Caltech Library rules

# EXAMPLES

~~~shell
    {app_name} < htdocs/people/people.json \
	     >people-vocabulary.yaml

	{app_name} -csv < htdocs/people/people.csv \
	     >people-vocabulary.yaml
~~~

`
)

func fmtYAML(o interface{}) string {
	src, _ := yaml.Marshal(o)
	return fmt.Sprintf("%s", src)
}

func mapField(person *simplified.Person, key string, val string) error {
	if val == "" {
		// NOTE: An empty value isn't an error, we just don't map it.
		return nil
	}
	switch key {
	case "family_name":
		person.Family = val
	case "given_name":
		person.Given = val
	case "clpid":
		identifier := new(simplified.Identifier)
		identifier.Scheme = "clpid"
		identifier.Identifier = val
		person.Identifiers = append(person.Identifiers, identifier)
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
		inputFName  string
		outputFName string

		showHelp    bool
		showVersion bool
		showLicense bool

		clRules bool
		inputIsCSV bool
	)
	appName := path.Base(os.Args[0])
	version := irdmtools.Version
	releaseDate := irdmtools.ReleaseDate
	releaseHash := irdmtools.ReleaseHash
	fmtHelp := irdmtools.FmtHelp

	in := os.Stdin
	out := os.Stdout
	eout := os.Stderr

	flag.BoolVar(&showHelp, "help", false, "display help text")
	flag.BoolVar(&showVersion, "version", false, "display version")
	flag.BoolVar(&showLicense, "license", false, "display license")
	flag.StringVar(&inputFName, "i", "", "input filename")
	flag.StringVar(&outputFName, "o", "", "output filename")
	flag.BoolVar(&inputIsCSV, "csv", true, "input is CSV format")
	flag.BoolVar(&clRules, "clrules", true, "use Caltech Library specific rules")
	flag.Parse()
	args := flag.Args()
	if showHelp {
		fmt.Fprintf(out, "%s\n", fmtHelp(helpText, appName, version, releaseDate, releaseHash))
		os.Exit(0)
	}
	if showLicense {
		fmt.Fprintf(out, "%s\n", irdmtools.LicenseText)
		os.Exit(0)
	}
	if showVersion {
		fmt.Fprintf(out, "%s %s %s\n", appName, version, releaseHash)
		os.Exit(0)
	}
	if (len(args) > 0) && (inputFName == "") {
		inputFName = args[0]
	}
	if (len(args) > 1) && (outputFName == "") {
		outputFName = args[1]
	}
	if (inputFName != "") && (inputFName != "-") {
		in, err = os.Open(inputFName)
		if err != nil {
			fmt.Fprintf(eout, "%s\n", err)
			os.Exit(1)
		}
		defer in.Close()
	}
	if (outputFName != "") && (outputFName != "-") {
		out, err = os.Create(outputFName)
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
		//NOTE: spreadsheet conversion process will filter out none
		// RDM identifiers when producing the YAML.
		r := csv.NewReader(bytes.NewBuffer(src))
		fields := []string{}
		rowNo := 0
		e := 0
		// NOTE:This is the Caltech affiliation, needed by clsRules
		// where this list of of Caltech people (from feeds).
		caltech := new(simplified.Affiliation)
		caltech.ID = "05dxps055"
		caltech.Name = "Caltech"
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
				if clRules {
					// Check if Caltech affiliation is asserted
					if ! person.HasAffiliation(caltech) {
						person.Affiliations = append(person.Affiliations, caltech)
					}
				}

				peopleList = append(peopleList, person)
			}
			rowNo++
		}
		if e > 0 {
			os.Exit(1)
		}
		src, err = yaml.Marshal(peopleList)
		if err != nil {
			fmt.Fprintf(eout, "%s\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(out, "%s\n", src)
		os.Exit(0)
	}

	// Import is JSON array of Person
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
		// filter for ORCID and clpid in the identifier list
		for j, identifier := range obj.Identifiers {
			switch obj.Identifiers[j].Scheme {
			case "clpid":
				
				obj.Identifiers = append(obj.Identifiers, identifier)
				orcidPeople = append(orcidPeople, obj)
			case "orcid":
				// Reset the identifier list so we can save it in our orcidPeople list
				obj.Identifiers = append(obj.Identifiers, identifier)
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
