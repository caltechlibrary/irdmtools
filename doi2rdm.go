// / irdmtools is a package for working with institutional repositories and
// data management systems. Current implementation targets Invenio-RDM.
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
package irdmtools

import (
	"fmt"
	"io"
	"os"
	"strings"

	// 3rd Party packages
	"gopkg.in/yaml.v3"

	// Caltech Library packages
	"github.com/caltechlibrary/crossrefapi"
	"github.com/caltechlibrary/simplified"
)

const (
	EXIT_OK = 0
	ENOENT = 2
	ENOEXEC = 8
	EAGAIN = 11
)

// Doi2Rdm holds the configuration for doi2rdm cli.
type Doi2Rdm struct {
	Cfg *Config
}

type Doi2RdmOptions struct {
	MailTo              string            `json:"mailto,omitempty" yaml:"mailto,omitempty"`
	Download            bool              `json:"download,omitempty" yaml:"download,omitempty"`
	DotInitials         bool              `json:"dot_initials,omitempty" yaml:"dot_initials,omitempty"`
	ContributorTypes    map[string]string `json:"contributor_types,omitemptpy" yaml:"contributor_types,omitempty"`
	ResourceTypes       map[string]string `json:"resource_types,omitempty" yaml:"resource_types,omitempty"`
	DoiPrefixPublishers map[string]string `json:"doi_prefix_publishers,omitempty" yaml:"doi_prefix_publishers,omitempty"`
	ISSNJournals        map[string]string `json:"issn_journals,omitempty" yaml:"issn_journals,omitempty"`
	ISSNPublishers      map[string]string `json:"issn_publishers,omitempty" yaml:"issn_publishers,omitempty"`
	Debug               bool              `json:"debug,omitempty" yaml:"debug,omitempty"`
}

var (
	DefaultDoi2RdmOptionsYAML = []byte(`# This YAML file controls the mappings of 
# CrossRef records to RDM records values. It is based on the practice
# of Caltech Library in the development of CaltechAUTHORS and CaltechTHESIS
# over the last decades.
#
# Set the mail to used when connecting to CrossRef. This is usually the
# email address for our organization but could be for a person.
# It is the email address CrossRef will use if you're causing a problem
# and they need you to stop.
#
#mailto: jane.doe@example.edu
mailto: helpdesk@library.caltech.edu
# Add a period after initials is missing
dot_initials: true
# Map the CrossRef type to RDM type
contributor_types:
  author: author
  editor: editor
  reviewer: reviewer
  review-assistent: other
  stats-reviewer: other
  reader: other
  translator: translator
# Map the CrossRef and DataCite resource type to the RDM type
resource_types:
  article: publication-article
  Preprint: publication-preprint
  preprint: publication-preprint
  journal-article: publication-article
  book: publication-book
  book_section: publication-section
  book-chapter: publication-section
  conference_item: conference-paper
  proceedings-article: conference-paper
  dataset: dataset
  experiment: publication-deliverable
  journal_issue: publication-issue
  lab_notes: labnotebook
  monograph: publication-report
  oral_history: publication-oralhistory
  patent: publication-patent
  software: software
  teaching_resource: teachingresource
  thesis: publication-thesis
  video: video
  website: other
  other: other
  image: other
  report: publication-workingpaper
  report-component: publication-workingpaper
  posted-content: publication-preprint
  DataPaper: publication-datapaper
  Text: publication-other
# Mapping DOI prefixes to Publisher names (used to normalize publisher names)
doi_prefix_publishers:
# Mapping ISSN prefixes to Journals (used to normalize journal titles names)
issn_journals:
# Mapping ISSN prefixes to Publishers (used to normalize publisher names)
issn_publishers:
`)
)

// Configure reads the configuration file and environtment
// initialing the Cfg attribute of a Doi2Rdm object. It returns an error
// if problem were encounter.
//
// ```
//
//	app := new(irdmtools.Doi2Rdm)
//	if err := app.Configure("irdmtools.yaml", "TEST_"); err != nil {
//	   // ... handle error ...
//	}
//	fmt.Printf("Invenio RDM API UTL: %q\n", app.Cfg.IvenioAPI)
//	fmt.Printf("Invenio RDM token: %q\n", app.Cfg.InvenioToken)
//
// ```
func (app *Doi2Rdm) Configure(configFName string, envPrefix string, debug bool) error {
	if app == nil {
		app = new(Doi2Rdm)
	}
	cfg := NewConfig()
	// Load the config file if name isn't an empty string
	if configFName != "" {
		err := cfg.LoadConfig(configFName)
		if err != nil {
			return err
		}
	}
	// Merge settings from the environment
	if err := cfg.LoadEnv(envPrefix); err != nil {
		return err
	}
	app.Cfg = cfg
	if debug {
		app.Cfg.Debug = true
	}
	// Make sure we have a minimal useful configuration
	if app.Cfg.InvenioAPI == "" || app.Cfg.InvenioToken == "" {
		return fmt.Errorf("RDM_URL or RDM_TOK available")
	}
	return nil
}

// RunCrossRefToRdm implements the doi2rdm cli behaviors using the CrossRef service.
// With the exception of the "setup" action you should call `app.LoadConfig()` before execute
// Run.
//
// ```
//
//	app := new(irdmtools.Doi2Rdm)
//	// Load irdmtools settings
//	if err := app.LoadConfig("irdmtools.json"); err != nil {
//	   // ... handle error ...
//	}
//	// If options are provided then we need to set the filename
//	optionsFName := "doi2rdm.yaml"
//	doi := "10.3847/1538-3881/ad2765"
//	src, exitCode, err := app.Run(os.Stdin, os.Stdout, os.Stderr, optionFName, doi, "", false)
//	if err != nil {
//	    // ... handle error ...
//      os.Exit(exitCode)
//	}
//	fmt.Printf("%s\n", src)
//
// ```
func (app *Doi2Rdm) RunCrossRefToRdm(in io.Reader, out io.Writer, eout io.Writer, optionFName, doi string, diffFName string) (int, error) {
	var (
		err error
		src []byte
	)
	src = DefaultDoi2RdmOptionsYAML
	if optionFName != "" {
		src, err = os.ReadFile(optionFName)
		if err != nil {
			return ENOENT, err
		}
	}
	options := new(Doi2RdmOptions)
	if err := yaml.Unmarshal(src, &options); err != nil {
		return ENOEXEC, err
	}
	if app.Cfg.Debug {
		options.Debug = app.Cfg.Debug
	}
	if options.MailTo == "" {
		//mailTo = fmt.Sprintf("%s@%s", os.Getenv("USER"), os.Getenv("HOSTNAME"))
		options.MailTo = "helpdesk@library.caltech.edu"
	}
	var (
		oRecord *simplified.Record
		nRecord *simplified.Record
	)
	if diffFName != "" {
		oWork := new(crossrefapi.Works)
		src, err := os.ReadFile(diffFName)
		if err != nil {
			return ENOENT, err
		}
		if err := JSONUnmarshal(src, &oWork); err != nil {
			return ENOEXEC, err
		}
		oRecord, err = CrosswalkCrossRefWork(app.Cfg, oWork, options)
		if err != nil {
			return ENOEXEC, err
		}
	}
	nWork, err := QueryCrossRefWork(app.Cfg, doi, options)
	if err != nil {
		return ENOENT, err
	}
	nRecord, err = CrosswalkCrossRefWork(app.Cfg, nWork, options)
	if err != nil {
		return ENOEXEC, err
	}
	if diffFName != "" {
		src, err = oRecord.DiffAsJSON(nRecord)
	} else {
		src, err = JSONMarshalIndent(nRecord, "", "    ")
	}
	if err != nil {
		return ENOEXEC, err
	}
	fmt.Fprintf(out, "%s\n", src)
	return EXIT_OK, nil
}

// RunDataCiteToRdm implements the doi2rdm cli behaviors using the DataCite service.
// With the exception of the "setup" action you should call `app.LoadConfig()` before execute
// Run.
//
// ```
//
//		app := new(irdmtools.Doi2Rdm)
//	 // Load irdmtools settings
//		if err := app.LoadConfig("irdmtools.json"); err != nil {
//		   // ... handle error ...
//		}
//	 // If options are provided then we need to set the filename
//	 optionsFName := "doi2rdm.yaml"
//		doi := "10.48550/arXiv.2104.02480"
//		src, err := app.RunDataCiteToRdm(os.Stdin, os.Stdout, os.Stderr, optionFName, doi, "", false)
//		if err != nil {
//		    // ... handle error ...
//		}
//		fmt.Printf("%s\n", src)
//
// ```
func (app *Doi2Rdm) RunDataCiteToRdm(in io.Reader, out io.Writer, eout io.Writer, optionFName, doi string, diffFName string) (int, error) {
	var (
		err error
		src []byte
	)
	src = DefaultDoi2RdmOptionsYAML
	if optionFName != "" {
		src, err = os.ReadFile(optionFName)
		if err != nil {
			return ENOENT, err
		}
	}
	options := new(Doi2RdmOptions)
	if err := yaml.Unmarshal(src, &options); err != nil {
		return ENOEXEC, err
	}
	if app.Cfg.Debug {
		options.Debug = app.Cfg.Debug
	}
	if options.MailTo == "" {
		//mailTo = fmt.Sprintf("%s@%s", os.Getenv("USER"), os.Getenv("HOSTNAME"))
		options.MailTo = "helpdesk@library.caltech.edu"
	}
	var (
		oRecord *simplified.Record
		nRecord *simplified.Record
	)
	if diffFName != "" {
		object := map[string]interface{}{}
		src, err := os.ReadFile(diffFName)
		if err != nil {
			return ENOENT, err
		}
		if err := JSONUnmarshal(src, &object); err != nil {
			return ENOEXEC, err
		}
		oRecord, err = CrosswalkDataCiteObject(app.Cfg, object, options)
		if err != nil {
			return ENOEXEC, err
		}
	}
	nWork, err := QueryDataCiteObject(app.Cfg, doi, options)
	if err != nil {
		return ENOENT, err
	}
	if len(nWork) == 0 {
		return ENOENT, fmt.Errorf("not data received for %q", doi)
	}
	nRecord, err = CrosswalkDataCiteObject(app.Cfg, nWork, options)
	if err != nil {
		return ENOEXEC, err
	}
	if diffFName != "" {
		src, err = oRecord.DiffAsJSON(nRecord)
	} else {
		src, err = JSONMarshalIndent(nRecord, "", "    ")
	}
	if err != nil {
		return ENOEXEC, err
	}
	fmt.Fprintf(out, "%s\n", src)
	return EXIT_OK, nil
}

// RunDoiToRDMCombined implements the doi2rdm cli behaviors using the CrossRead and DataCite service.
// With the exception of the "setup" action you should call `app.LoadConfig()` before execute
// Run.
//
// ```
//
//		app := new(irdmtools.Doi2Rdm)
//	 // Load irdmtools settings
//		if err := app.LoadConfig("irdmtools.json"); err != nil {
//		   // ... handle error ...
//		}
//	 // If options are provided then we need to set the filename
//	 optionsFName := "doi2rdm.yaml"
//		doi := "10.48550/arXiv.2104.02480"
//		src, err := app.RunDoiToRdmCombined(os.Stdin, os.Stdout, os.Stderr, optionFName, doi, "", false)
//		if err != nil {
//		    // ... handle error ...
//		}
//		fmt.Printf("%s\n", src)
//
// ```
func (app *Doi2Rdm) RunDoiToRdmCombined(in io.Reader, out io.Writer, eout io.Writer, optionFName, doi string, diffFName string) (int, error) {
	// Do we have an arXiv id?
	if strings.HasPrefix(strings.ToLower(doi), "arxiv:") {
		return app.RunDataCiteToRdm(in, out, eout, optionFName, doi, diffFName)
	}
	if _, crErr := app.RunCrossRefToRdm(in, out, eout, optionFName, doi, diffFName); crErr != nil  {
		// Then try DataCiteToRdm
		if exitCode, dcErr := app.RunDataCiteToRdm(in, out, eout, optionFName, doi, diffFName); dcErr != nil {
			return exitCode, fmt.Errorf("crossref: %s, datacite: %s", crErr, dcErr)
		}
	}
	return EXIT_OK, nil
}
